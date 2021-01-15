package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Lambda environment variable for defining TTL
const (
	CacheTimeOut = "CACHE_EXTENSION_TTL"
)

var (
	ExtensionName = filepath.Base(os.Args[0]) // extension name has to match the filename
	PrintPrefix   = fmt.Sprintf("[%s] ", ExtensionName)
)

// Struct for storing cache data with expiry timestamp [time.Now() + CACHE_EXTENSION_TTL]
type CacheData struct {
	Data        string
	CacheExpiry time.Time
}

// Check whether cache has expired
func IsExpired(cacheExpiry time.Time) bool {
	return cacheExpiry.Before(time.Now())
}

// Return cache expiry timestamp based on "time.Now() + CACHE_EXTENSION_TTL"
func GetCacheExpiry() time.Time {
	// Refresh cache is required via environment variable
	timeOut := os.Getenv(CacheTimeOut)
	if timeOut == "" {
		timeOut = "60m"
	}

	timeOutInMinutes, err := time.ParseDuration(timeOut)
	if err != nil {
		panic("Error while converting CACHE_EXTENSION_TTL env variable " + timeOut)
	}

	return time.Now().Add(timeOutInMinutes)
}

// Method for pretty printing objects in logs
func PrettyPrint(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return ""
	}
	return string(data)
}
