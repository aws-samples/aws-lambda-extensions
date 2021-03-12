# Example Logs API Extension in Node.js

The provided code sample demonstrates how to get a basic Logs API extension written in Node.js 12 up and running.

> Note: This extension requires the Node.js 12 runtime to be present in the Lambda execution environment of your function.

There are two components to this sample:

* `extensions/`: This sub-directory should be extracted to /opt/extensions where the Lambda platform will scan for executables to launch extensions
* `nodejs-example-logs-api-extension/`: This sub-directory should be extracted to /opt/nodejs-example-logs-api-extension which is referenced by the `extensions/nodejs-example-logs-api-extension` executable and includes a nodejs executable along with all of its necessary dependencies.

## Prep Extension Dependencies

Install the extension dependencies locally, which will be mounted along with the extension code.

```bash
cd nodejs-example-logs-api-extension
chmod +x index.js
npm install
cd ..
```

## Layer Setup Process

The extensions .zip file should contain a root directory called `extensions/`, where the extension executables are located and another root directory called `nodejs-example-logs-api-extension/`, where the core logic of the extension and its dependencies are located.

Creating zip package for the extension:

```bash
chmod +x extensions/nodejs-example-logs-api-extension
zip -r extension.zip ./nodejs-example-logs-api-extension
zip -r extension.zip ./extensions
```

Ensure that you have aws-cli v2 for the commands below.
Publish a new layer using the `extension.zip`. The output of the following command should provide you a layer arn.

```bash
aws lambda publish-layer-version \
 --layer-name "nodejs-example-logs-api-extension" \
 --region <use your region> \
 --zip-file  "fileb://extension.zip"
```

Note the LayerVersionArn that is produced in the output.
eg. `"LayerVersionArn": "arn:aws:lambda:<region>:123456789012:layer:<layerName>:1"`

Add the newly created layer version to a Node.js 12 runtime Lambda function.

## Upload to S3

To upload logs to S3, add the `LOGS_S3_BUCKET_NAME` Environment Variable and add S3 write permissions for that bucket to your Lambda's IAM Role.

**All these instructions were arranged together for convenience in `deploy.sh`**
