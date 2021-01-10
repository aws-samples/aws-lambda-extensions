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
	Table        string
	HashKey      string
	HashKeyType  string
	HashKeyValue string
	SortKey      string
	SortKeyType  string
	SortKeyValue string
}

// Struct for caching the information
type Dynamodb struct {
	CacheData             CacheData
	DynamodbConfiguration DynamodbConfiguration
}

var dynamoDbCache = make(map[string]Dynamodb)
var dynamoDbClient = GetDynamoDbClient()

// Initialize map and cache data (only if requested)
func InitDynamodb(dynamodbConfiguration []DynamodbConfiguration, initializeCache bool) {
	for _, dynamodbConfig := range dynamodbConfiguration {
		if initializeCache {
			// Read data from Dynamodb
			GetData(dynamodbConfig)
		} else {
			dynamoDbCache[GetKey(dynamodbConfig)] = Dynamodb{
				CacheData:             CacheData{},
				DynamodbConfiguration: dynamodbConfig,
			}
		}
	}
}

// Read data from Dynamodb
func GetData(dynamodbConfig DynamodbConfiguration) string {
	if dynamodbConfig.HashKey != "" {
		// Create attributeValue map based on hash and sort key
		var attributeMap = map[string]*dynamodb.AttributeValue{}
		UpdateAttributeMap(attributeMap, dynamodbConfig)

		result, err := dynamoDbClient.GetItem(&dynamodb.GetItemInput{
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
					println(PrintPrefix, PrettyPrint(aerr.Error()))
				}
			} else {
				println(PrintPrefix, PrettyPrint(err.Error()))
			}
			return ""
		}
		if result.Item == nil {
			println(PrintPrefix, "Could not find '"+dynamodbConfig.HashKeyValue+"'")
			return ""
		}

		// Convert data from Map to JSON string
		var data = make(map[string]string)
		_ = dynamodbattribute.UnmarshalMap(result.Item, &data)

		// Convert map to JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			println(err.Error())
		}

		// Add it to the cache
		var value = string(jsonData)
		dynamoDbCache[GetKey(dynamodbConfig)] = Dynamodb{
			CacheData: CacheData{
				Data:        value,
				CacheExpiry: GetCacheExpiry(),
			},
			DynamodbConfiguration: dynamodbConfig,
		}

		return value
	} else {
		println(PrintPrefix, "HashKey not available so caching will not be enabled for %s", dynamodbConfig.HashKey)
		return ""
	}
}

// Generate key to store in map based with a format "tableName+"-"+hashKeyValue+"-"+sortKeyValue"
func GetKey(dynamodbConfig DynamodbConfiguration) string {
	var key = dynamodbConfig.Table + "-" + dynamodbConfig.HashKeyValue
	if dynamodbConfig.SortKey != "" {
		key += "-" + dynamodbConfig.SortKeyValue
	}
	return key
}

// Get Dynamodb to read data
func GetDynamoDbClient() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create Dynamodb client
	return dynamodb.New(sess)
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
	var dbCache = dynamoDbCache[name]

	// If expired or not available in cache then read it from Dynamodb, else return from cache
	if dbCache.CacheData.Data == "" || IsExpired(dbCache.CacheData.CacheExpiry) {
		return GetData(dynamoDbCache[name].DynamodbConfiguration)
	} else {
		return dynamoDbCache[name].CacheData.Data
	}
}
