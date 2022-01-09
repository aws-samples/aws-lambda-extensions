build:
	GOOS=linux GOARCH=amd64 go build -o bin/extensions/kinesis-stream-logs-extension-demo main.go

build-KinesisStreamLogsApiExtensionLayer:
	GOOS=linux GOARCH=amd64 go build -o $(ARTIFACTS_DIR)/extensions/kinesis-stream-logs-extension-demo main.go
	chmod +x $(ARTIFACTS_DIR)/extensions/kinesis-stream-logs-extension-demo

run-KinesisStreamLogsApiExtensionLayer:
	go run kinesis-stream-logs-extension-demo/main.go
