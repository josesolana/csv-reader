package processor

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/josesolana/csv-reader/cmd/database"
	"github.com/josesolana/csv-reader/constants"
)

// Processor Read and save the csv file into DB
type Processor struct {
	poolWorker          []chan []string
	runningWorkers, job *sync.WaitGroup
	runCh               *chan error
	db                  database.Db
	csvFile             *os.File
}

// NewProcessor Factory pattern
func NewProcessor() *Processor {
	p := &Processor{
		job:            new(sync.WaitGroup),
		runningWorkers: new(sync.WaitGroup),
		db:             database.NewDb(),
	}
	runCh := make(chan error, constants.Workers)
	p.runCh = &runCh
	p.createPoolWorker()
	return p
}

// Migrate a csv file
//
// - Read a csv file
//
// - Save it into DB
//
// - Send info to CrmIntegrator
func (p *Processor) Migrate(name string) error {
	defer p.finish()
	reader, err := p.openCsv(name)
	if err != nil {
		return err
	}
	err = p.saveIntoDb(reader)
	return err
}

func (p *Processor) saveIntoDb(reader *csv.Reader) error {
	var line []string
	var err error
	for {
		select {
		case err := <-*p.runCh:
			return err
		default:
			line, err = reader.Read()
			switch err {
			case io.EOF:
				log.Println("CSV file has been complete")
				return nil
			case nil:
				p.balanceLoad(line)
			default:
				log.Printf("Skipped Line. Error: %s", err.Error())
			}
		}
	}
}

func (p *Processor) openCsv(name string) (*csv.Reader, error) {
	filePath, err := p.getFilePath(name)
	if err != nil {
		return nil, err
	}

	p.csvFile, err = os.Open(filePath)
	if err != nil {
		log.Printf("Cannot open file: %s", filePath)
		return nil, err
	}

	reader := csv.NewReader(p.csvFile)
	line, err := reader.Read()
	if err == io.EOF {
		log.Println("CSV file is empty")
		return nil, err
	} else if err != nil {
		log.Printf("Cannot read a CSV line")
		return nil, err
	}
	log.Printf("Database should have these columns: %s", line)
	return reader, nil
}

func (p *Processor) getFilePath(fileName string) (string, error) {
	if filepath.Ext(fileName) != constants.AcceptedExt {
		log.Println("File type not accepted.")
		return "", errors.New("Should be a CSV file")
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Println("Cannot get the working directory")
		return "", err
	}
	return filepath.Join(wd, fileName), nil
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
		go p.processRow(&w, p.job, p.runningWorkers, p.runCh)
	}

	p.poolWorker = workers
}

func (p *Processor) processRow(ch *chan []string, job, runningWorkers *sync.WaitGroup, runCh *chan error) {
	ok := true
	for row := range *ch {
		if ok {
			err := p.db.Insert(row...)
			if err != nil {
				ok = false
				*runCh <- err
			}
		}
		job.Done()
	}
	runningWorkers.Done()
}

func (p *Processor) finish() {
	p.csvFile.Close()
	p.job.Wait()
	for _, ch := range p.poolWorker {
		close(ch)
	}
	p.runningWorkers.Wait()
	p.db.Close()
	log.Printf("Every worker's channels has been closed")
}

func (p *Processor) balanceLoad(line []string) {
	// ID is used to load balance jobs between workers.
	if id, err := strconv.Atoi(line[0]); err == nil {
		w := id % constants.Workers
		p.job.Add(1)
		p.poolWorker[w] <- line
	} else {
		log.Printf("Skipping Line: Cannot Parse an ID")
	}
}
