// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
)

func searchForFilesAndUploadAndDelete(ctx context.Context, s3bucket string, rootDirectory string, substringToMatch string, creds credentials.Value) {
	for {
		time.Sleep(30 * time.Second)

		filesToUpload, err := getFilesWithSubstring(rootDirectory, substringToMatch)
		if err != nil {
			println(printPrefix, "Cannot list files", err.Error())
			res, _ := extensionClient.ExitError(ctx, err.Error())
			println(printPrefix, "Cannot report ExitError", prettyPrint(res))
		}

		if len(filesToUpload) > 0 {
			println(printPrefix, "Found", len(filesToUpload), "files to upload")
			svc, err := createS3Client(creds)
			if err != nil {
				println(printPrefix, "Cannot create an S3 client", err.Error())
				res, _ := extensionClient.ExitError(ctx, err.Error())
				println(printPrefix, "Cannot report ExitError", prettyPrint(res))
			}

			count := 0
			for _, file := range filesToUpload {
				err := uploadFile(svc, s3bucket, file, file)
				if err != nil {
					println(printPrefix, "Cannot upload file", err.Error())
					res, _ := extensionClient.ExitError(ctx, err.Error())
					println(printPrefix, "Cannot report ExitError", prettyPrint(res))
				} else {
					count++
				}
				err = os.Remove(file)
				if err != nil {
					println(printPrefix, "Cannot remove uploaded file", err.Error())
					res, _ := extensionClient.ExitError(ctx, err.Error())
					println(printPrefix, "Cannot report ExitError", prettyPrint(res))
				}
			}
		}
	}
}

func uploadFile(svc *s3.S3, s3bucket string, key string, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	println(printPrefix, "Uploading filename", filename, "with size", size)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(s3bucket),
		Key:                aws.String(key),
		ACL:                aws.String("private"),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(size),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
	})

	println(printPrefix, "uploaded to s3:", s3bucket, key)

	return err
}
