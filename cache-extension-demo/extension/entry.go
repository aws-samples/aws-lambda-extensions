// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package extension

import (
	"aws-lambda-extensions/cache-extension-demo/plugins"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

// Constants definition
const (
	Parameters = "parameters"
	Dynamodb   = "dynamodb"
	FileName                 = "/var/task/config.yaml"
	InitializeCacheOnStartup = "CACHE_EXTENSION_INIT_STARTUP"
)

// Struct for storing CacheConfiguration
type CacheConfig struct {
	Parameters []plugins.ParameterConfiguration
	Dynamodb   []plugins.DynamodbConfiguration
}

var cacheConfig = CacheConfig{}

// Initialize cache and start the background process to refresh cache
func InitCacheExtensions() {
	// Read the cache config file
	data := LoadConfigFile()

	// Unmarshal the configuration to struct
	err := yaml.Unmarshal([]byte(data), &cacheConfig)
	if err != nil {
		log.Fatalf(plugins.PrintPrefix, "error: %v", err)
	}

	// Initialize Cache
	InitCache()
	println(plugins.PrintPrefix, "Cache successfully loaded")
}

// Initialize individual cache
func InitCache() {

	// Read Lambda env variable
	var initCache = os.Getenv(InitializeCacheOnStartup)
	var initCacheInBool = false
	if initCache != "" {
		cacheInBool, err := strconv.ParseBool(initCache)
		if err != nil {
			panic(plugins.PrintPrefix + "Error while converting CACHE_EXTENSION_INIT_STARTUP env variable " +
				initCache)
		} else {
			initCacheInBool = cacheInBool
		}
	}

	// Initialize map and load data from individual services if "CACHE_EXTENSION_INIT_STARTUP" = true
	plugins.InitParameters(cacheConfig.Parameters, initCacheInBool)
	plugins.InitDynamodb(cacheConfig.Dynamodb, initCacheInBool)
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
