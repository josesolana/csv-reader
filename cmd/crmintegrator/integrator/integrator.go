package integrator

import (
	"database/sql"
	"errors"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/jpillora/backoff"

	"github.com/josesolana/csv-reader/cmd/crmintegrator/database"
	c "github.com/josesolana/csv-reader/constants"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

var errGotSign = errors.New(c.ErrGotSignal)
var errFailWork = errors.New(c.ErrFailureWorker)

// Integrator Reads from DB and send info to JSON CRM API
type Integrator struct {
	poolWorker     []*worker
	runningWorkers *sync.WaitGroup
	jobs           *sync.WaitGroup
	db             database.DB
	quitCh         chan interface{}
	workerFailCh   chan error
	close          chan os.Signal
}

// NewIntegrator Factory pattern
func NewIntegrator(name string, close chan os.Signal) *Integrator {
	i := &Integrator{
		runningWorkers: new(sync.WaitGroup),
		jobs:           new(sync.WaitGroup),
		db:             database.NewDB(name),
		quitCh:         make(chan interface{}),
		workerFailCh:   make(chan error),
		close:          close,
	}

	i.createPoolWorker(name)
	return i
}

// Migrate Reads from DB and send info to JSON CRM API
func (i *Integrator) Migrate() {
	var sleep time.Duration
	backOff := &backoff.Backoff{
		// Min value to retry, if DB is down(It starts at Min)
		Min: 10 * time.Second,
		// After every call to Duration() it is multiplied by Factor
		Factor: 2,
		// It is capped at Max
		Max: 5 * time.Minute,
		// Adds some randomization to the backoff durations.
		Jitter: true,
	}

	for {
		select {
		case err := <-i.workerFailCh:
			log.Println("A worker got a failure: ", err)
			i.finish()
			return
		case s := <-i.close:
			log.Printf("Got signal: %s\n", s.String())
			i.finish()
			return
		case <-time.After(sleep):
			if err := i.processRows(&sleep, backOff); err != nil {
				if err == errGotSign {
					return
				}
				log.Fatalln(err)
			}
		}
	}
}

// processRows Process a row batch.
//
// - Start a transaction for Select for Update.
//
// - Read from DB if has rows, otherwise wait for new rows.
//
// - Randomly balance load to between Workers.
//
// - Commit when every jobs has been Done.
//	 Only those rows whose has been correctly read will be updated, because
//	 a the waitgroup increase only if a row was correctly send to a worker.
//	 If fails in a worker's job, Retry Column became Retry+1
//
// - Return error if, and only if, a commit cannot be executed.
func (i *Integrator) processRows(sleep *time.Duration, bo *backoff.Backoff) error {
	if err := i.db.Begin(); err != nil {
		log.Println("Cannot Begin a transaction")
		return nil
	}

	rows, err := i.db.Read()
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("No more Data. Waiting...")
		} else {
			log.Println("Cannot read from DB: ", err)
		}
		*sleep = bo.Duration()
		log.Println("Sleeping by BackOff ", sleep)
		return nil
	}

	errBL := i.balanceLoad(rows)
	if errBL == errFailWork || errBL == errGotSign {
		//Try to commit finalized work.
		if err := i.finishCommit(); err != nil {
			log.Println("Cannot Commit a transaction")
			return err
		}
		return errBL
	}

	log.Println("Waiting to jobs being done by workers")
	i.jobs.Wait()

	if err := i.db.Commit(); err != nil {
		log.Println("Cannot Commit a transaction")
		return err
	}
	if errBL != nil {
		log.Println("Cannot Read correctly from DB. Error: ", err)
		*sleep = bo.Duration()
		log.Println("Sleeping by BackOff ", sleep)
		return nil
	}
	*sleep = 0
	return nil
}

func (i *Integrator) balanceLoad(rows *sql.Rows) error {
	for {
		select {
		case err := <-i.workerFailCh:
			log.Println("A worker got a failure: ", err)
			return errFailWork
		case s := <-i.close:
			log.Printf("Got signal: %s\n", s.String())
			return errGotSign
		default:
			if rows.Next() {
				vals, err := i.createScanSlice(rows)
				if err != nil {
					return err
				}
				if err := rows.Scan(vals...); err != nil {
					log.Printf("Cannot retrieve info from DB. Error: %s\n", err)
					return err
				}
				w := *(vals[0]).(*int) % c.Workers
				i.poolWorker[w].sourceCh <- vals
				i.jobs.Add(1)
			} else {
				return nil
			}
		}
	}
}

func (i *Integrator) createScanSlice(rows *sql.Rows) ([]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	vals := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = new(sql.RawBytes)
	}
	vals[c.IDPos] = new(int)
	vals[c.IsProcessedPos] = new(bool)
	vals[c.RetryPos] = new(int)
	return vals, nil
}

// createPoolWorker Create a workers's slice.
// There are go routines as workers set in constants.Workers.
// Workers are a channel to a function which do the job.
func (i *Integrator) createPoolWorker(name string) {
	log.Printf("Starting %d Workers\n", c.Workers)
	i.runningWorkers.Add(c.Workers)
	workers := make([]*worker, c.Workers)
	for index, _ := range workers {
		w := worker{
			sourceCh: make(chan []interface{}, c.Buff),
			quitCh:   i.quitCh,
			db:       &i.db,
			errorCh:  i.workerFailCh,
		}
		w.Start(i.runningWorkers, i.jobs)
		workers[index] = &w
	}

	i.poolWorker = workers
}

func (i *Integrator) finish() {
	log.Println("Send close Broadcast")
	close(i.quitCh)
	log.Println("Waiting for finish workers")
	i.runningWorkers.Wait()
	log.Println("Closing DB")
	if err := i.db.Close(); len(err) != 0 {
		log.Fatalln(err)
	}
	log.Println("Everythings has been closed")
}

func (i *Integrator) finishCommit() error {
	log.Println("Send close Broadcast")
	close(i.quitCh)
	log.Println("Waiting for finish workers")
	i.runningWorkers.Wait()

	log.Println("Forced Commit")
	if err := i.db.Commit(); err != nil {
		log.Println("Cannot Commit a transaction")
		return err
	}

	log.Println("Closing DB")
	if err := i.db.Close(); len(err) != 0 {
		log.Fatalln(err)
	}
	log.Println("Everythings has been closed")
	return nil
}
