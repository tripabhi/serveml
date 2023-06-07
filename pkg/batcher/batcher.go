package batcher

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"time"
)


const (
	SleepTime = time.Microsecond * 100
	MaxBatchSize = 32
	MaxLatency = 5000
)

type Request struct {
	Queries []interface{} `json:"queries"`
}


type Response struct {
	Message string `json:"message"`
	Prediction interface{} `json:"prediction"`
}

type PredictionResponse struct {
	Predictions []interface{} `json:"predictions"`
}

type Input struct {
	ContextInput *context.Context
	Queries *[]interface{}
	Path string
	ChannelOut *chan Response
}

type InputInfo struct {
	ChannelOut *chan Response
	Index []int
}

type BatcherInfo struct {
	BatchID string
	Path string
	Request *http.Request
	Queries []interface{}
	PredictionResponse PredictionResponse
	ContextMap map[*context.Context]InputInfo
	Start time.Time
	Now time.Time
}

type BatchHandler struct {
	next http.Handler
	MaxBatchSize int
	MaxLatency int
	channelIn chan Input
	batcherInfo BatcherInfo
}

func (info *BatcherInfo) Init() {
	info.Queries = make([]interface{}, 0)
	info.BatchID = ""
	info.ContextMap = make(map[*context.Context]InputInfo)
	info.PredictionResponse = PredictionResponse{}
	info.Start = time.Now().UTC()
}

func New(maxBatchSize int, maxLatency int, handler http.Handler) *BatchHandler {
	batchHandler := BatchHandler{
		next : handler,
		channelIn: make(chan Input, maxBatchSize),
		MaxBatchSize: maxBatchSize,
		MaxLatency: maxLatency,
	}
	batchHandler.batcherInfo.Init()
	go batchHandler.Consume()
	return &batchHandler

}

func (handler *BatchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var containsPredictPattern = regexp.MustCompile(`predict$`)
	if !containsPredictPattern.MatchString(r.URL.Path) {
		handler.next.ServeHTTP(w, r)
		return
	}
	
	var req	Request
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

	var ctx = context.Background()
	var outChan = make(chan Response)

	handler.channelIn <- Input{
		&ctx,
		&req.Queries,
		r.URL.Path,
		&outChan,
	}

	response := <-outChan
	close(outChan)
	responseJson, err := json.Marshal(response)
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
}

func (handler *BatchHandler) predict() {
	// log.Println("Running Batch Prediction")
	batchReqJson, _ := json.Marshal(Request{
		handler.batcherInfo.Queries,
	})

	reader := bytes.NewReader(batchReqJson)
	req := httptest.NewRequest(http.MethodPost, handler.batcherInfo.Path, reader)
	recorder := httptest.NewRecorder()

	handler.next.ServeHTTP(recorder, req)
	// log.Println("Received Response from Inference Server")

	responseBody := recorder.Body.Bytes()

	if recorder.Code != http.StatusOK {
		for _, v := range handler.batcherInfo.ContextMap {
			res := Response{
				Message: string(responseBody),
				Prediction: nil,
			}
			*v.ChannelOut <- res
		}
	} else {
		err := json.Unmarshal(responseBody, &handler.batcherInfo.PredictionResponse)
		// log.Println("Unmarshalled Response")
		if err != nil {

		} else {
			if len(handler.batcherInfo.PredictionResponse.Predictions) != len(handler.batcherInfo.Queries) {
				for _, v := range handler.batcherInfo.ContextMap {
					res := Response{
						Message: "size of prediction is not equal to the size of instances",
					}
					*v.ChannelOut <- res
				}
			} else {
				// log.Printf("Length of contextMap %d\n", len(handler.batcherInfo.ContextMap))
				for _, v := range handler.batcherInfo.ContextMap {
					predictions := make([]interface{}, 0)
					for _, i := range v.Index {
						predictions = append(predictions, 
							handler.batcherInfo.PredictionResponse.Predictions[i])
					}
					res := Response{
						Message:     "",
						Prediction: predictions,
					}
					// log.Printf("Returning response %+v\n", res)
					*v.ChannelOut <- res
				}
			}
		}

	}
	handler.batcherInfo.Init()
}

func (handler *BatchHandler) workerLoop() {
	for {
		select {
		case req := <-handler.channelIn:
			if len(handler.batcherInfo.Queries) == 0 {
				handler.batcherInfo.Start = time.Now().UTC()
			}
			handler.batcherInfo.Path = req.Path
			var currentLen = len(handler.batcherInfo.Queries)
			handler.batcherInfo.Queries = append(handler.batcherInfo.Queries, *req.Queries...)
			
			var index = make([]int, 0)

			for i:= 0; i < len(*req.Queries); i++ {
				index = append(index, currentLen + i)
			}

			handler.batcherInfo.ContextMap[req.ContextInput] = InputInfo{
				req.ChannelOut,
				index,
			}
		case <-time.After(SleepTime):
		}
		if len(handler.batcherInfo.Queries) >= handler.MaxBatchSize || 
			(time.Now().UTC().Sub(handler.batcherInfo.Start).Milliseconds() >= int64(handler.MaxLatency) && len(handler.batcherInfo.Queries) > 0) {
				handler.predict()
		}
	}
}



func (handler *BatchHandler) Consume() {
	if handler.MaxBatchSize <= 0 {
		handler.MaxBatchSize = MaxBatchSize
	}
	if handler.MaxLatency <= 0 {
		handler.MaxLatency = MaxLatency
	}
	handler.batcherInfo.Init()
	handler.workerLoop()
}