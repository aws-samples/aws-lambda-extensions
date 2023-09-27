BASENAME := $(shell basename $(CURDIR))
ARTIFACTS_DIR ?= out
targetArch := amd64

all: build-GolangExampleRuntimeApiProxyExtensionLayer

.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf out/*

build-GolangExampleRuntimeApiProxyExtensionLayer:
	echo "Building Extension layer for $(targetArch)"
	rm -rf $(ARTIFACTS_DIR)/*
	mkdir -p $(ARTIFACTS_DIR)
	echo "Starting $(targetArch) build"
	GOOS=linux GOARCH=$(targetArch) go build -o $(ARTIFACTS_DIR)/extensions/golang-example-lambda-runtime-api-proxy-extension src/main.go
	cp -R wrapper-script.sh $(ARTIFACTS_DIR)
	chmod +x $(ARTIFACTS_DIR)/extensions/golang-example-lambda-runtime-api-proxy-extension
	chmod +x $(ARTIFACTS_DIR)/wrapper-script.sh
	zip -r $(ARTIFACTS_DIR)/extension.zip $(ARTIFACTS_DIR)/wrapper-script.sh
	zip -r $(ARTIFACTS_DIR)/extension.zip $(ARTIFACTS_DIR)/extensions
