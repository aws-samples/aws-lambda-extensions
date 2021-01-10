// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package ipc

import (
	"aws-lambda-extensions/cache-extension-demo/extension"
	"aws-lambda-extensions/cache-extension-demo/plugins"
	"github.com/gorilla/mux"
	"net/http"
)

// Start begins running the sidecar
func Start(port string) {
	go startHTTPServer(port)
}

// Method that responds back with the cached values
func startHTTPServer(port string) {
	router := mux.NewRouter()
	router.HandleFunc("/{cacheType}/{key}",
		func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			value := extension.RouteCache(vars["cacheType"], vars["key"])

			if len(value) != 0 {
				_, _ = w.Write([]byte(value))
			} else {
				_, _ = w.Write([]byte("No data found"))
			}
		})

	println(plugins.PrintPrefix, "Starting Httpserver on port ", port)
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		panic(err)
	}
}
