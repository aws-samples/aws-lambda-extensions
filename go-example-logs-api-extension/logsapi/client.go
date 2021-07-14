// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package logsapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const lambdaAgentIdentifierHeaderKey string = "Lambda-Extension-Identifier"

// Client is the client used to subscribe to the Logs API
type Client struct {
	httpClient     *http.Client
	logsApiBaseUrl string
}

// NewClient returns a new Client with the given URL
func NewClient(logsApiBaseUrl string) (*Client, error) {
	return &Client{
		httpClient:     &http.Client{},
		logsApiBaseUrl: logsApiBaseUrl,
	}, nil
}

// EventType represents the type of logs in Lambda
type EventType string

const (
	// Platform is to receive logs emitted by the platform
	Platform EventType = "platform"
	// Function is to receive logs emitted by the function
	Function EventType = "function"
	// Extension is to receive logs emitted by the extension
	Extension EventType = "extension"
)

type SubEventType string

const (
	// RuntimeDone event is sent when lambda function is finished it's execution
	RuntimeDone SubEventType = "platform.runtimeDone"
)

// BufferingCfg is the configuration set for receiving logs from Logs API. Whichever of the conditions below is met first, the logs will be sent
type BufferingCfg struct {
	// MaxItems is the maximum number of events to be buffered in memory. (default: 10000, minimum: 1000, maximum: 10000)
	MaxItems uint32 `json:"maxItems"`
	// MaxBytes is the maximum size in bytes of the logs to be buffered in memory. (default: 262144, minimum: 262144, maximum: 1048576)
	MaxBytes uint32 `json:"maxBytes"`
	// TimeoutMS is the maximum time (in milliseconds) for a batch to be buffered. (default: 1000, minimum: 100, maximum: 30000)
	TimeoutMS uint32 `json:"timeoutMs"`
}

// URI is used to set the endpoint where the logs will be sent to
type URI string

// HttpMethod represents the HTTP method used to receive logs from Logs API
type HttpMethod string

const (
	//HttpPost is to receive logs through POST.
	HttpPost HttpMethod = "POST"
	//HttpPUT is to receive logs through PUT.
	HttpPut HttpMethod = "PUT"
)

// HttpProtocol is used to specify the protocol when subscribing to Logs API for HTTP
type HttpProtocol string

const (
	HttpProto HttpProtocol = "HTTP"
)

// HttpEncoding denotes what the content is encoded in
type HttpEncoding string

const (
	JSON HttpEncoding = "JSON"
)

// Destination is the configuration for listeners who would like to receive logs with HTTP
type Destination struct {
	Protocol   HttpProtocol `json:"protocol"`
	URI        URI          `json:"URI"`
	HttpMethod HttpMethod   `json:"method"`
	Encoding   HttpEncoding `json:"encoding"`
}

type SchemaVersion string

const (
	SchemaVersion20210318 = "2021-03-18"
	SchemaVersionLatest   = SchemaVersion20210318
)

// SubscribeRequest is the request body that is sent to Logs API on subscribe
type SubscribeRequest struct {
	SchemaVersion SchemaVersion `json:"schemaVersion"`
	EventTypes    []EventType   `json:"types"`
	BufferingCfg  BufferingCfg  `json:"buffering"`
	Destination   Destination   `json:"destination"`
}

// SubscribeResponse is the response body that is received from Logs API on subscribe
type SubscribeResponse struct {
	body string
}

// Subscribe calls the Logs API to subscribe for the log events.
func (c *Client) Subscribe(types []EventType, bufferingCfg BufferingCfg, destination Destination, extensionId string) (*SubscribeResponse, error) {

	data, err := json.Marshal(
		&SubscribeRequest{
			SchemaVersion: SchemaVersionLatest,
			EventTypes:    types,
			BufferingCfg:  bufferingCfg,
			Destination:   destination,
		})
	if err != nil {
		return nil, errors.WithMessage(err, "failed to marshal SubscribeRequest")
	}

	headers := make(map[string]string)
	headers[lambdaAgentIdentifierHeaderKey] = extensionId
	url := fmt.Sprintf("%s/2020-08-15/logs", c.logsApiBaseUrl)
	resp, err := httpPutWithHeaders(c.httpClient, url, data, &headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		fmt.Println("WARNING!!! Logs API is not supported! Is this extension running in a local sandbox?")
	} else if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Errorf("%s failed: %d[%s]", url, resp.StatusCode, resp.Status)
		}

		return nil, errors.Errorf("%s failed: %d[%s] %s", url, resp.StatusCode, resp.Status, string(body))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	return &SubscribeResponse{string(body)}, nil
}

func httpPutWithHeaders(client *http.Client, url string, data []byte, headers *map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	contentType := "application/json"
	req.Header.Set("Content-Type", contentType)
	if headers != nil {
		for k, v := range *headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
