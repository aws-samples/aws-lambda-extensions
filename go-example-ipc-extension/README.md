# Example Inter-Process Communication (IPC) Extension in Go

The provided code sample illustrates a sample extension written in Go that acts as a companion process which the AWS Lambda function runtime can communicate with.

This sample extension:

- Writes to a file at /tmp/test.txt that can be read by the Lambda function
- Starts a local HTTP server at the provided port (default 2772) that replies to requests with "Hello from http server"

## Compile package and dependencies

To run this example, you will need to ensure that your build architecture matches that of the Lambda execution environment by compiling with `GOOS=linux` and `GOARCH=amd64` if you are not running in a Linux environment.

Building and saving package into a `bin/extensions` directory:
```bash
$ cd go-example-ipc-extension
$ GOOS=linux GOARCH=amd64 go build -v -ldflags '-d -s -w' -a -tags netgo -installsuffix netgo -o bin/extensions/go-example-ipc-extension main.go
$ chmod +x bin/extensions/go-example-ipc-extension
```

## Layer Setup Process
The extensions .zip file should contain a root directory called `extensions/`, where the extension executables are located. In this sample project we must include the `go-example-ipc-extension` binary.

Creating zip package for the extension:
```bash
$ cd bin
$ zip -r extension.zip extensions/
```

Ensure that you have aws-cli v2 for the commands below.
Publish a new layer using the `extension.zip`. The output of the following command should provide you a layer arn.
```bash
aws lambda publish-layer-version \
 --layer-name "go-example-ipc-extension" \
 --region <use your region> \
 --zip-file  "fileb://extension.zip"
```
Note the LayerVersionArn that is produced in the output.
eg. `"LayerVersionArn": "arn:aws:lambda:<region>:123456789012:layer:<layerName>:1"`

Add the newly created layer version to a Lambda function.

> Note: You can use the provided `hello.sh` in the `function/` directory with a custom (provided or provided.al2) runtime to see the IPC integartions via network and file.


## Function Invocation and Extension Execution

When invoking the function, you should now see log messages from the example extension similar to the following:
```
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    EXTENSION Name: go-example-ipc-extension State: Ready Events: [INVOKE,SHUTDOWN]
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    START RequestId: 9ca08945-de9b-46ec-adc6-3fe9ef0d2e8d Version: $LATEST
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [go-example-ipc-extension]  Registering...
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [go-example-ipc-extension]  Register response: {
                "functionName": "my-function",
                "functionVersion": "$LATEST",
                "handler": "function.handler"
        }
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [go-example-ipc-extension]  Waiting for event...
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [go-example-ipc-extension]  Received event: {
                "eventType": "INVOKE",
                "deadlineMs": 1234567890123,
                "requestId": "9ca08945-de9b-46ec-adc6-3fe9ef0d2e8d",
                "invokedFunctionArn": "arn:aws:lambda:<region>:123456789012:function:my-function",
                "tracing": {
                        "type": "X-Amzn-Trace-Id",
                        "value": "XXXXXXXXXX"
                }
        }
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [go-example-ipc-extension]  Waiting for event...
    ...
    ...
    Function logs...
    ...
    ...
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    END RequestId: 9ca08945-de9b-46ec-adc6-3fe9ef0d2e8d
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    REPORT RequestId: 9ca08945-de9b-46ec-adc6-3fe9ef0d2e8d Duration: 3.78 ms	Billed Duration: 100 ms	Memory Size: 128 MB	Max Memory Used: 59 MB	Init Duration: 264.75 ms
```

If you used the example `hello.sh` for your function, you should see a response similar to:
```
Hello from http server
Hello I'm a temp file
Echoing request: '<payload from request>'
```