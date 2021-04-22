// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package agent

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"aws-lambda-extensions/go-example-adaptive-batching-extension/queuewrapper"
	log "github.com/sirupsen/logrus"
)

var metricLogger = log.WithFields(log.Fields{"agent": "metricsAgent"})

const (
	DEFAULT_SHIP_RATE_BYTES        int   = 4 * 1024  // Default value for ship rate 4kb
	DEFAULT_SHIP_RATE_INVOKES      int   = 10        // Default invoke rate
	DEFAULT_SHIP_RATE_MILLISECONDS int64 = 10 * 1000 // Default time interval in milliseconds

	// Maximum values that rates can hold. If exceeded, the maximum value will be assumed.
	// Protect users from using too much lambda memory.
	MAX_SHIP_RATE_BYTES int = 50 * 1024 * 1024 // Maximum for bytes ship rate 100 megabytes
)

type MetricsMonitor struct {
	lq                   *queuewrapper.QueueWrapper
	invokeCount          int       // Number of invokes that have occured since last ship
	lastShipTime         time.Time // Last time logs were shipped
	shipRateBytes        int       // The maximum number of bytes between shipping logs
	shipRateInvokes      int       // Maximum number of invokes since last time logs were shipped
	shipRateMilliseconds int64     // Maximum amount of time between shipping logs
}

// Creates a new MetricsMonitor
func NewMetricsMonitor(lq *queuewrapper.QueueWrapper) *MetricsMonitor {

	// Set the proper log shipping rates
	shipRateBytes := int(retrieveEnvironmentVariable(
		"ADAPTIVE_BATCHING_EXTENSION_SHIP_RATE_BYTES",
		int64(DEFAULT_SHIP_RATE_BYTES), int64(MAX_SHIP_RATE_BYTES)))
	shipRateInvokes := int(retrieveEnvironmentVariable(
		"ADAPTIVE_BATCHING_EXTENSION_SHIP_RATE_INVOKES",
		int64(DEFAULT_SHIP_RATE_INVOKES), int64(-1)))
	shipRateMilliseconds := retrieveEnvironmentVariable(
		"ADAPTIVE_BATCHING_EXTENSION_SHIP_RATE_MILLISECONDS",
		DEFAULT_SHIP_RATE_MILLISECONDS, -1)

	return &MetricsMonitor{
		lq:                   lq,
		invokeCount:          0,
		lastShipTime:         time.Now(),
		shipRateBytes:        shipRateBytes,
		shipRateInvokes:      shipRateInvokes,
		shipRateMilliseconds: shipRateMilliseconds,
	}
}

// Count a new invoke
func (monitor *MetricsMonitor) CountInvoke() {
	monitor.invokeCount++
}

// Returns whether logs should be shipped or not based on the rates
func (monitor *MetricsMonitor) ShouldShip() bool {

	timeElapsed := int64(time.Now().Sub(monitor.lastShipTime) / 1e6)

	bytesInQueue := monitor.lq.Size()

	shouldShip := false
	if monitor.invokeCount >= monitor.shipRateInvokes {
		shouldShip = true
		metricLogger.Info("Invoke threshold met, ", monitor.invokeCount, " invokes counted.")
	} else if timeElapsed >= monitor.shipRateMilliseconds {
		shouldShip = true
		metricLogger.Info("Time threshold met, ", timeElapsed, " milliseconds elapsed.")
	} else if bytesInQueue >= int64(monitor.shipRateBytes) {
		shouldShip = true
		metricLogger.Info("Log size threshold met, log is ", bytesInQueue, " bytes .")
	}

	return shouldShip
}

// Reset the monitor.
// Zero out all the counters, set time to current time
func (monitor *MetricsMonitor) Reset() {
	monitor.invokeCount = 0
	monitor.lastShipTime = time.Now()
}

// Helper function
// Given an environment variable, read it in and set the proper value
// Returns the value received, value will never exceed the maximum value
func retrieveEnvironmentVariable(variable string, defaultValue int64, maxValue int64) (output int64) {
	// Fetch environment variable
	valueString, present := os.LookupEnv(variable)

	// Verify it exists
	if !present {
		metricLogger.Info(variable, " not set. Using default value ", defaultValue)
		return defaultValue
	}

	valueInteger, conversionError := strconv.ParseInt(valueString, 10, 64)

	// Check if conversion fails
	// Do nothing if it does
	if conversionError != nil {
		metricLogger.Info(variable, " cannot be converted to integer. Using default value ",
			defaultValue)
		return defaultValue
	}

	// Check if the value exceeds maximum value.
	// If max value is -1, we ignore the max value
	if maxValue < valueInteger && maxValue != -1 {
		metricLogger.Info(variable, " (", valueInteger, ") provided is greater than maximum accepted value(",
			maxValue, "), using maximum accepted value.")
		return maxValue
	}

	// Set the new value
	metricLogger.Info(variable, " set to ", valueInteger)
	return valueInteger
}

// String function
func (monitor *MetricsMonitor) String() string {
	queueSize := strconv.FormatInt(monitor.lq.Size(), 10)
	timeElapsed := int(time.Now().Sub(monitor.lastShipTime) / 1e6)
	outputString := fmt.Sprintf("Invokes: %d Time elapsed: %d Queue Size: %s",
		monitor.invokeCount,
		timeElapsed,
		queueSize)

	return outputString
}
