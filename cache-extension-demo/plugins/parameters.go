// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package plugins

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// Struct for storing parameter cache configurations
type ParameterConfiguration struct {
	Region string
	Names  []string
}

var parameterCache = make(map[string]string)

// Initialize parameter cache
func InitParameterCache(ParameterConfig ParameterConfiguration) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(ParameterConfig.Region)},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		panic(err)
	}

	// Fetch from parameter store based on name
	for _, name := range ParameterConfig.Names {
		if name != "" {
			ssmsvc := ssm.New(sess, aws.NewConfig().WithRegion(ParameterConfig.Region))
			param, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
				Name:           aws.String(name),
				WithDecryption: aws.Bool(true),
			})
			if err != nil {
				println("Error while fetching parameter %s - %s\n", name, err)
				continue
			}

			value := *param.Parameter.Value
			parameterCache[name] = value
		} else {
			println("Parameter names is invalid, so skipping the entry")
		}
	}
}

// Fetch Parameter cache
func GetParameterCache(name string) string {
	return parameterCache[name]
}
