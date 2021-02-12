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

// States for the S3 Logger
type state uint64

const (
	// CREATE_BUCKET denotes the state in which the S3 bucket will be created, if not created already
	CREATE_BUCKET state = 1 + iota
	// CREATE_MULTI_PART_UPLOAD denotes the state in which the multipart uploader will be initialized
	CREATE_MULTI_PART_UPLOAD
	// PUT_LOG_PARTS denotes the state in which the buffer will be uploaded in parts to S3 with MAX_PART_SIZE sized chunks
	PUT_LOG_PARTS
)

// MultiPartsData is used to keep track of each part uploaded to S3
type MultiPartsData struct {
	completedParts []*s3.CompletedPart
	partNumber     int64
}

// NewMultiPartsData returns a new MultiPartsData
func NewMultiPartsData() *MultiPartsData {
	return &MultiPartsData{
		completedParts: make([]*s3.CompletedPart, 0),
		partNumber:     1,
	}
}

// S3Logger is the logger that writes the logs received from Logs API to S3
type S3Logger struct {
	multiPartsData *MultiPartsData
	svc            *s3.S3
	bucket         string
	key            string
	uploadId       string
	logBuffer      *bytes.Buffer
	state          state
}

// NewS3Logger returns an S3 Logger
func NewS3Logger() (*S3Logger, error) {
	fName := strings.ToLower(os.Getenv("AWS_LAMBDA_FUNCTION_NAME"))
	bucket, present := os.LookupEnv("LOGS_API_EXTENSION_S3_BUCKET")
	if !present {
		return nil, errors.New("Environment variable LOGS_API_EXTENSION_S3_BUCKET is not set.")
	} else {
		fmt.Println("Sending logs to:", bucket)
	}
	ts := int(time.Now().UnixNano() / 1000000)
	timestampMilli := strconv.Itoa(ts)
	key := fmt.Sprintf("%s-%s-%s.log", fName, timestampMilli, uuid.New())
	buffer := bytes.NewBuffer([]byte(""))
	buffer.Grow(2 * MAX_PART_SIZE)

	return &S3Logger{
		multiPartsData: NewMultiPartsData(),
		svc:            s3.New(session.New()),
		bucket:         bucket,
		key:            key,
		logBuffer:      buffer,
		state:          CREATE_BUCKET,
	}, nil
}

// PushLog writes the received logs to a buffer and takes actions depending on the current state of the logger.
func (l *S3Logger) PushLog(log string) error {
	l.logBuffer.Write([]byte(log))
L:
	for {
		switch l.state {
		case CREATE_BUCKET:
			err := l.createBucket()
			if err != nil {
				return err
			}
			l.state = CREATE_MULTI_PART_UPLOAD
			continue
		case CREATE_MULTI_PART_UPLOAD:
			err := l.createMultiPartUpload()
			if err != nil {
				return err
			}
			l.state = PUT_LOG_PARTS
			continue
		case PUT_LOG_PARTS:
			err := l.putLogParts()
			if err != nil {
				return err
			}
			break L
		}
	}
	return nil
}

// Shutdown calls the function that should be executed before the program terminates
func (l *S3Logger) Shutdown() error {
	return l.finalizeLogsAndCreateS3File()
}

// createBucket creates an S3 bucket to write the received logs from Logs API.
// If bucket creation is not successful or the bucket is not owned by the same user, it will return an error.
func (l *S3Logger) createBucket() error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(l.bucket),
	}

	_, err := l.svc.CreateBucket(input)
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
	logger.Infof("S3 Bucket '%s' is created successfully", l.bucket)
	return nil
}

// createMultiPartUpload initiates multipart upload process to S3
// After the multipart upload is initiated, parts can start to be uploaded.
func (l *S3Logger) createMultiPartUpload() error {
	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(l.bucket),
		Key:    aws.String(l.key),
	}

	resp, err := l.svc.CreateMultipartUpload(input)
	if err != nil {
		return errors.New(fmt.Sprintf("Multipart upload creation operation has failed: %v", err))
	}

	l.uploadId = *(resp.UploadId)
	return nil
}

