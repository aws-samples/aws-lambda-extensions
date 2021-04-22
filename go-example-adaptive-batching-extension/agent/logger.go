// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package agent

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var logger = log.WithFields(log.Fields{"agent": "logsApiAgent"})

const (
	// MAX_RETRIES is the maximum number or retrials before failing when uploading part operation fails.
	MAX_RETRIES = 3
	// MAX_PART_SIZE is the minimum size that a part needs to have when uploading to S3 with multipart uploader.
	// Only the final part is allowed to be smaller than 5MB for a successful multipart upload.
	// 5 * 1024 * 1024 = 5MB = 5242880B
	MAX_PART_SIZE = 5242880
)

// S3Logger is the logger that writes the logs received from Logs API to S3
type S3Logger struct {
	svc       *session.Session
	bucket    string
	logBuffer *bytes.Buffer
	prefix    string
	fileName  string
	uploader  *s3manager.Uploader
}

// NewS3Logger returns an S3 Logger
func NewS3Logger() (*S3Logger, error) {
	// Find bucket name
	bucket, present := os.LookupEnv("ADAPTIVE_BATCHING_EXTENSION_S3_BUCKET")
	if !present {
		return nil, errors.New("Environment variable ADAPTIVE_BATCHING_EXTENSION_S3_BUCKET is not set.")
	} else {
		fmt.Println("Sending logs to:", bucket)
	}
	// Setup buffer
	buffer := bytes.NewBuffer([]byte(""))
	buffer.Grow(2 * MAX_PART_SIZE)

	// Create the S3 Bucket
	err := createBucket(bucket)
	if err != nil {
		logger.Error("Error creating S3 Bucket")
		return nil, err
	}

	// Create the prefix for the s3 file. Unique to the sandbox environment that this extension is running in
	// Format {year}-{month}-{day}-{uuid}
	environmentId := uuid.New().String()
	t := time.Now().Format("2006-01-02")
	prefix := t + "-" + environmentId + "/"

	logger.Info("Environment ID: " + environmentId)

	// Create filename
	fileName := generateFileName()

	svc := session.Must(session.NewSession())

	// Initialize uploader
	uploader := s3manager.NewUploader(svc, func(u *s3manager.Uploader) {
		u.PartSize = MAX_PART_SIZE
	})

	return &S3Logger{
		svc:       svc,
		bucket:    bucket,
		logBuffer: buffer,
		fileName:  fileName,
		prefix:    prefix,
		uploader:  uploader,
	}, nil
}

// ResetLogger renames the logger and resets the logBuffer
func (l *S3Logger) reset() {
	l.logBuffer.Reset()
	l.logBuffer.Grow(2 * MAX_PART_SIZE)
	l.fileName = generateFileName()

}

// PushLog writes the received logs to a buffer
func (l *S3Logger) WriteLog(log string) {
	l.logBuffer.Write([]byte(log))
}

// FlushLog writes the log buffer to S3 in a file
func (l *S3Logger) FlushLog() error {
	logger.Info("Flushing Logger to S3")

	// If the log buffer is empty, then return and don't do anything
	if l.logBuffer.Len() == 0 {
		logger.Info("Log buffer is empty, no file being shipped.")
		l.reset()
		return nil
	}

	// Include the prefix in the file name
	fullFileName := l.prefix + l.fileName

	// Setup s3 inputs
	upParams := s3manager.UploadInput{
		Bucket: &l.bucket,
		Key:    &fullFileName,
		Body:   l.logBuffer,
	}

	// Upload the data
	_, err := l.uploader.Upload(&upParams)

	// If no errors, reset
	if err == nil {
		l.reset()
		logger.Info("New file written to S3: ", l.fileName)
	} else {
		logger.Error("Error writing ", l.fileName, "to S3")
	}

	return err
}

// Shutdown calls the function that should be executed before the program terminates
func (l *S3Logger) Shutdown() error {
	return l.FlushLog()
}

// createBucket creates an S3 bucket to write the received logs from Logs API.
// If bucket creation is not successful or the bucket is not owned by the same user, it will return an error.
func createBucket(bucket string) error {

	// Create a new session to
	svc := s3.New(session.New())

	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}

	_, err := svc.CreateBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				return fmt.Errorf("%s, %s", s3.ErrCodeBucketAlreadyExists, aerr.Error())
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				return nil
			default:
				return fmt.Errorf(aerr.Error())
			}
		} else {
			return fmt.Errorf(err.Error())
		}
	}
	logger.Infof("S3 Bucket '%s' is created successfully", bucket)
	return nil
}

func generateFileName() string {

	// Get file name
	fName := strings.ToLower(os.Getenv("AWS_LAMBDA_FUNCTION_NAME"))
	ts := int(time.Now().UnixNano() / 1000000)
	timestampMilli := strconv.Itoa(ts)
	key := fmt.Sprintf("%s-%s-%s.log", fName, timestampMilli, uuid.New())

	return key

}
