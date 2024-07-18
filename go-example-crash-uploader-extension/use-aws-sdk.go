// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"errors"
	"os"

	"github.com/aws/aws-sdk-go/aws"
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
