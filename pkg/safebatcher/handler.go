package safebatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"time"
)

type Request struct {
	Queries []interface{} `json:"queries"`
}

type Response struct {
	Message     string        `json:"message"`
	Metrics     interface{}   `json:"metrics"`
	Predictions []interface{} `json:"predictions"`
}

type PredictionResponse struct {
	Metrics     interface{}   `json:"metrics"`
	Predictions []interface{} `json:"predictions"`
}

type Batch struct {
	Queries     []interface{}
	Path        string
	subscribers []BatchSubscriber
	timer       *time.Timer
}

type BatchSubscriber struct {
	rangeStart int
	rangeEnd   int
	respCh     chan *Response
}

type Config struct {
	MaxBatchSize int
	MaxLatency   time.Duration
	IdleTimeout  time.Duration
}

type PredictionRequestBatcher struct {
	sync.Mutex
	next         http.Handler
	MaxBatchSize int
	MaxLatency   time.Duration
	IdleTimeout  time.Duration
	running      bool
	parentCtx    context.Context
	currBatch    *Batch
}

func NewPredictionRequestBatcher(
	ctx context.Context,
	config *Config,
	next http.Handler,
) *PredictionRequestBatcher {

	if next == nil {
		log.Fatal("cannot create Batcher without next handler")
	}

	batcher := &PredictionRequestBatcher{
		next:         next,
		MaxBatchSize: config.MaxBatchSize,
		MaxLatency:   config.MaxLatency,
		IdleTimeout:  config.IdleTimeout,
		running:      true,
		parentCtx:    ctx,
	}

	go func(b *PredictionRequestBatcher) {
		<-b.parentCtx.Done()
		log.Println("Received TERM signal for parent")
		b.stop()
	}(batcher)

	return batcher
}

func (b *PredictionRequestBatcher) stop() {
	b.Lock()
	defer b.Unlock()
	log.Println("Stopping Batcher")

	b.running = false
	if b.currBatch != nil {
		b.currBatch.timer.Stop()
		for i := len(b.currBatch.subscribers) - 1; i >= 0; i-- {
			close(b.currBatch.subscribers[i].respCh)
		}
	}
}

func (b *PredictionRequestBatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var containsPredictPattern = regexp.MustCompile(`predict$`)
	if !containsPredictPattern.MatchString(r.URL.Path) {
		b.next.ServeHTTP(w, r)
		return
	}

	var req Request
	var err error

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Cannot read Request Body", http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Cannot Unmarshall Request Body", http.StatusBadRequest)
		return
	}

	respCh, _ := b.submitQueries(&req, r.URL.Path)

	ctx, cancel := context.WithTimeout(b.parentCtx, b.IdleTimeout)
	defer cancel()

	select {
	case resp := <-respCh:
		responseJson, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(responseJson)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case <-ctx.Done():
		http.Error(w, "Request Timeout", http.StatusRequestTimeout)
		return
	}
}

func (b *PredictionRequestBatcher) submitQueries(
	request *Request,
	path string,
) (<-chan *Response, error) {

	respCh := make(chan *Response, 1)

	sub := BatchSubscriber{
		respCh: respCh,
	}

	b.Lock()
	defer b.Unlock()

	if b.currBatch != nil {
		if len(b.currBatch.Queries) < b.MaxBatchSize {

			sub.rangeStart = len(b.currBatch.Queries)
			sub.rangeEnd = sub.rangeStart + len(request.Queries)

			b.currBatch.Queries = append(b.currBatch.Queries, request.Queries...)
			b.currBatch.subscribers = append(b.currBatch.subscribers, sub)

			if len(b.currBatch.Queries) >= b.MaxBatchSize {
				b.currBatch.timer.Stop()
				b.forwardCurrentBatch()

			}

			return respCh, nil
		}

		b.currBatch.timer.Stop()
		b.forwardCurrentBatch()
	}

	b.currBatch = &Batch{
		Queries:     make([]interface{}, 0),
		subscribers: make([]BatchSubscriber, 0),
		Path:        path,
	}
	sub.rangeStart = 0
	sub.rangeEnd = len(request.Queries)
	b.currBatch.Queries = append(b.currBatch.Queries, request.Queries...)
	b.currBatch.subscribers = append(b.currBatch.subscribers, sub)

	b.currBatch.timer = time.AfterFunc(b.MaxLatency, b.forwardCurrentBatchWithMutex)

	return respCh, nil

}

func (b *PredictionRequestBatcher) forwardCurrentBatch() {
	batch := b.currBatch
	b.currBatch = nil

	if batch != nil {
		go func(fbatch *Batch) {
			b.forward(fbatch)
		}(batch)
	}
}

func (b *PredictionRequestBatcher) forwardCurrentBatchWithMutex() {
	b.Lock()
	batch := b.currBatch
	b.currBatch = nil
	b.Unlock()

	if batch != nil {
		go func(fbatch *Batch) {
			b.forward(fbatch)
		}(batch)
	}
}

func (batcher *PredictionRequestBatcher) forward(batch *Batch) {

	batchReqJson, _ := json.Marshal(Request{
		batch.Queries,
	})

	reader := bytes.NewReader(batchReqJson)
	req := httptest.NewRequest(http.MethodPost, batch.Path, reader)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	batcher.next.ServeHTTP(recorder, req)

	responseBody := recorder.Body.Bytes()

	if recorder.Code != http.StatusOK {
		for i := len(batch.subscribers) - 1; i >= 0; i-- {
			batch.subscribers[i].respCh <- &Response{
				Message:     "inference service replied with non-200 status code",
				Predictions: nil,
			}
			close(batch.subscribers[i].respCh)
		}
		return
	}

	var resp PredictionResponse
	err := json.Unmarshal(responseBody, &resp)

	if err != nil {
		for i := len(batch.subscribers) - 1; i >= 0; i-- {
			batch.subscribers[i].respCh <- &Response{
				Message:     "unable to unmarshall response",
				Predictions: nil,
			}
			close(batch.subscribers[i].respCh)
		}
		return
	} else {
		if len(resp.Predictions) != len(batch.Queries) {
			for i := len(batch.subscribers) - 1; i >= 0; i-- {
				batch.subscribers[i].respCh <- &Response{
					Message:     "Length of response from inference service does not match queries",
					Predictions: nil,
				}
				close(batch.subscribers[i].respCh)
			}
			return
		} else {
			for i := len(batch.subscribers) - 1; i >= 0; i-- {
				start := batch.subscribers[i].rangeStart
				end := batch.subscribers[i].rangeEnd

				batch.subscribers[i].respCh <- &Response{
					Message:     "Successful",
					Predictions: resp.Predictions[start:end],
					Metrics:     resp.Metrics,
				}
				close(batch.subscribers[i].respCh)
			}
		}
	}
}
