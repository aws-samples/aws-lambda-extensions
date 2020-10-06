// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"context"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
)

func searchForFilesAndUploadAndDelete(ctx context.Context, s3bucket string, rootDirectory string, substringToMatch string, creds credentials.Value) {
	filesToUpload, err := getFilesWithSubstring(rootDirectory, substringToMatch)
	if err != nil {
		res, _ := extensionClient.ExitError(ctx, err.Error())
		println(printPrefix, "ExitError response:", prettyPrint(res))
	}

	if len(filesToUpload) > 0 {
		println(printPrefix, "Found", len(filesToUpload), "files to upload")
		svc, err := createS3Client(creds)
		if err != nil {
			res, _ := extensionClient.ExitError(ctx, err.Error())
			println(printPrefix, "ExitError response:", prettyPrint(res))
		}

		count := 0
		for _, file := range filesToUpload {
			err := uploadFileToS3(svc, s3bucket, file)
			if err != nil {
				res, _ := extensionClient.ExitError(ctx, err.Error())
				println(printPrefix, "ExitError response:", prettyPrint(res))
			} else {
				count++
			}
			err = os.Remove(file)
			if err != nil {
				res, _ := extensionClient.ExitError(ctx, err.Error())
				println(printPrefix, "ExitError response:", prettyPrint(res))
			}
		}
	}
}

const maxPartSize int64 = 5 * 1024 * 1024 // 5MB
func uploadFileToS3(svc *s3.S3, s3bucket, filename string) error {
	println(printPrefix, "Uploading", filename)
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	resp, err := createMultipartUpload(svc, s3bucket, filename)
	if err != nil {
		return err
	}

	buffer := make([]byte, maxPartSize)
	var completedParts []*s3.CompletedPart

	for {
		_, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				_ = abortMultipartUpload(svc, resp)
				return err
			}
			break
		}
		completedPart, err := uploadPart(svc, resp, buffer, len(completedParts)+1)
		if err != nil {
			_ = abortMultipartUpload(svc, resp)
			return err
		}

		completedParts = append(completedParts, completedPart)
	}

	completeResponse, err := completeMultipartUpload(svc, resp, completedParts)
	if err != nil {
		return err
	}
	println(completeResponse.String())
	return nil
}
