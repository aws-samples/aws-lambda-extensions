// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package extension

import (
	"aws-lambda-extensions/cache-extension-demo/plugins"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const (
	Parameters            = "parameters"
	Dynamodb              = "dynamodb"
	FileName              = "/var/task/config.yaml"
	CacheTimeOutInMinutes = "CACHE_TIMEOUT_IN_MINUTES"
)

// Struct for storing CacheConfiguration
type CacheConfig struct {
	Parameters plugins.ParameterConfiguration
	Dynamodb   []plugins.DynamodbConfiguration
}

var cacheConfig = CacheConfig{}

func InitCacheExtensions() {
	// Read the cache config file
	data := LoadConfigFile()

	// Unmarshal the configuration to struct
	err := yaml.Unmarshal([]byte(data), &cacheConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Initialize Cache
	InitCache()
	println("Cache successfully loaded")

	// Refresh cache is required via environment variable
	timeOut := os.Getenv(CacheTimeOutInMinutes)
	if timeOut != "" {
		go RefreshCache(timeOut)
	}
}

func RefreshCache(timeOut string) {
	timeOutInMinutes, err := time.ParseDuration(timeOut)
	if err != nil {
		panic("Error while converting CACHE_TIMEOUT_IN_MINUTES env variable " + timeOut)
	}

	for {
		select {
		case <-time.After(timeOutInMinutes):
			InitCache()
			println("Cache reload complete")
		}
	}
}

func InitCache() {
	plugins.InitParameterCache(cacheConfig.Parameters)
	plugins.InitDynamodbCache(cacheConfig.Dynamodb)
}

// Route request to corresponding cache handlers
func RouteCache(cacheType string, name string) string {
	switch cacheType {
	case Parameters:
		return plugins.GetParameterCache(name)
	case Dynamodb:
		return plugins.GetDynamodbCache(name)
	default:
		return ""
	}
}

// Load the config file
func LoadConfigFile() string {
	data, err := ioutil.ReadFile(FileName)
	if err != nil {
		panic(err)
	}

	return string(data)
}
