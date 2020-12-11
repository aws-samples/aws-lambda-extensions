#!/bin/bash
# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

EXTENSION_NAME=$1
LAMBDA_FUNCTION=$2

# Build artifacts
mvn clean install

# Perform cleanup & create zip
rm -Rf extensions/*.jar
rm -Rf extensions/*.zip
mv target/java-example-extension-1.0-SNAPSHOT.jar extensions
chmod +x extensions/java-example-extension
cd extensions
zip -r extension.zip .

# Push extension
aws lambda publish-layer-version --layer-name "${EXTENSION_NAME}" --zip-file "fileb://extension.zip"

# Update Lambda function to the latest version of external pushed as Lambda layers
aws lambda update-function-configuration \
  --function-name ${LAMBDA_FUNCTION} --layers $(aws lambda list-layer-versions --layer-name ${EXTENSION_NAME} \
    --max-items 1 --no-paginate --query 'LayerVersions[0].LayerVersionArn' \
    --output text)

cd -
