build:
	GOOS=linux GOARCH=amd64 go build -o bin/extensions/go-example-extension main.go

build-GoExampleExtensionLayer:
	GOOS=linux GOARCH=amd64 go build -o $(ARTIFACTS_DIR)/extensions/go-example-extension main.go
	chmod +x $(ARTIFACTS_DIR)/extensions/go-example-extension

run-GoExampleExtensionLayer:
	go run go-example-extension/main.go
