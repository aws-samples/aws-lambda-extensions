// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"aws-lambda-extensions/go-example-crash-uploader-extension/extension"
)

var (
	extensionName        = filepath.Base(os.Args[0]) // extension name has to match the filename
	extensionClient      = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	printPrefix          = fmt.Sprintf("[%s]", extensionName)
	directoryToSearch    = "/tmp"
	substringToSearchFor = "dump.upload"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		println(printPrefix, "Received", s)
		println(printPrefix, "Exiting")
	}()

	res, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}
	println(printPrefix, "Register response:", prettyPrint(res))

	// Get the required credentials
	creds := credentials.NewEnvCredentials()
	credsValue, err := creds.Get()
	if err != nil {
		extensionClient.InitError(ctx, err.Error())
	}

	// Get the name of the S3 bucket
	bucket, bucketFound := os.LookupEnv("BUCKET")
	if !bucketFound {
		extensionClient.InitError(ctx, errors.New("BUCKET_NOT_FOUND").Error())
	}

	go searchForFilesAndUploadAndDelete(ctx, bucket, directoryToSearch, substringToSearchFor, credsValue)

	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx)
}

func processEvents(ctx context.Context) {
	var requestID string
	for {
		select {
		case <-ctx.Done():
			return
		default:
			println(printPrefix, "Waiting for event...")
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				println(printPrefix, "Error:", err.Error())
				println(printPrefix, "Exiting")
				return
			}
			println(printPrefix, "Received event:", prettyPrint(res))
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				println(printPrefix, "Received SHUTDOWN event")
				numFiles, err := renameFilesWithSubstring("/tmp", "core", fmt.Sprintf("dump.upload.%s", requestID))
				if err != nil {
					extensionClient.InitError(ctx, err.Error())
				}
				println(printPrefix, "Renamed", numFiles, "files")
				println(printPrefix, "Exiting")
				return
			} else if res.EventType == extension.Invoke {
				requestID = res.RequestID
			}
		}
	}
}

func prettyPrint(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return ""
	}
	return string(data)
}
