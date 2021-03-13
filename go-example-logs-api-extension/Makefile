build:
	GOOS=linux GOARCH=amd64 go build -o bin/extensions/go-example-logs-api-extension main.go

build-GoExampleLogsApiExtensionLayer:
	GOOS=linux GOARCH=amd64 go build -o $(ARTIFACTS_DIR)/extensions/go-example-logs-api-extension main.go
	chmod +x $(ARTIFACTS_DIR)/extensions/go-example-logs-api-extension

run-GoExampleLogsApiExtensionLayer:
	go run go-example-logs-api-extension/main.go
