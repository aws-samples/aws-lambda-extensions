// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"aws-lambda-extensions/go-example-adaptive-batching-extension/agent"
	"aws-lambda-extensions/go-example-adaptive-batching-extension/extension"
	"aws-lambda-extensions/go-example-adaptive-batching-extension/queuewrapper"
	log "github.com/sirupsen/logrus"
)

// INITIAL_QUEUE_SIZE is the initial size set for the synchronous logQueue
const INITIAL_QUEUE_SIZE = 5

func main() {

	extensionName := path.Base(os.Args[0])
	printPrefix := fmt.Sprintf("[%s]", extensionName)
	logger := log.WithFields(log.Fields{"agent": extensionName})

	extensionClient := extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		logger.Info(printPrefix, "Received", s)
		logger.Info(printPrefix, "Exiting")
	}()

	// Register extension as soon as possible
	_, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}

	// Create S3 Logger
	logsApiLogger, err := agent.NewS3Logger()
	if err != nil {
		logger.Fatal(err)
	}

	// A synchronous queue that is used to put logs from the goroutine (producer)
	// and process the logs from main goroutine (consumer)
	logQueue := queuewrapper.New(INITIAL_QUEUE_SIZE)

	// Helper function to empty the log queue
	flushLogQueue := func() {
		logger.Info(printPrefix, "Flush Queue")
		for !logQueue.Empty() {
			logs, err := logQueue.Get(1)
			if err != nil {
				logger.Error(printPrefix, err)
				return
			}
			logString := fmt.Sprintf("%v", logs[0])
			// write log to logger
			logsApiLogger.WriteLog(logString)
		}
	}

	// Create Logs API agent
	logsApiAgent, err := agent.NewHttpAgent(logsApiLogger, logQueue)
	if err != nil {
		logger.Fatal(err)
	}

	// Subscribe to logs API
	// Logs start being delivered only after the subscription happens.
	agentID := extensionClient.ExtensionID
	err = logsApiAgent.Init(agentID)
	if err != nil {
		logger.Fatal(err)
	}

	// Initialize metrics monitor
	monitor := agent.NewMetricsMonitor(logQueue)

	// Will block until invoke or shutdown event is received or cancelled
	// via the context.
	for {
		select {
		case <-ctx.Done():
			return
		default:
			logger.Info(printPrefix, " Waiting for event...")
			// This is a blocking call
			res, err := extensionClient.NextEvent(ctx)

			if err != nil {
				logger.Info(printPrefix, "Error:", err)
				logger.Info(printPrefix, "Exiting")
				return
			}

			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				logger.Info(printPrefix, "Received SHUTDOWN event")
				flushLogQueue()
				logsApiAgent.Shutdown()
				logger.Info(printPrefix, "Exiting")
				return
			}

			// Tell the monitor an invoke has occured
			monitor.CountInvoke()

			// Flush logs if monitor has reached its thresholds
			if monitor.ShouldShip() {
				// Print the metrics
				logger.Info(monitor.String())

				// Flush the Queue
				flushLogQueue()

				// Ship the logs to S3
				err = logsApiLogger.FlushLog()
				if err != nil {
					logger.Errorf("Error shipping to S3: %v", err)
				}

				// Reset the monitor
				monitor.Reset()
			}

		}
	}
}
