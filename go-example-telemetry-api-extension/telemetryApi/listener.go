// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package telemetryApi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
)

const defaultListenerPort = "4323"
const initialQueueSize = 5

// Used to listen to the Telemetry API
type TelemetryApiListener struct {
	httpServer *http.Server
	// LogEventsQueue is a synchronous queue and is used to put the received log events to be dispatched later
	LogEventsQueue *queue.Queue
}

func NewTelemetryApiListener() *TelemetryApiListener {
	return &TelemetryApiListener{
		httpServer:     nil,
		LogEventsQueue: queue.New(initialQueueSize),
	}
}

func listenOnAddress() string {
	env_aws_local, ok := os.LookupEnv("AWS_SAM_LOCAL")
	var addr string
	if ok && env_aws_local == "true" {
		addr = ":" + defaultListenerPort
	} else {
		addr = "sandbox:" + defaultListenerPort
	}

	return addr
}

// Starts the server in a goroutine where the log events will be sent
func (s *TelemetryApiListener) Start() (string, error) {
	address := listenOnAddress()
	l.Info("[listener:Start] Starting on address", address)
	s.httpServer = &http.Server{Addr: address}
	http.HandleFunc("/", s.http_handler)
	go func() {
		err := s.httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
			l.Error("[listener:goroutine] Unexpected stop on Http Server:", err)
			s.Shutdown()
		} else {
			l.Info("[listener:goroutine] Http Server closed:", err)
		}
	}()
	return fmt.Sprintf("http://%s/", address), nil
}

// http_handler handles the requests coming from the Telemetry API.
// Everytime Telemetry API sends log events, this function will read them from the response body
// and put into a synchronous queue to be dispatched later.
// Logging or printing besides the error cases below is not recommended if you have subscribed to
// receive extension logs. Otherwise, logging here will cause Telemetry API to send new logs for
// the printed lines which may create an infinite loop.
func (s *TelemetryApiListener) http_handler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Error("[listener:http_handler] Error reading body:", err)
		return
	}

	// Parse and put the log messages into the queue
	var slice []interface{}
	_ = json.Unmarshal(body, &slice)

	for _, el := range slice {
		s.LogEventsQueue.Put(el)
	}

	l.Info("[listener:http_handler] logEvents received:", len(slice), " LogEventsQueue length:", s.LogEventsQueue.Len())
	slice = nil
}

// Terminates the HTTP server listening for logs
func (s *TelemetryApiListener) Shutdown() {
	if s.httpServer != nil {
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			l.Error("[listener:Shutdown] Failed to shutdown http server gracefully:", err)
		} else {
			s.httpServer = nil
		}
	}
}
