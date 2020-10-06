// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package ipc

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// Start begins running the sidecar
func Start(port string) {
	writeToFileSystem("/tmp/test.txt", "Hello I'm a temp file")
	go startHTTPServer(port)
}

func writeToFileSystem(filename string, data string) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		panic(err)
	}
	err = file.Sync()
	if err != nil {
		panic(err)
	}
}

func startHTTPServer(port string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello from http server")
	})
	// port 8080 is used by the Lambda Invoke API
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
