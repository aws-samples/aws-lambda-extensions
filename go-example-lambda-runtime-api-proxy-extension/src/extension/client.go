// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

// Read about Lambda Runtime API here
// https://docs.aws.amazon.com/lambda/latest/dg/runtimes-api.html

package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// RegisterResponse is the body of the response for /register
type RegisterResponse struct {
	FunctionName    string `json:"functionName"`
	FunctionVersion string `json:"functionVersion"`
	Handler         string `json:"handler"`
}

// NextEventResponse is the response for /event/next
type NextEventResponse struct {
	EventType          EventType `json:"eventType"`
	DeadlineMs         int64     `json:"deadlineMs"`
	RequestID          string    `json:"requestId"`
	InvokedFunctionArn string    `json:"invokedFunctionArn"`
	Tracing            Tracing   `json:"tracing"`
}

// Tracing is part of the response for /event/next
type Tracing struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// StatusResponse is the body of the response for /init/error and /exit/error
type StatusResponse struct {
	Status string `json:"status"`
}

// EventType represents the type of events received from /event/next
type EventType string

const (
	// Invoke is a lambda invoke
	Invoke EventType = "INVOKE"

	// Shutdown is a shutdown event for the environment
	Shutdown EventType = "SHUTDOWN"
	printPrefix string = "[LRAP:ExtensionsApiClient]"
	extensionNameHeader      = "Lambda-Extension-Name"
	extensionIdentifierHeader = "Lambda-Extension-Identifier"
	extensionErrorType       = "Lambda-Extension-Function-Error-Type"
)

// Client is a simple client for the Lambda Extensions API
type Client struct {
	baseURL     string
	httpClient  *http.Client
	extensionID string
}

// NewClient returns a Lambda Extensions API client
func NewClient(awsLambdaRuntimeAPI string) *Client {
	println(printPrefix, "Creating extension client")
	baseURL := fmt.Sprintf("http://%s/2020-01-01/extension", awsLambdaRuntimeAPI)
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// Register will register the extension with the Extensions API
func (e *Client) Register(ctx context.Context, filename string) (*RegisterResponse, error) {
	println(printPrefix, "register endpoint=", filename)
	const action = "/register"

	url := e.baseURL + action

	// We only register for Shutdown events as the proxy
	reqBody, err := json.Marshal(map[string]interface{}{
		"events": []EventType{}, // You can register for INVOKE and SHUTDOWN events here
	})
	if err != nil {
		println(printPrefix, "failed to create request body:", err)
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		println(printPrefix, "failed to create http request:", err)
		return nil, err
	}
	httpReq.Header.Set(extensionNameHeader, filename)
	httpRes, err := e.httpClient.Do(httpReq)
	if err != nil {
		println(printPrefix, "failed to send request:", err)
		return nil, err
	}
	if httpRes.StatusCode != 200 {
		println(printPrefix, "request failed with status", httpRes.Status)
		return nil, fmt.Errorf("request failed with status %s", httpRes.Status)
	}
	defer httpRes.Body.Close()
	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		println(printPrefix, "failed to read response body:", err)
		return nil, err
	}
	res := RegisterResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
	println(printPrefix, "failed to unmarshal response body:", err)
		return nil, err
	}
	e.extensionID = httpRes.Header.Get(extensionIdentifierHeader)
	println(printPrefix, "register success, extensionID=", e.extensionID)
	return &res, nil
}

// NextEvent blocks while long polling for the next lambda invoke or shutdown
func (e *Client) NextEvent(ctx context.Context) (*NextEventResponse, error) {
	println(printPrefix, "awaiting next event")
	const action = "/event/next"
	url := e.baseURL + action

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		println(printPrefix, "failed to create http request:", err)
		return nil, err
	}
	httpReq.Header.Set(extensionIdentifierHeader, e.extensionID)
	httpRes, err := e.httpClient.Do(httpReq)
	if err != nil {
		println(printPrefix, "failed to send request:", err)
		return nil, err
	}
	if httpRes.StatusCode != 200 {
		println(printPrefix, "get request failed with status", httpRes.Status)
		return nil, fmt.Errorf("request failed with status %s", httpRes.Status)
	}
	defer httpRes.Body.Close()
	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		println(printPrefix, "failed to read response body:", err)
		return nil, err
	}
	res := NextEventResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		println(printPrefix, "failed to unmarshal response body:", err)
		return nil, err
	}
	println(printPrefix, "Next success")
	return &res, nil
}