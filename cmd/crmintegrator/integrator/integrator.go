package integrator

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/jpillora/backoff"

	"github.com/josesolana/csv-reader/cmd/database"
	"github.com/josesolana/csv-reader/cmd/models"
	"github.com/josesolana/csv-reader/constants"
)

// Integrator Reads from DB and send info to JSON CRM API
type Integrator struct {
	poolWorker          []chan models.Customer
	runningWorkers, job *sync.WaitGroup
	db                  database.Db
	offSet              int64
}

// NewIntegrator Factory pattern
func NewIntegrator() *Integrator {
	i := &Integrator{
		job:            new(sync.WaitGroup),
		runningWorkers: new(sync.WaitGroup),
		db:             database.NewDb(),
		offSet:         0,
	}
	i.createPoolWorker()
	return i
}

// Migrate Reads from DB and send info to JSON CRM API
func (i *Integrator) Migrate(runCh *chan os.Signal) {
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
	rand.Seed(time.Now().UTC().UnixNano())

	for {
		select {
		case s := <-*runCh:
			log.Printf("Got signal: %s", s.String())
			i.finish()
			return
		case <-time.After(sleep):
			if rows, err := i.db.Read(constants.BatchSizeRow, i.offSet); err != nil {
				if err == orm.ErrNoRows {
					log.Printf("No more Data. Waiting...")
				} else {
					log.Printf("Failed when connect to DB: %v", err)
				}
				sleep = backOff.Duration()
				log.Printf("Sleeping by BackOff %v", sleep)
			} else {
				log.Printf("Processing %d rows", i.offSet)
				i.balanceLoad(rows)
				sleep = 0
			}
		default:

		}
	}
}

func (i *Integrator) balanceLoad(rows *sql.Rows) error {
	customer := models.Customer{}
	for rows.Next() {
		if err := rows.Scan(&customer.ID, &customer.FirstName, &customer.LastName, &customer.Email, &customer.Phone); err != nil {
			return err
		}
		w := customer.ID % constants.Workers
		i.poolWorker[w] <- customer
	}
	return nil
}

func (i *Integrator) processRow(ch *chan models.Customer, job, runningWorkers *sync.WaitGroup) {
	//TODO: TO CRM It going to implement Exponential BackOff if Fails.
	for row := range *ch {
		job.Done()
		fmt.Printf("ROW: %v", row)
	}
	runningWorkers.Done()
}

// createPoolWorker Create a channel's slice.
// There are go routines as workers set constants.Workers.
// Workers are a channel to a function which do the job.
func (i *Integrator) createPoolWorker() {
	log.Printf("Starting %v Workers", constants.Workers)
	i.runningWorkers.Add(constants.Workers)

	workers := make([]chan models.Customer, constants.Workers)
	for index, _ := range workers {
		w := make(chan models.Customer, constants.Buff)
		workers[index] = w
		go i.processRow(&w, i.job, i.runningWorkers)
	}

	i.poolWorker = workers
}

func (i *Integrator) finish() {
	i.db.Close()

	i.job.Wait()
	for _, ch := range i.poolWorker {
		close(ch)
	}
	i.runningWorkers.Wait()
	log.Printf("Every worker's channels has been closed")
}
