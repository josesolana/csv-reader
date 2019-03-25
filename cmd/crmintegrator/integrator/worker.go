package integrator

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/josesolana/csv-reader/cmd/crmintegrator/database"
	c "github.com/josesolana/csv-reader/constants"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type worker struct {
	sourceCh chan []interface{}
	quitCh   chan interface{}
	errorCh  chan error
	db       *database.DB
	cancel   context.CancelFunc
	cx       *context.Context
}

func (w *worker) Start(runningWorkers, jobs *sync.WaitGroup) {
	cx, cancel := context.WithCancel(context.Background())
	w.cx = &cx
	w.cancel = cancel

	go func() {
		var err error
		for {
			select {
			case vals := <-w.sourceCh:
				err = w.makeRequest(vals)
				jobs.Done()
				if err != nil {
					w.errorCh <- err
				}
			case <-w.quitCh:
				w.cancel()
				runningWorkers.Done()
				return
			}
		}
	}()
}

func (w *worker) makeRequest(vals []interface{}) error {
	id := *vals[c.IDPos].(*int)
	httpClient := http.Client{
		Timeout: c.TimeOut,
	}
	postReq, err := w.preparePostRequest(vals)

	if err != nil {
		log.Println("Fail creating POST request", err)
		return (*w.db).IncreaseRetry(id)
	}

	resp, err := httpClient.Do(postReq)
	if err != nil {
		log.Println("Cannot make a request to JSON API")
		return (*w.db).IncreaseRetry(id)
	}
	defer resp.Body.Close()

	if resp.StatusCode > http.StatusBadRequest {
		log.Println("JSON API response a Bad Request")
		return (*w.db).IncreaseRetry(id)
	}

	return (*w.db).SetAsProcessed(id)
}

func (w *worker) preparePostRequest(vals []interface{}) (*http.Request, error) {
	var url string

	// Skipped those values whom has been added to handle row flow.
	body, err := json.Marshal(vals[3:])
	if err != nil {
		log.Printf("Cannot serialize a row. ID: %d. Error: %s\n", vals[c.IDPos], err)
		return nil, err
	}

	//Fail rate 60%
	if rand.Intn(100) > 40 {
		url = c.CRMUrlFail
	} else {
		url = c.CRMUrl
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Cannot make a POST request. ID: %d. Error: %s\n", vals[c.IDPos], err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(*w.cx)
	return req, nil
}
