# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0
# Build artifacts
mvn clean install

# Create zip
chmod +x extensions/java-example-extension
archive="extension.zip"
if [ -f "$archive" ] ; then
    rm "$archive"
fi
zip "$archive" -j target/java-example-extension-1.0-SNAPSHOT.jar
zip "$archive" extensions/*
