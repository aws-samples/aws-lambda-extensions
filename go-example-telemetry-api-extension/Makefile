# ----- Target used with SAM (Serverless Application Model) -----
build-GoExampleTelemetryApiExtensionLayer:
	echo :build-GoExampleTelemetryApiExtensionLayer
	GOOS=linux GOARCH=amd64 go build -o $(ARTIFACTS_DIR)/extensions/go-example-telemetry-api-extension main.go
	chmod +x $(ARTIFACTS_DIR)/extensions/go-example-telemetry-api-extension

# ----- Target illustrating manual steps required to create the extension layer -----
buildAndDeployExtensionLayer:
	echo :buildAndDeployExtensionLayer
	rm -rf bin 
	GOOS=linux GOARCH=amd64 go build -o bin/extensions/go-example-telemetry-api-extension main.go 
	chmod +x bin/extensions/go-example-telemetry-api-extension
	cd bin && zip -r extension.zip extensions

	aws lambda publish-layer-version \
		--layer-name "go-example-telemetry-api-extension-layer" \
		--zip-file "fileb://bin/extension.zip"
