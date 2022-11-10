# ----- Targets used with SAM (Serverless Application Model) -----
build-NodejsExampleTelemetryApiExtensionLayer:
	echo :build-NodejsExampleTelemetryApiExtensionLayer
	cd nodejs-example-telemetry-api-extension && npm install 
	chmod +x nodejs-example-telemetry-api-extension/index.js
	chmod +x extensions/nodejs-example-telemetry-api-extension
	cp -R extensions "$(ARTIFACTS_DIR)"
	cp -R nodejs-example-telemetry-api-extension "$(ARTIFACTS_DIR)"

# ----- Targets illustrating manual steps required to create the extension layer -----
buildAndDeployExtensionLayer:
	echo :buildAndDeployExtensionLayer
	cd nodejs-example-telemetry-api-extension && npm install 
	chmod +x nodejs-example-telemetry-api-extension/index.js
	chmod +x extensions/nodejs-example-telemetry-api-extension
	mkdir -p out
	zip -r out/extension.zip ./extensions
	zip -r out/extension.zip ./nodejs-example-telemetry-api-extension

	aws lambda publish-layer-version \
		--layer-name "nodejs-example-telemetry-api-extension" \
		--zip-file "fileb://out/extension.zip"
