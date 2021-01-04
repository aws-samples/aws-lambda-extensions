// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package plugins

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Struct to store Dynamodb cache confirmation
type DynamodbConfiguration struct {
	Region       string
	Table        string
	HashKey      string
	HashKeyType  string
	HashKeyValue string
	SortKey      string
	SortKeyType  string
	SortKeyValue string
}

var dynamoDbCache = make(map[string]string)

func InitDynamodbCache(DynamodbConfig []DynamodbConfiguration) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	for _, dynamodbConfig := range DynamodbConfig {
		if dynamodbConfig.HashKey != "" {
			// Create attributeValue map based on hash and sort key
			var attributeMap = map[string]*dynamodb.AttributeValue{}
			UpdateAttributeMap(attributeMap, dynamodbConfig)

			result, err := svc.GetItem(&dynamodb.GetItemInput{
				TableName: aws.String(dynamodbConfig.Table),
				Key:       attributeMap,
			})
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case dynamodb.ErrCodeProvisionedThroughputExceededException:
						println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
					case dynamodb.ErrCodeResourceNotFoundException:
						println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
					case dynamodb.ErrCodeRequestLimitExceeded:
						println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
					case dynamodb.ErrCodeInternalServerError:
						println(dynamodb.ErrCodeInternalServerError, aerr.Error())
					default:
						println(aerr.Error())
					}
				} else {
					println(err.Error())
				}
				return
			}
			if result.Item == nil {
				msg := "Could not find '" + dynamodbConfig.HashKeyValue + "'"
				println(msg)
			}

			// Convert data from Map to JSON string
			var data = make(map[string]string)
			_ = dynamodbattribute.UnmarshalMap(result.Item, &data)

			//Stores data in map like map[tableName+"-"+hashKeyValue+"-"+sortKeyValue] = Json
			var key = dynamodbConfig.Table + "-" + dynamodbConfig.HashKeyValue
			if dynamodbConfig.SortKey != "" {
				key += "-" + dynamodbConfig.SortKeyValue
			}

			// Convert map to JSON string
			jsonData, err := json.Marshal(data)
			if err != nil {
				println(err.Error())
			}

			dynamoDbCache[key] = string(jsonData)
		} else {
			println("HashKey not available so caching will not be enabled for %s", dynamodbConfig.HashKey)
		}
	}
}

// Create attributeValue based on key type and presence of sortKey definition
func UpdateAttributeMap(attributeMap map[string]*dynamodb.AttributeValue, dynamodbConfig DynamodbConfiguration) {
	GetAttributeValue(attributeMap, dynamodbConfig.HashKey, dynamodbConfig.HashKeyValue, dynamodbConfig.HashKeyType)
	if dynamodbConfig.SortKey != "" {
		GetAttributeValue(attributeMap, dynamodbConfig.SortKey, dynamodbConfig.SortKeyValue, dynamodbConfig.SortKeyType)
	}
}

// Supports attributeValue with data types "S" and "N"
func GetAttributeValue(attributeMap map[string]*dynamodb.AttributeValue, key string, value string, keyType string) {
	switch keyType {
	case "S":
		attributeMap[key] = &dynamodb.AttributeValue{S: aws.String(value)}
	case "N":
		attributeMap[key] = &dynamodb.AttributeValue{N: aws.String(value)}
	}
}

// Fetch Dynamodb cache
func GetDynamodbCache(name string) string {
	data, err := json.Marshal(dynamoDbCache[name])
	if err != nil {
		println("Error while marshalling request %s", err)
		return ""
	}

	return string(data)
}
