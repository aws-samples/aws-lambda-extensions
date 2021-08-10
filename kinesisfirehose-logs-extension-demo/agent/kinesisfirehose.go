// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package agent

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

var logger = log.WithFields(log.Fields{"agent": "logsApiAgent"})

const (
	// MaxRetries maximum retry attempt
	MaxRetries    = 5
	// sleepDuration wait before next retry to implement exponential backoff
	sleepDuration = 2 * time.Second
)

// KinesisFirehoseLogger is the logger that writes the logs received from Logs API to kinesis firehose
type KinesisFirehoseLogger struct {
	svc    *firehose.Firehose
	record *firehose.Record
	stream string
}

// NewKinesisFirehoseLogger returns an Kinesis Logger
func NewKinesisFirehoseLogger() (*KinesisFirehoseLogger, error) {
	return &KinesisFirehoseLogger{
		svc:    firehose.New(session.New()),
		stream: os.Getenv("AWS_KINESIS_STREAM_NAME"),
	}, nil
}

// PushLog converts to byte array and pushes it to kinesis firehose
func (l *KinesisFirehoseLogger) PushLog(data string) error {
	l.record = &firehose.Record{Data: append([]byte(data), '\n')}
	return l.writeLogs(0, sleepDuration)
}

// Shutdown calls the function that should be executed before the program terminates
func (l *KinesisFirehoseLogger) Shutdown() error {
	if l.record != nil {
		err := l.writeLogs(0, sleepDuration)
		if err != nil {
			return err
		}
	}

	return nil
}

// Send logs to kinesis firehose
func (l *KinesisFirehoseLogger) writeLogs(retry int, sleep time.Duration) error {
	// Check for retry attempt before proceeding
	if retry < MaxRetries {
		_, err := l.svc.PutRecord(&firehose.PutRecordInput{
			DeliveryStreamName: aws.String(l.stream),
			Record:             l.record,
		})

		// Check for error
		if err != nil {
			logger.Errorf("error while writing records to firehose %s", err.Error())
			if aErr, ok := err.(awserr.Error); ok {
				switch aErr.Code() {
				// Retry in case of `ErrCodeServiceUnavailableException` error to perform retry
				case firehose.ErrCodeServiceUnavailableException:
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
func (l *KinesisFirehoseLogger) retryPut(retry int, sleep time.Duration) error {
	time.Sleep(sleep)
	retry++
	return l.writeLogs(retry, sleep*2)
}
