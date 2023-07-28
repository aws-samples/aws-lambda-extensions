// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package telemetryApi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-collections/go-datastructures/queue"
)

type Dispatcher struct {
	httpClient   *http.Client
	postUri      string
	minBatchSize int64
}

func NewDispatcher() *Dispatcher {
	dispatchPostUri := os.Getenv("DISPATCH_POST_URI")
	if len(dispatchPostUri) == 0 {
		panic("dispatchPostUri undefined")
	}

	dispatchMinBatchSize, err := strconv.ParseInt(os.Getenv("DISPATCH_MIN_BATCH_SIZE"), 0, 16)
	if err != nil {
		dispatchMinBatchSize = 1
	}

	return &Dispatcher{
		httpClient:   &http.Client{},
		postUri:      dispatchPostUri,
		minBatchSize: dispatchMinBatchSize,
	}

}

func (d *Dispatcher) Dispatch(ctx context.Context, logEventsQueue *queue.Queue, force bool) {
	l.Info("[dispatcher:Dispatch] Dispatching", logEventsQueue.Len(), "log events")
	logEntries, _ := logEventsQueue.Get(logEventsQueue.Len())
	body, err := json.Marshal(logEntries)
	if err != nil {
		l.Error("[dispatcher:Dispatch] Failed to marshal log entries", err)
		return
	}
	l.Info("[dispatcher:Dispatch] Dispatched", logEventsQueue.Len(), "log events")
	if strings.Contains(string(body), "logsDropped") {
		l.Info("[dispatcher:Dispatch] LOG DROPPED!")
	}
}
