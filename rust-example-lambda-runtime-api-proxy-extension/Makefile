# 
# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0
#
#
#-- config vars
ZIP_NAME = layer-lrap.zip
LAYER_NAME = lrap
LAMBDA_LAYER_MARKER := .lambda-layer
CARGO_FEATURES := 

#-- current-condition vars
# Check if Docker is available or running-- needed by `cargo cross`.
#    modify if not cross-compiling or if using different tooling
DOCKER_RUNNING := $(shell docker ps > /dev/null 2>&1 && echo -n yes)
RS_FILES := $(shell find src -name "*.rs")


.phony: build clean cargo zip clean-build clean-cargo deploy-layer doc

# * Build both x86_64 and aarch64 binaries
# * create a Layer '.zip'
# * use AWS CLI to publish Lambda layer
#
default: deploy-layer

clean: clean-build clean-cargo

clean-build: 
	@rm -rf build
	@rm -f .lambda-layer

clean-cargo:
	@cargo clean

build/lrap_x86_64: $(RS_FILES) Cargo.toml
	@echo Building Rust application for x86_64
	@mkdir -p build
	@cross build --release --target x86_64-unknown-linux-gnu ${CARGO_FEATURES}
	@cp target/x86_64-unknown-linux-gnu/release/aws-lambda-runtime-api-proxy-rs build/lrap_x86_64

build/lrap_aarch64: $(RS_FILES) Cargo.toml
	@echo Building Rust application for aarch64
	@mkdir -p build
	@cross build --release --target aarch64-unknown-linux-gnu ${CARGO_FEATURES}
	@cp target/aarch64-unknown-linux-gnu/release/aws-lambda-runtime-api-proxy-rs build/lrap_aarch64



build/$(ZIP_NAME): build/lrap_x86_64 build/lrap_aarch64 opt/entrypoint opt/wrapper
	@rm -f build/layer-lrap.zip
	@mkdir -p build/extensions
	@cp opt/entrypoint build/extensions/lrap
	@cp opt/wrapper build/wrapper
	@cd build && zip -r $(ZIP_NAME) lrap_* wrapper extensions


zip: build/$(ZIP_NAME)


$(LAMBDA_LAYER_MARKER): build/$(ZIP_NAME)
	@echo "Publishing Lambda Extension to layer \"$(LAYER_NAME)\""
	@aws lambda publish-layer-version --layer-name $(LAYER_NAME) --zip-file fileb://build/$(ZIP_NAME) \
		--description "Layer to intercept and sanitize Lambda input and output data. Compatible with all runtimes" \
		--compatible-architectures x86_64 arm64
	@touch $(LAMBDA_LAYER_MARKER)

# 
deploy-layer: $(LAMBDA_LAYER_MARKER)

doc: 
	@cargo doc
	@echo
	@echo "Docs are located in target/doc/aws_lambda_runtime_api_proxy_rs/index.html"
