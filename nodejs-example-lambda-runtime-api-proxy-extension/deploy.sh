#!/bin/bash

set -euxo pipefail

export FUNCTION_NAME="lambda-demo-function"
export EXTENSION_NAME="nodejs-example-lambda-runtime-api-proxy-extension"
export LAYER_NAME=$EXTENSION_NAME-layer

function packageExtension {
    echo "> packageExtension"
    rm -rf out
    mkdir -p out
	
	cd $EXTENSION_NAME
    npm install 
    cd ..
	chmod +x extensions/$EXTENSION_NAME
    chmod +x wrapper-script.sh
	zip -r out/extension.zip ./extensions
	zip -r out/extension.zip ./$EXTENSION_NAME
	zip -r out/extension.zip wrapper-script.sh
}

function publishLayerVersion {
    echo "> publishLayerVersion"
	export LAYER_VERSION_ARN=$( \
        aws lambda publish-layer-version \
		--layer-name $LAYER_NAME \
		--zip-file "fileb://out/extension.zip" \
        --output text \
        --query 'LayerVersionArn')
    echo $LAYER_VERSION_ARN
}

function updateFunctionConfiguration {
    echo "> updateFunctionConfiguration"
    aws lambda update-function-configuration \
        --function-name $FUNCTION_NAME \
        --layers $LAYER_VERSION_ARN \
        --environment 'Variables={AWS_LAMBDA_EXEC_WRAPPER=/opt/wrapper-script.sh}'
}

packageExtension
publishLayerVersion
updateFunctionConfiguration