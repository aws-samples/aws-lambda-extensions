# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0
# Build artifacts
mvn clean install

# Perform cleanup & create zip
rm -Rf extensions/*.jar
rm -Rf extensions/*.zip
mv target/java-example-extension-1.0-SNAPSHOT.jar extensions
chmod +x extensions/java-example-extension
cd extensions
zip -r extension.zip .
mv extension.zip ../
cd -
