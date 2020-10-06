// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"bytes"
	"errors"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func createS3Client(creds credentials.Value) (*s3.S3, error) {
	// Create session and S3 client
	s3region, s3regionFound := os.LookupEnv("AWS_REGION")
	if !s3regionFound {
		return nil, errors.New("AWS_REGION is not set")
	}

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(s3region),
			Credentials: credentials.NewStaticCredentials(
				creds.AccessKeyID,
				creds.SecretAccessKey,
				creds.SessionToken,
			),
		})
	if err != nil {
		return nil, err
	}
	svc := s3.New(sess)
	return svc, nil
}

func createMultipartUpload(svc *s3.S3, s3bucket string, filename string) (*s3.CreateMultipartUploadOutput, error) {
	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(s3bucket),
		Key:    aws.String(path.Base(filename)),
	}
	resp, err := svc.CreateMultipartUpload(input)
	return resp, err
}

func uploadPart(svc *s3.S3, resp *s3.CreateMultipartUploadOutput, fileChunk []byte, partNumber int) (*s3.CompletedPart, error) {
	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(fileChunk),
		Bucket:        resp.Bucket,
		Key:           resp.Key,
		PartNumber:    aws.Int64(int64(partNumber)),
		UploadId:      resp.UploadId,
		ContentLength: aws.Int64(int64(len(fileChunk))),
	}

	tryNum := 1
	for {
		uploadResult, err := svc.UploadPart(partInput)
		if err != nil {
			if tryNum >= 3 {
				if aerr, ok := err.(awserr.Error); ok {
					return nil, aerr
				}
				return nil, err
			}
			tryNum++
		} else {
			return &s3.CompletedPart{
				ETag:       uploadResult.ETag,
				PartNumber: aws.Int64(int64(partNumber)),
			}, nil
		}
	}
}

func abortMultipartUpload(svc *s3.S3, resp *s3.CreateMultipartUploadOutput) error {
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
	}
	_, err := svc.AbortMultipartUpload(abortInput)
	return err
}

func completeMultipartUpload(svc *s3.S3, resp *s3.CreateMultipartUploadOutput, completedParts []*s3.CompletedPart) (*s3.CompleteMultipartUploadOutput, error) {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}

	completedOutput, err := svc.CompleteMultipartUpload(completeInput)

	return completedOutput, err
}
