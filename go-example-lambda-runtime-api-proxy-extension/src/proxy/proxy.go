// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

// Read about Lambda Runtime API here
// https://docs.aws.amazon.com/lambda/latest/dg/runtimes-api.html

package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	//"time"

	"github.com/go-chi/chi/v5"
)

const (
	printPrefix     = "[LRAP:RuntimeApiProxy]"
)

var (
	awsLambdaRuntimeAPI string
	client = &http.Client{}
)


// NewProxy returns a Lambda Extensions API client
func StartProxy(endpoint string, port int) {
	println(printPrefix, "Starting proxy server")
	awsLambdaRuntimeAPI = endpoint

	r := chi.NewRouter()
	// Lambda runtime API
	r.Use(simpleLogger)
	r.Get("/2018-06-01/runtime/invocation/next", handleNext)
	r.Post("/2018-06-01/runtime/invocation/{requestId}/response", handleResponse)
	r.Post("/2018-06-01/runtime/init/error", handleInitError)
	r.Post("/2018-06-01/runtime/invocation/{requestId}/error", handleInvokeError)

	// NotFound defines a handler to respond whenever a route could
	// not be found.
	r.NotFound(handleError)

	// MethodNotAllowed defines a handler to respond whenever a method is
	// not allowed.
	r.MethodNotAllowed(handleError)

	proxy := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        r,
		// ReadTimeout:    30 * time.Second,
		// WriteTimeout:   30 * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}

	go func ()  {
		err := proxy.ListenAndServe()
		if err != nil {
			println(printPrefix, "proxy reported error:", fmt.Sprintf("%s", err))
		}
	}()
	println(printPrefix, "Proxy Server Started")
}

func handleNext(w http.ResponseWriter, r *http.Request) {
	println(printPrefix, "Handle Next Request")

	url := fmt.Sprintf("http://%s/2018-06-01/runtime/invocation/next", awsLambdaRuntimeAPI)

	resp, err := request("GET", url, r.Body, r.Header)
	if err != nil {
		return
	}

	body, err := readBody(resp.Body)
	if err != nil {
		return
	}

	body, headers := processRequest(body, resp.Header)

	finalizeResponse(w, body, headers)
	println(printPrefix, "handleNext posted")
}

func handleResponse(w http.ResponseWriter, r *http.Request) {
	requestId := chi.URLParam(r, "requestId")
	println(printPrefix, "Handle Response for requestID:", requestId)

	body, err := readBody(r.Body)
	if err != nil {
		return
	}

	body, headers := processResponse(body, r.Header, )
	url := fmt.Sprintf("http://%s/2018-06-01/runtime/invocation/%s/response", awsLambdaRuntimeAPI, requestId)
	bodyBuffer := io.NopCloser(bytes.NewReader(body))

	proxyPost(w, headers, url, bodyBuffer)
	println(printPrefix, "handleResponse posted")
}

func handleInitError(w http.ResponseWriter, r *http.Request) {
	println(printPrefix, "Handle Init Error")

	url := fmt.Sprintf("http://%s/2018-06-01/runtime/init/error", awsLambdaRuntimeAPI)
	proxyPost(w, r.Header, url, r.Body)

	println(printPrefix, "handleInitError posted")
}

func handleInvokeError(w http.ResponseWriter, r *http.Request) {
	requestId := chi.URLParam(r, "requestId")
	println(printPrefix, "Handle Invoke Error for requestID:", requestId)

	url := fmt.Sprintf("http://%s/2018-06-01/runtime/invocation/%s/error", awsLambdaRuntimeAPI, requestId)
	proxyPost(w, r.Header, url, r.Body)

	println(printPrefix, "handleInvokeError posted")
}

func proxyPost(w http.ResponseWriter, headers http.Header, url string, body io.ReadCloser) {
	resp, err := request("POST", url, body, headers)
	if err != nil {
		return
	}

	respBody, err := readBody(resp.Body)
	if err != nil {
		return
	}

	finalizeResponse(w, respBody, resp.Header)
}

func handleError(w http.ResponseWriter, r *http.Request) {
	println(printPrefix, "Path or Protocol Error")
	http.Error(w, http.StatusText(404), 404)
}

func copyHeaders(original http.Header, target http.Header) {
	for key, value := range original {
  		target[strings.ToLower(key)]  = value
	}
}

func finalizeResponse(w http.ResponseWriter, body []byte, headers http.Header) {
	copyHeaders(headers, w.Header())
	_, err := w.Write(body)
	if err != nil {
		println(printPrefix, "Error writing response body")
		return
	}
}

func readBody(bodyBuffer io.ReadCloser) ([]byte, error) {
	defer bodyBuffer.Close()
	body, err := io.ReadAll(bodyBuffer)
	if err != nil {
		println(printPrefix, "Error reading body", err)
		return nil, err
	}
	return body, nil
}

func unmarshalBody(body []byte) (map[string]interface{}, error) {
	var temp = make(map[string]interface{})
	err := json.Unmarshal(body, &temp)
	if err != nil {
		println(printPrefix, "failed to unmarshal response body:", err)
		return nil, err
	}
	return temp, nil
}

func request(verb string, url string, body io.Reader, headers http.Header) (*http.Response, error) {
	request, err := http.NewRequest(verb, url, body)
	if err != nil {
		println(printPrefix, "Error creating http request")
		return nil, err
	}
	if (headers != nil) {
		copyHeaders(request.Header, headers)
	}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("%s Error doing http request\nHeaders: %+v\nBody: %+v\nURL: %+v\n",
			printPrefix, request.Header, body, url)
		return nil, err
	}
	return resp, nil
}

func simpleLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s proxyRequestMetadata method=%s url=%s\n", printPrefix, r.Method, r.URL)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// Assumes body is a JSON object. Expand as needed
func processRequest(body []byte, headers http.Header) ([]byte, http.Header) {
	jsonBody, err := unmarshalBody(body)
	if err != nil {
		println(printPrefix, "Error unmarshalling body, returning original body")
		return body, headers
	}

	jsonBody["LRAP RequestModified"] = true

	newBody, err :=json.Marshal(jsonBody)
	if err != nil {
		println(printPrefix, "Error marshalling body, returning original body")
		return body, headers
	}
	return newBody, headers
}

// Assumes body is a JSON object. Expand as needed
func processResponse(body []byte, headers http.Header) ([]byte, http.Header) {
	jsonBody, err := unmarshalBody(body)
	if err != nil {
		println(printPrefix, "Error unmarshalling body, returning original body")
		return body, headers
	}

	jsonBody["LRAP ResponseModified"] = true

	newBody, err :=json.Marshal(jsonBody)
	if err != nil {
		println(printPrefix, "Error marshalling body, returning original body")
		return body, headers
	}
	return newBody, headers
}