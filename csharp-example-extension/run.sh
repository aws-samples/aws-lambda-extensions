#!/bin/bash
# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

EXTENSION_NAME=$1
LAMBDA_FUNCTION=$2

rm -rf bin
rm -rf obj

dotnet publish -c Release

cd bin/Release/netcoreapp3.1/publish

zip -rm ./deploy.zip *

aws lambda publish-layer-version \
    --compatible-runtimes "dotnetcore3.1" \
	--layer-name "${EXTENSION_NAME}" \
	--zip-file "fileb://deploy.zip"

aws lambda update-function-configuration \
	--function-name ${LAMBDA_FUNCTION} --layers $(aws lambda list-layer-versions --layer-name ${EXTENSION_NAME} \
	--max-items 1 --no-paginate --query 'LayerVersions[0].LayerVersionArn' \
	--output text)

