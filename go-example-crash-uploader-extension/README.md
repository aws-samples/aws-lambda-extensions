# Example Crash Uploader Extension in Go

The provided code sample implements a sample extension that looks for core dumps in the execution environment and uploads them to an Amazon S3 bucket for later inspection and troubleshooting.

## Quick Start

1. Run `./pack.sh` which will compile the crash uploader extension for both x86_64 and arm64 architectures.

    ```bash
    # Build and create the zip only:
    ./pack.sh

    # Build and publish the zip files as layers:
    PUBLISH=1 ./pack.sh
    ```

1. You may need to increase your `/tmp` size, depending on the expected size of the core dump: <https://aws.amazon.com/blogs/aws/aws-lambda-now-supports-up-to-10-gb-ephemeral-storage/>
1. `./pack.sh` will create one layer zip file per architecture, upload and attach it to your target Lambda function.
    1. Note the LayerVersionArn that is produced in the output.
        eg. `"LayerVersionArn": "arn:aws:lambda:<region>:123456789012:layer:<layerName>:1"`
1. Add a new `BUCKET` env var to the target Lambda function with the S3 Bucket core dumps would be uploaded to.
1. Lambda function needs to have permission to upload to the specified bucket.

## Function Invocation and Extension Execution

When invoking the function, you should now see log messages from the example extension similar to the following:

```
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    EXTENSION Name: go-example-crash-uploader-extension State: Ready Events: [INVOKE,SHUTDOWN]
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    START RequestId: 9ca08945-de9b-46ec-adc6-3fe9ef0d2e8d Version: $LATEST
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [go-example-crash-uploader-extension]  Registering...
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [go-example-crash-uploader-extension]  Register response: {
                "functionName": "my-function",
                "functionVersion": "$LATEST",
                "handler": "function.handler"
        }
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [go-example-crash-uploader-extension]  Waiting for event...
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [go-example-crash-uploader-extension]  Received event: {
                "eventType": "INVOKE",
                "deadlineMs": 1234567890123,
                "requestId": "9ca08945-de9b-46ec-adc6-3fe9ef0d2e8d",
                "invokedFunctionArn": "arn:aws:lambda:<region>:123456789012:function:my-function",
                "tracing": {
                        "type": "X-Amzn-Trace-Id",
                        "value": "XXXXXXXXXX"
                }
        }
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [go-example-crash-uploader-extension]  Waiting for event...
    ...
    ...
    Function logs...
    ...
    ...
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    END RequestId: 9ca08945-de9b-46ec-adc6-3fe9ef0d2e8d
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    REPORT RequestId: 9ca08945-de9b-46ec-adc6-3fe9ef0d2e8d Duration: 3.78 ms Billed Duration: 100 ms Memory Size: 128 MB Max Memory Used: 59 MB Init Duration: 264.75 ms
```
