build:
	GOOS=linux GOARCH=amd64 go build -o bin/extensions/kinesisfirehose-logs-extension-demo main.go

build-KinesisFireHoseLogsApiExtensionLayer:
	GOOS=linux GOARCH=amd64 go build -o $(ARTIFACTS_DIR)/extensions/kinesisfirehose-logs-extension-demo main.go
	chmod +x $(ARTIFACTS_DIR)/extensions/kinesisfirehose-logs-extension-demo

run-KinesisFireHoseLogsApiExtensionLayer:
	go run kinesisfirehose-logs-extension-demo/main.go
