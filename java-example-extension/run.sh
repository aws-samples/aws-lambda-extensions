#!/bin/bash
# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

EXTENSION_NAME=$1
LAMBDA_FUNCTION=$2

# Build artifacts
mvn clean install

# Create zip
chmod +x extensions/java-example-extension
zip extension.zip -j target/java-example-extension-1.0-SNAPSHOT.jar
zip extension.zip extensions/*

# Push extension
aws lambda publish-layer-version --layer-name "${EXTENSION_NAME}" --zip-file "fileb://extension.zip"

# Update Lambda function to the latest version of external pushed as Lambda layers
aws lambda update-function-configuration \
  --function-name ${LAMBDA_FUNCTION} --layers $(aws lambda list-layer-versions --layer-name ${EXTENSION_NAME} \
    --max-items 1 --no-paginate --query 'LayerVersions[0].LayerVersionArn' \
    --output text)
