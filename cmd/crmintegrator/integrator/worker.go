package integrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/josesolana/csv-reader/cmd/models"
	"github.com/josesolana/csv-reader/constants"
	"github.com/jpillora/backoff"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type worker struct {
	source chan models.Customer
	quit   *chan interface{}
	cancel context.CancelFunc
	cx     context.Context
}

func (w *worker) Start(runningWorkers *sync.WaitGroup) {
	cx, cancel := context.WithCancel(context.Background())
	w.cx = cx
	w.cancel = cancel
	go func() {
		for {
			select {
			case customer := <-w.source:
				w.makeRequest(&customer)
			case <-*w.quit:
				w.cancel()
				runningWorkers.Done()
				return
			}
		}
	}()
}

func (w *worker) makeRequest(customer *models.Customer) {
	var sleep time.Duration

	backOff := &backoff.Backoff{
		// Min value to retry, if Json Api is down(It starts at Min)
		Min: 1 * time.Second,
		// After every call to Duration() it is multiplied by Factor
		Factor: 1.1,
		// It is capped at Max
		Max: 5 * time.Minute,
		// Adds some randomization to the backoff durations.
		Jitter: true,
	}

	httpClient := http.Client{
		Timeout: constants.TimeOut,
	}

	for retry := 0; retry < constants.TotalRetry; retry++ {
		select {
		case <-time.After(sleep):
			postReq, err := w.preparePostRequest(customer)
			if err != nil {
				sleep = backOff.Duration()
				break
			}
			resp, err := httpClient.Do(postReq)
			if err != nil {
				sleep = backOff.Duration()
				break
			}
			defer resp.Body.Close()
			if resp.StatusCode > http.StatusBadRequest {
				sleep = backOff.Duration()
				break
			}

			fmt.Println("Json Api Add Customer. ID: ", customer.ID)
			return
		}
	}
}

func (w *worker) preparePostRequest(customer *models.Customer) (*http.Request, error) {
	var url string
	body, err := json.Marshal(customer)
	if err != nil {
		log.Printf("Cannot serialize Customer. Skipping ID: %d. Error: %s\n", customer.ID, err)
		return nil, err
	}

	if rand.Intn(100) > 40 {
		url = constants.CRMUrlFail
	} else {
		url = constants.CRMUrl
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Cannot make a POST request. Skipping ID: %d. Error: %s\n", customer.ID, err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(w.cx)
	return req, nil
}
