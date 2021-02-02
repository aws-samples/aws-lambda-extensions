// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"aws-lambda-extensions/cache-extension-demo/extension"
	"aws-lambda-extensions/cache-extension-demo/ipc"
	"aws-lambda-extensions/cache-extension-demo/plugins"
	"context"
	"os"
	"os/signal"
	"syscall"
)

var (
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		println(plugins.PrintPrefix, "Received", s)
		println(plugins.PrintPrefix, "Exiting")
	}()

	res, err := extensionClient.Register(ctx, plugins.ExtensionName)
	if err != nil {
		panic(err)
	}
	println(plugins.PrintPrefix, "Register response:", plugins.PrettyPrint(res))

	// Initialize all the cache plugins
	extension.InitCacheExtensions()

	// Start HTTP server
	ipc.Start("4000")

	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx)
}

// Method to process events
func processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			println(plugins.PrintPrefix, "Waiting for event...")
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				println(plugins.PrintPrefix, "Error:", err)
				println(plugins.PrintPrefix, "Exiting")
				return
			}

			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				println(plugins.PrintPrefix, "Received SHUTDOWN event")
				println(plugins.PrintPrefix, "Exiting")
				return
			}
		}
	}
}
