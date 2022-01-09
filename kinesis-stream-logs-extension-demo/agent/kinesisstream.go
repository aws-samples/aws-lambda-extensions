// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package agent

import (
	"errors"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var logger = log.WithFields(log.Fields{"agent": "logsApiAgent"})

const (
	// MaxRetries maximum retry attempt
	MaxRetries = 5
	// sleepDuration wait before next retry to implement exponential backoff
	sleepDuration = 2 * time.Second
)

// KinesisStreamLogger is the logger that writes the logs received from Logs API to kinesis stream
type KinesisStreamLogger struct {
	svc    *kinesis.Kinesis
	record *kinesis.Record
	stream string
}

// NewKinesisStreamLogger returns an Kinesis Logger
func NewKinesisStreamLogger() (*KinesisStreamLogger, error) {
	return &KinesisStreamLogger{
		svc:    kinesis.New(session.New()),
		stream: os.Getenv("AWS_KINESIS_STREAM_NAME"),
	}, nil
}

// PushLog converts to byte array and pushes it to kinesis stream
func (l *KinesisStreamLogger) PushLog(data string) error {
	l.record = &kinesis.Record{Data: append([]byte(data), '\n')}
	return l.writeLogs(0, sleepDuration)
}

// Shutdown calls the function that should be executed before the program terminates
func (l *KinesisStreamLogger) Shutdown() error {
	if l.record != nil {
		err := l.writeLogs(0, sleepDuration)
		if err != nil {
			return err
		}
	}

	return nil
}

// Send logs to kinesis stream
func (l *KinesisStreamLogger) writeLogs(retry int, sleep time.Duration) error {
	// Check for retry attempt before proceeding
	if retry < MaxRetries {
		_, err := l.svc.PutRecord(&kinesis.PutRecordInput{
			StreamName:   aws.String(l.stream),
			Data:         l.record.Data,
			PartitionKey: aws.String(uuid.New().String()),
		})

		// Check for error
		if err != nil {
			logger.Errorf("error while writing records to kinesis stream %s", err.Error())
			if aErr, ok := err.(awserr.Error); ok {
				switch aErr.Code() {
				// Retry in case of `ErrCodeProvisionedThroughputExceededException` error to perform retry
				case kinesis.ErrCodeProvisionedThroughputExceededException:
					return l.retryPut(retry, sleep)
				default:
					return err
				}
			}
			return err
		}

		l.record = nil
	} else {
		return errors.New("maximum retries has exceeded")
	}

	return nil
}

// Perform exponential backoff
func (l *KinesisStreamLogger) retryPut(retry int, sleep time.Duration) error {
	time.Sleep(sleep)
	retry++
	return l.writeLogs(retry, sleep*2)
}
