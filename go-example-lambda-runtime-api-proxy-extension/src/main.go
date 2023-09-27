// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"LAMBDA-RUNTIME-API-PROXY-EXTENSION-MAIN/golang-example-lambda-runtime-api-proxy-example/src/extension"
	"LAMBDA-RUNTIME-API-PROXY-EXTENSION-MAIN/golang-example-lambda-runtime-api-proxy-example/src/proxy"
)

const (
	printPrefix     = "[LRAP:Main]"
)

func main() {
	println(printPrefix, "Starting")
	runtimeApiEndpoint := getRuntimeApiEndpoint()
	listenerPort := getListenerPort()
	extensionName := filepath.Base(os.Args[0]) // extension name has to match the filename

	proxy.StartProxy(runtimeApiEndpoint, listenerPort)
	extensionClient := extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		println(printPrefix, "Received", s)
		println(printPrefix, "Exiting")
	}()

	println(printPrefix, "Registering extension")
	res, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		println(printPrefix, "Error registering extension")
		panic(err)
	}
	println(printPrefix, "Register response:", fmt.Sprintf("%+v",res))
	extensionClient.NextEvent(ctx)

	//processEvents(ctx)
	<-ctx.Done()
	println(printPrefix, "FINISHED")
}

func getListenerPort() int {
	port := os.Getenv("LRAP_LISTENER_PORT")
	portInt, err := strconv.Atoi(port)
	if (err != nil || portInt == 0) {
		portInt = 9009
	}
	return portInt
}

func getRuntimeApiEndpoint() string {
	endpoint:= os.Getenv("LRAP_RUNTIME_API_ENDPOINT")
	if (endpoint == "") {
		endpoint = os.Getenv("AWS_LAMBDA_RUNTIME_API")
	}
	return endpoint
}