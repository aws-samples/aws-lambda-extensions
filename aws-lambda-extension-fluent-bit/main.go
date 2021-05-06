package main

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"

	"os"
	"os/signal"
	"path"
	"syscall"

	extension "aws-lambda-extensions/aws-lambda-extensions-fluent-bit/extensions"
	"aws-lambda-extensions/aws-lambda-extensions-fluent-bit/logsapi"
	log "github.com/sirupsen/logrus"
)

const (
	defaultListenerPort = "1234" // DefaultListenerPort sets the http port to 1234 if not provided by function env variable "HOST"
	maxItems            = 10000
	maxBytes            = 262144
	timeOut             = 1000
	fluentBitPath       = "/opt/extensions/fluent-bit"
)

// Subscribes Lambda Extension to Logs API
func subscribeExtensionToLogsAPI(logsApiClient *logsapi.Client, agentID string, listenerPort string) (*logsapi.SubscribeResponse, error) {
	eventTypes := []logsapi.EventType{logsapi.Platform, logsapi.Function}
	bufferingCfg := logsapi.BufferingCfg{
		MaxItems:  maxItems,
		MaxBytes:  maxBytes,
		TimeoutMS: timeOut,
	}

	destination := logsapi.Destination{
		Protocol:   logsapi.HttpProto,
		URI:        logsapi.URI(fmt.Sprintf("http://sandbox:%s", listenerPort)),
		HttpMethod: logsapi.HttpPost,
		Encoding:   logsapi.JSON,
	}

	resp, err := logsApiClient.Subscribe(eventTypes, bufferingCfg, destination, agentID)

	return resp, err
}

func main() {
	extensionName := path.Base(os.Args[0])
	printPrefix := fmt.Sprintf("[%s]", extensionName)
	logger := log.WithFields(log.Fields{"agent": extensionName})

	// Determines listener port, use default 1234 if not provided as Lambda environment variable
	listenerPort := os.Getenv("PORT")
	if os.Getenv("PORT") == "" {
		listenerPort = defaultListenerPort
	}

	// Resolves host address "sandbox" for Fluent Bit
	addr, DNSErr := net.LookupHost("sandbox")
	if DNSErr != nil {
		logger.Fatal("The DNS name sandbox cannot be resolved.")
	}
	host := addr[0]
	err := os.Setenv("HOST", host)
	if err != nil {
		logger.Fatal("Cannot set the host environment variable.")
	}

	// Optional log level for fluent bit debugging purposes
	flbErr := os.Setenv("FLB_LOG_LEVEL", "debug")
	if flbErr != nil {
		logger.Warn("Cannot set the environment variable for fluent bit debugging.")
	}

	lambdaRuntimeAPI, ok := os.LookupEnv("AWS_LAMBDA_RUNTIME_API")
	if !ok {
		logger.Fatal("AWS_LAMBDA_RUNTIME_API is not set")
	}

	// Create Lambda Extension
	extensionClient := extension.NewClient(lambdaRuntimeAPI)
	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-signals
		cancel()
		logger.Info(printPrefix, "Received", s)
		logger.Info(printPrefix, "Exiting")
	}()

	// Register Extension as soon as possible
	_, registerErr := extensionClient.Register(ctx, extensionName)
	agentID := extensionClient.ExtensionID
	if registerErr != nil {
		logger.Fatal(registerErr)
	} else {
		logger.Info("Extension registered successfully.")
		logger.Infof("Extension ID: %s", agentID)
	}

	// Find Absolute paths for fluent-bit and the .conf files
	fluentBitBin, fluentBitPathErr := filepath.Abs(fluentBitPath + "/fluent-bit")
	if fluentBitPathErr != nil {
		logger.Fatalf("File Not Found.: %s", fluentBitPathErr)
	}

	conf, confPathErr := filepath.Abs(fluentBitPath + "/fluent-bit.conf")
	if confPathErr != nil {
		logger.Fatalf("File Not Found.: %s", confPathErr)
	}

	// Start Fluent Bit Binary
	cmdFluentBit := &exec.Cmd{
		Path:   fluentBitBin,
		Args:   []string{fluentBitBin, "-c", conf},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	if cmdErr := cmdFluentBit.Start(); cmdErr != nil {
		logger.Fatalf("Fluent Bit could not be started.: %s", cmdErr)
	}

	// Create Logs API Client
	logsApiBaseUrl := fmt.Sprintf("http://%s", lambdaRuntimeAPI)
	logsApiClient, err := logsapi.NewClient(logsApiBaseUrl)
	if err != nil {
		logger.Fatal(err)
	} else {
		logger.Info("Logs API Client created.")
	}

	// Subscribe to logs API
	_, err = subscribeExtensionToLogsAPI(logsApiClient, agentID, listenerPort)
	if err != nil {
		logger.Fatal(err)
	} else {
		logger.Info("Extension " + agentID + " subscribed to Logs API.")
	}

	// Will block until invoke or shutdown event is received or cancelled via the context.
	for {
		select {
		case <-ctx.Done():
			return
		default:
			logger.Info(printPrefix, " Waiting for event...")
			// This is a blocking call
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				logger.Info(printPrefix, " Error:", err)
				logger.Info(printPrefix, " Exiting")
				return
			}
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				logger.Info(printPrefix, " Received SHUTDOWN event")
				sigErr := cmdFluentBit.Process.Signal(syscall.SIGTERM)
				if sigErr != nil {
					logger.Warn("Cannot send SIGTERM to Fluent Bit.")
				}
				logger.Info(printPrefix, " Exiting")
				return
			}
		}
	}
}
