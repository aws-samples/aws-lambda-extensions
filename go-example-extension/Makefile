build:
	GOOS=linux GOARCH=amd64 go build -v -ldflags '-d -s -w' -a -tags netgo -installsuffix netgo -o bin/extensions/go-example-extension main.go

build-GoExampleExtensionLayer:
	GOOS=linux GOARCH=amd64 go build -v -ldflags '-d -s -w' -a -tags netgo -installsuffix netgo -o $(ARTIFACTS_DIR)/extensions/go-example-extension main.go
	chmod +x $(ARTIFACTS_DIR)/extensions/go-example-extension

run-GoExampleExtensionLayer:
	go run go-example-extension/main.go
