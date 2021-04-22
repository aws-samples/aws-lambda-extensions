// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package queuewrapper

import (
    "fmt"
    "github.com/golang-collections/go-datastructures/queue"
    "sync"
)

// Wrapper around queue, includes the size of the queue in bytes
type QueueWrapper struct {
    simpleQueue *queue.Queue  // queue
    queueSize   int64         // Size of the queue
    queueLock   *sync.RWMutex // Mutex to protect QueueWrapper
}

// Create and initialize the queue
func New(hint int64) *QueueWrapper {
    return &QueueWrapper{
        simpleQueue: queue.New(hint),
        queueSize:   0,
        queueLock:   &sync.RWMutex{},
    }
}

// Abstracts the put function, still puts item on queue in a thread safe
// manner
func (qw *QueueWrapper) Put(items ...interface{}) error {
    qw.queueLock.Lock()
    defer qw.queueLock.Unlock()
    for _, item := range items {
        str := fmt.Sprintf("%v", item)
        qw.queueSize += int64(len(str))
    }
    err := qw.simpleQueue.Put(items)

    return err
}

// Get number of items off the queue
func (qw *QueueWrapper) Get(number int64) ([]interface{}, error) {
    qw.queueLock.Lock()
    defer qw.queueLock.Unlock()

    // Call get to underlying queue
    poppedItems, error := qw.simpleQueue.Get(number)

    // Start subtracting the size of the log
    for _, item := range poppedItems {
        qw.queueSize -= int64(len(fmt.Sprintf("%v", item)))
    }

    return poppedItems, error
}

// How much the queue is storing in bytes
func (qw *QueueWrapper) Size() int64 {
    qw.queueLock.RLock()
    defer qw.queueLock.RUnlock()
    return qw.queueSize
}

// The number of elements in the queue
func (qw *QueueWrapper) Len() int64 {
    qw.queueLock.RLock()
    defer qw.queueLock.RUnlock()
    return qw.simpleQueue.Len()
}

// Is the queue empty?
func (qw *QueueWrapper) Empty() bool {
    qw.queueLock.RLock()
    defer qw.queueLock.RUnlock()
    return qw.simpleQueue.Empty()
}
