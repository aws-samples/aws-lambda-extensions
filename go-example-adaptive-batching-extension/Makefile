build:
	GOOS=linux GOARCH=amd64 go build -o bin/extensions/go-example-adaptive-batching-extension main.go

build-GoExampleAdaptiveBatchingExtensionLayer:
	GOOS=linux GOARCH=amd64 go build -o $(ARTIFACTS_DIR)/extensions/go-example-adaptive-batching-extension main.go
	chmod +x $(ARTIFACTS_DIR)/extensions/go-example-adaptive-batching-extension

run-GoExampleAdaptiveBatchingExtensionLayer:
	go run go-example-adaptive-batching-extension/main.go
