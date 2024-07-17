#!/bin/bash

set -euo pipefail

main() {
    rm -rf bin
    mkdir -p bin

    ( pack_one amd64 )
    ( pack_one arm64 )
}

pack_one() {
    local -r arch=$1

    rm -rf "bin/$arch/extensions/"
    mkdir -p "bin/$arch/extensions/"

    echo "Building $arch..."
    GOPROXY=direct GOOS=linux GOARCH=$arch go build -o "bin/$arch/extensions/go-example-crash-uploader-extension"
    chmod +x "bin/$arch/extensions/go-example-crash-uploader-extension"

    echo "Packing $arch..."
    cd "bin/$arch"
    ls -alh extensions/go-example-crash-uploader-extension
    zip -r9 "../extension.$arch.zip" extensions
    ls -alh "../extension.$arch.zip"

    if [[ "${PUBLISH:-}" == "1" ]]; then
        echo "Publishing for $arch..."
        aws lambda publish-layer-version \
            --layer-name "crash-uploader-$arch" \
            --region us-west-2 \
            --zip-file  "fileb://../extension.$arch.zip"
    fi
}

main "$@"
