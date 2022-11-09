// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package telemetryApi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"net/http"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const lambdaAgentIdentifierHeaderKey string = "Lambda-Extension-Identifier"

var l = log.WithFields(log.Fields{"pkg": "telemetryApi"})

// The client used for subscribing to the Telemetry API
type Client struct {
	httpClient *http.Client
	baseUrl    string
}

func NewClient() *Client {
	baseUrl := fmt.Sprintf("http://%s/2022-07-01/telemetry", os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	return &Client{
		httpClient: &http.Client{},
		baseUrl:    baseUrl,
	}
}

// Represents the type of log events in Lambda
type EventType string

const (
	// Used to receive log events emitted by the platform
	Platform EventType = "platform"
	// Used to receive log events emitted by the function
	Function EventType = "function"
	// Used is to receive log events emitted by the extension
	Extension EventType = "extension"
)

// Configuration for receiving telemetry from the Telemetry API.
// Telemetry will be sent to your listener when one of the conditions below is met.
type BufferingCfg struct {
	// Maximum number of log events to be buffered in memory. (default: 10000, minimum: 1000, maximum: 10000)
	MaxItems uint32 `json:"maxItems"`
	// Maximum size in bytes of the log events to be buffered in memory. (default: 262144, minimum: 262144, maximum: 1048576)
	MaxBytes uint32 `json:"maxBytes"`
	// Maximum time (in milliseconds) for a batch to be buffered. (default: 1000, minimum: 100, maximum: 30000)
	TimeoutMS uint32 `json:"timeoutMs"`
}

// URI is used to set the endpoint where the logs will be sent to
type URI string

// HttpMethod represents the HTTP method used to receive logs from Logs API
type HttpMethod string

const (
	// Receive log events via POST requests to the listener
	HttpPost HttpMethod = "POST"
	// Receive log events via PUT requests to the listener
	HttpPut HttpMethod = "PUT"
)

// Used to specify the protocol when subscribing to Telemetry API for HTTP
type HttpProtocol string

const (
	HttpProto HttpProtocol = "HTTP"
)

// Denotes what the content is encoded in
type HttpEncoding string

const (
	JSON HttpEncoding = "JSON"
)

// Configuration for listeners that would like to receive telemetry via HTTP
type Destination struct {
	Protocol   HttpProtocol `json:"protocol"`
	URI        URI          `json:"URI"`
	HttpMethod HttpMethod   `json:"method"`
	Encoding   HttpEncoding `json:"encoding"`
}

type SchemaVersion string

const (
	SchemaVersion20220701 = "2022-07-01"
	SchemaVersionLatest   = SchemaVersion20220701
)

// Request body that is sent to the Telemetry API on subscribe
type SubscribeRequest struct {
	SchemaVersion SchemaVersion `json:"schemaVersion"`
	EventTypes    []EventType   `json:"types"`
	BufferingCfg  BufferingCfg  `json:"buffering"`
	Destination   Destination   `json:"destination"`
}

// Response body that is received from the Telemetry API on subscribe
type SubscribeResponse struct {
	body string
}

// Subscribes to the Telemetry API to start receiving the log events
func (c *Client) Subscribe(ctx context.Context, extensionId string, listenerUri string) (*SubscribeResponse, error) {
	eventTypes := []EventType{
		Platform,
		// Function,
		// Extension,
	}

	bufferingConfig := BufferingCfg{
		MaxItems:  1000,
		MaxBytes:  256 * 1024,
		TimeoutMS: 1000,
	}

	destination := Destination{
		Protocol:   HttpProto,
		HttpMethod: HttpPost,
		Encoding:   JSON,
		URI:        URI(listenerUri),
	}

	data, err := json.Marshal(
		&SubscribeRequest{
			SchemaVersion: SchemaVersionLatest,
			EventTypes:    eventTypes,
			BufferingCfg:  bufferingConfig,
			Destination:   destination,
		})

	if err != nil {
		return nil, errors.WithMessage(err, "Failed to marshal SubscribeRequest")
	}

	headers := make(map[string]string)
	headers[lambdaAgentIdentifierHeaderKey] = extensionId

	l.Info("[client:Subscribe] Subscribing using baseUrl:", c.baseUrl)
	resp, err := httpPutWithHeaders(ctx, c.httpClient, c.baseUrl, data, &headers)
	if err != nil {
		l.Error("[client:Subscribe] Subscription failed:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		l.Error("[client:Subscribe] Subscription failed. Logs API is not supported! Is this extension running in a local sandbox?")
	} else if resp.StatusCode != http.StatusOK {
		l.Error("[client:Subscribe] Subscription failed")
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Errorf("%s failed: %d[%s]", c.baseUrl, resp.StatusCode, resp.Status)
		}

		return nil, errors.Errorf("%s failed: %d[%s] %s", c.baseUrl, resp.StatusCode, resp.Status, string(body))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	l.Info("[client:Subscribe] Subscription success:", string(body))

	return &SubscribeResponse{string(body)}, nil
}

func httpPutWithHeaders(ctx context.Context, client *http.Client, url string, data []byte, headers *map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(data))
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