// putLogParts uploads a MAX_PART_SIZE sized part if the logBuffer's size is at least MAX_PART_SIZE.
func (l *S3Logger) putLogParts() error {
	if l.logBuffer.Len() < MAX_PART_SIZE {
		return nil
	}

	completedPart, err := l.uploadPart(l.logBuffer.Next(MAX_PART_SIZE))
	if err != nil {
		return errors.New(fmt.Sprintf("File part [%d] could not be uploaded: %v", l.multiPartsData.partNumber, err))
	}

	logger.Infof("File part [%d] is uploaded.", l.multiPartsData.partNumber)
	l.multiPartsData.partNumber++
	l.multiPartsData.completedParts = append(l.multiPartsData.completedParts, completedPart)
	return nil
}

// putLogPartsComplete uploads all the remaining buffer in a single part and completes the multipart upload process.
func (l *S3Logger) putLogPartsComplete() error {
	completedPart, err := l.uploadPart(l.logBuffer.Bytes())
	if err != nil {
		return errors.New(fmt.Sprintf("File part [%d] could not be uploaded: %v", l.multiPartsData.partNumber, err))
	} else {
		logger.Infof("File part [%d] is uploaded.", l.multiPartsData.partNumber)
		l.multiPartsData.completedParts = append(l.multiPartsData.completedParts, completedPart)
		l.multiPartsData.partNumber++
	}

	_, err = l.completeMultipartUpload()
	if err != nil {
		return errors.New(fmt.Sprintf("Multipart upload completion operation could not be completed: %v", err))
	}
	return nil
}

func (l *S3Logger) uploadPart(buffer []byte) (*s3.CompletedPart, error) {
	singlePartInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(buffer),
		Bucket:        aws.String(l.bucket),
		Key:           aws.String(l.key),
		PartNumber:    aws.Int64(l.multiPartsData.partNumber),
		UploadId:      aws.String(l.uploadId),
		ContentLength: aws.Int64(int64(len(buffer))),
	}

	tryNumber := 0
	for tryNumber < MAX_RETRIES {
		uploadResult, err := l.svc.UploadPart(singlePartInput)
		if err != nil {
			if tryNumber == MAX_RETRIES {
				if aerr, ok := err.(awserr.Error); ok {
					return nil, aerr
				}
				return nil, err
			}
			tryNumber++
		} else {
			return &s3.CompletedPart{
				ETag:       uploadResult.ETag,
				PartNumber: aws.Int64(l.multiPartsData.partNumber),
			}, nil
		}
	}
	return nil, nil
}

func (l *S3Logger) completeMultipartUpload() (*s3.CompleteMultipartUploadOutput, error) {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(l.bucket),
		Key:      aws.String(l.key),
		UploadId: aws.String(l.uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: l.multiPartsData.completedParts,
		},
	}
	return l.svc.CompleteMultipartUpload(completeInput)
}

// finalizeLogsAndCreateS3File finalizes uploading process or aborts if completion fails.
// If completion fails, uploading will be cancelled.
// Any part that has been uploaded to S3 so far will be LOST if this function can't execute successfully.
func (l *S3Logger) finalizeLogsAndCreateS3File() error {
	if l.state != PUT_LOG_PARTS {
		return errors.New("No logs are received at the time the program terminated, not writing any files to S3.")
	}

	err := l.putLogPartsComplete()
	if err != nil {
		err2 := l.AbortMultipartUpload()
		if err2 != nil {
			return fmt.Errorf("1.%v and 2.%v", err, err2)
		}
	}
	return err
}

// AbortMultipartUpload terminates the multi part upload process and discard any parts uploaded to S3 so far.
// This function is to be called whenever complete multipart upload operation fails or before the extension terminates
// to avoid getting charged for incomplete parts uploaded to S3. As a best practice, we recommend you configure
// a lifecycle rule (using the AbortIncompleteMultipartUpload action) to minimize your storage costs.
// See https://docs.aws.amazon.com/AmazonS3/latest/dev/mpuoverview.html for more information on this.
func (l *S3Logger) AbortMultipartUpload() error {
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(l.bucket),
		Key:      aws.String(l.key),
		UploadId: aws.String(l.uploadId),
	}
	_, err := l.svc.AbortMultipartUpload(abortInput)

	if err != nil {
		return errors.New(fmt.Sprintf("Abort upload operation could not be completed. %v", err))
	}

	logger.Infof("Multipart uploading is aborted. File could not be written to S3.")
	return nil
}
