// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package plugins

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/private/util"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// Struct for storing parameter cache configurations
type ParameterConfiguration struct {
	Region string
	Names  []string
}

// Struct for caching the information
type Parameter struct {
	CacheData CacheData
	Region    string
}

var parameterCache = make(map[string]Parameter)
var regionCache = make(map[string]*ssm.SSM)

// Initialize map and cache objects (if requested)
func InitParameters(parameters []ParameterConfiguration, initializeCache bool) {
	for _, config := range parameters {
		for _, parameter := range config.Names {
			_, isParameterPresent := parameterCache[parameter]
			if !isParameterPresent {
				if initializeCache {
					// Read from SSM and add it to the cache
					GetParameter(parameter, config.Region, GetSsmClient(config.Region))
				} else {
					parameterCache[parameter] = Parameter{
						CacheData: CacheData{},
						Region:    config.Region,
					}
				}
			} else {
				println(PrintPrefix, parameter+" already exists so skipping it")
			}
		}
	}
}

// Initialize parameter cache
func GetParameter(name string, region string, ssmsvc *ssm.SSM) string {
	param, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		println(PrintPrefix, "Error while fetching parameter ", name, util.PrettyPrint(err))
		return ""
	} else {

		// Read Parameter value from SSM and update cache
		var value = *param.Parameter.Value
		parameterCache[name] = Parameter{
			CacheData: CacheData{
				Data:        value,
				CacheExpiry: GetCacheExpiry(),
			},
			Region: region,
		}

		return value
	}
}

// Get SSM Client and cache it based on region
func GetSsmClient(region string) *ssm.SSM {
	ssmClient, isCachePresent := regionCache[region]
	if !isCachePresent {
		sess, err := session.NewSessionWithOptions(session.Options{
			Config:            aws.Config{Region: aws.String(region)},
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			panic(err)
		}
		ssmClient = ssm.New(sess, aws.NewConfig().WithRegion(region))
		regionCache[region] = ssmClient
	}

	return ssmClient
}

// Fetch Parameter cache
func GetParameterCache(name string) string {
	var parameter = parameterCache[name]

	// If expired or not available in cache then read it from SSM, else return from cache
	if parameter.CacheData.Data == "" || IsExpired(parameter.CacheData.CacheExpiry) {
		return GetParameter(name, parameter.Region, GetSsmClient(parameter.Region))
	} else {
		return parameterCache[name].CacheData.Data
	}
}
