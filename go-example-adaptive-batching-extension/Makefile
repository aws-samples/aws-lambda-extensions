build:
	GOOS=linux GOARCH=amd64 go build -o bin/extensions/go-example-adaptive-batching main.go

build-GoExampleAdaptiveBatchingExtensionLayer:
	GOOS=linux GOARCH=amd64 go build -o $(ARTIFACTS_DIR)/extensions/go-example-adaptive-batching main.go
	chmod +x $(ARTIFACTS_DIR)/extensions/go-example-adaptive-batching

run-GoExampleAdaptiveBatchingExtensionLayer:
	go run go-example-adaptive-batching/main.go
