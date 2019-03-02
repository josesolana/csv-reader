package integrator

import (
	"database/sql"
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
	poolWorker     []*worker
	runningWorkers *sync.WaitGroup
	db             database.Db
	offSet         int64
	runCh          *chan interface{}
}

// NewIntegrator Factory pattern
func NewIntegrator() *Integrator {
	i := &Integrator{
		runningWorkers: new(sync.WaitGroup),
		db:             database.NewDb(),
		offSet:         0,
	}
	ch := make(chan interface{}, 1)
	i.runCh = &ch
	i.createPoolWorker()
	return i
}

// Migrate Reads from DB and send info to JSON CRM API
func (i *Integrator) Migrate(close *chan os.Signal) {
	var sleep time.Duration
	var err error
	var rows *sql.Rows
	var newOffSet int64

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
		case s := <-*close:
			log.Printf("Got signal: %s\n", s.String())
			i.finish()
			return
		case <-time.After(sleep):
			if rows, err = i.db.Read(constants.BatchSizeRow, i.offSet); err != nil {
				if err == orm.ErrNoRows {
					log.Printf("No more Data. Waiting...\n")
				} else {
					log.Printf("Failed when connect to DB: %v\n", err)
				}
				sleep = backOff.Duration()
				log.Printf("Sleeping by BackOff %v\n", sleep)
			} else {
				log.Printf("Processing with offset %d\n", i.offSet)
				newOffSet, err = i.balanceLoad(rows)
				if err != nil {
					sleep = backOff.Duration()
					log.Printf("Sleeping by BackOff %v\n", sleep)
					break
				}
				if newOffSet < 0 {
					sleep = backOff.Duration()
					log.Println("Waiting for more Rows...")
					log.Printf("Sleeping by BackOff %v\n", sleep)
					break
				}
				i.offSet = newOffSet
				sleep = 0
			}
		}
	}
}

func (i *Integrator) balanceLoad(rows *sql.Rows) (int64, error) {
	var w int64
	customer := models.Customer{ID: -1}

	for rows.Next() {
		if err := rows.Scan(&customer.ID, &customer.FirstName, &customer.LastName, &customer.Email, &customer.Phone); err != nil {
			log.Printf("Cannot retrieve info from DB. Error: %s\n", err.Error())
			return 0, err
		}
		w = customer.ID % constants.Workers
		i.poolWorker[w].source <- customer
	}
	return customer.ID, nil
}

// createPoolWorker Create a workers's slice.
// There are go routines as workers set in constants.Workers.
// Workers are a channel to a function which do the job.
func (i *Integrator) createPoolWorker() {
	log.Printf("Starting %v Workers\n", constants.Workers)
	i.runningWorkers.Add(constants.Workers)
	workers := make([]*worker, constants.Workers)
	for index, _ := range workers {
		w := worker{
			source: make(chan models.Customer, constants.Buff),
			quit:   i.runCh,
		}
		w.Start(i.runningWorkers)
		workers[index] = &w
	}

	i.poolWorker = workers
}

func (i *Integrator) finish() {
	close(*i.runCh)
	i.db.Close()
	i.runningWorkers.Wait()
	log.Printf("Every worker's channels has been closed\n")
}
