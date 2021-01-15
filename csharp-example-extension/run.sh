#!/bin/bash
# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

###### START: parse command line arguments

SELF_CONTAINED=false

print_usage() {
    echo "Usage: $(basename $0) -e <extension_name> -f <function_name> [-s]"
    echo "  -e: extension/layer name"
    echo "  -f: lambda fucntion name to be attached to the layer. Function must exist already!!!"
    echo "  -s: (optional) extension will be publsihed as a standalone package that doesn't depend on any pre-configured runtime. Extension will depend on .NET Core 3.1 lambda runtime if this flag is not set."
}

while getopts ":e:f:sh" opt; do
  case ${opt} in
    e ) EXTENSION_NAME=$OPTARG
      ;;
    f ) LAMBDA_FUNCTION=$OPTARG
      ;;
    s ) SELF_CONTAINED=true
      ;;
    h )
      print_usage
      exit 0
      ;;
    : )
      echo "Invalid option: $OPTARG requires an argument" 1>&2
      exit 1
      ;;
  esac
done

shift $((OPTIND -1))

if [ -z "$EXTENSION_NAME" -o -z "$LAMBDA_FUNCTION" ] ; then
    print_usage
    exit 100
fi

echo "Using configuration:"
echo "  Extension/layer name: ${EXTENSION_NAME}"
echo "  Lambda function: ${LAMBDA_FUNCTION}"
echo "  Self contained deployment: ${SELF_CONTAINED}"
echo ''

###### END: parse command line arguments

rm -rf bin
rm -rf obj

if [ "${SELF_CONTAINED}" = "true" ] ; then
    echo 'Building self-contained extension...'
    echo ''

    dotnet publish -c Release -f net5.0
    cd bin/Release/net5.0/linux-x64/publish
else
    echo 'Building .NET Core 3.1 dependent extension...'
    echo ''

    dotnet publish -c Release -f netcoreapp3.1
    cd bin/Release/netcoreapp3.1/publish
fi

zip -rm ./deploy.zip *

aws lambda publish-layer-version \
	--layer-name "${EXTENSION_NAME}" \
	--zip-file "fileb://deploy.zip"

aws lambda update-function-configuration \
	--function-name ${LAMBDA_FUNCTION} --layers $(aws lambda list-layer-versions --layer-name ${EXTENSION_NAME} \
	--max-items 1 --no-paginate --query 'LayerVersions[0].LayerVersionArn' \
	--output text)

