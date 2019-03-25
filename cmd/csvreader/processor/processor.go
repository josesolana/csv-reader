package processor

import (
	"io"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/josesolana/csv-reader/cmd/csvreader/database"
	fh "github.com/josesolana/csv-reader/cmd/csvreader/filehandler"
	"github.com/josesolana/csv-reader/constants"
)

// Processor Read and save file into DB
type Processor struct {
	poolWorker          []chan []string
	runningWorkers, job *sync.WaitGroup
	runCh               chan error
	db                  database.DB
	reader              fh.Readable
}

// NewProcessor Factory pattern
func NewProcessor(name string) (*Processor, error) {
	reader, err := fh.NewFileHandler(name)
	if err != nil {
		return nil, err
	}

	row, err := reader.Read()
	if err == io.EOF {
		log.Println("File is empty")
		return nil, err
	} else if err != nil {
		log.Println("Cannot read a file line")
		return nil, err
	}
	log.Printf("Columns: %s\n", row)
	return NewProcessorWithValues(reader, database.NewDB(name, row)), nil
}

// NewProcessorWithValues Factory pattern
func NewProcessorWithValues(reader fh.Readable, db database.DB) *Processor {
	p := &Processor{
		db:             db,
		reader:         reader,
		job:            new(sync.WaitGroup),
		runningWorkers: new(sync.WaitGroup),
		runCh:          make(chan error, constants.Workers),
	}
	p.createPoolWorker()
	return p
}

// Migrate a file
//
// - Read from file
//
// - Save it into DB
func (p *Processor) Migrate() error {
	defer p.finish()
	var line []string
	var err error
	rand.Seed(time.Now().UTC().UnixNano())
	for {
		select {
		case err := <-p.runCh:
			return err
		default:
			line, err = p.reader.Read()
			switch err {
			case io.EOF:
				log.Println("File has been complete")
				return nil
			case nil:
				p.balanceLoad(line)
			default:
				log.Println("Skipped Line. Error: ", err)
			}
		}
	}
}

// createPoolWorker Create a channel's slice.
// There are go routines as workers set in constants.Workers.
// Workers are a channel to a function which do the job.
func (p *Processor) createPoolWorker() {
	log.Printf("Starting %v Workers", constants.Workers)
	p.runningWorkers.Add(constants.Workers)

	workers := make([]chan []string, constants.Workers)
	for i, _ := range workers {
		w := make(chan []string, constants.Buff)
		workers[i] = w
		go p.processRow(w, p.job, p.runningWorkers, p.runCh)
	}

	p.poolWorker = workers
}

func (p *Processor) processRow(ch chan []string, job, runningWorkers *sync.WaitGroup, runCh chan error) {
	ok := true
	for row := range ch {
		if ok {
			err := p.db.Insert(row...)
			if err != nil {
				ok = false
				runCh <- err
			}
		}
		job.Done()
	}
	runningWorkers.Done()
}

func (p *Processor) finish() {
	p.job.Wait()
	for _, ch := range p.poolWorker {
		close(ch)
	}
	p.runningWorkers.Wait()
	log.Printf("Every worker's channels has been closed")

	if err := p.db.Close(); err != nil {
		log.Println("Cannot close DB. Error: ", err)
	}
	if err := p.reader.Close(); err != nil {
		log.Println("Cannot close Reader. Error: ", err)
	}
}

func (p *Processor) balanceLoad(line []string) {
	// Random string value in Bytes used for a "random" balance
	randStr := line[rand.Intn(len(line))]
	randChar := randStr[rand.Intn(len(randStr))]
	w := randChar % constants.Workers
	p.job.Add(1)
	p.poolWorker[w] <- line
}
