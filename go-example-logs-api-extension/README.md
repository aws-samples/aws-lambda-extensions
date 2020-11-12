# Example Logs API Extension in Go

The provided code sample demonstrates how to get a basic Logs API extension written in Go up and running.

> This is a simple example extension to help you start investigating the Lambda Runtime Logs API. This example code is not production ready. Use it with your own discretion after testing thoroughly.

This sample extension: 
* Subscribes to recieve platform and function logs
* Runs with a main and a helper goroutine: The main goroutine registers to ExtensionAPI and process its invoke and shutdown events (see nextEvent call). The helper goroutine:
    - starts a local HTTP server at the provided port (default 1234) that receives requests from Logs API
    - puts the logs in a synchronized queue (Producer) to be processed by the main goroutine (Consumer)
* Writes the received logs to an S3 Bucket

## Compile package and dependencies

To run this example, you will need to ensure that your build architecture matches that of the Lambda execution environment by compiling with `GOOS=linux` and `GOARCH=amd64` if you are not running in a Linux environment.

Building and saving package into a `bin/extensions` directory:
```bash
$ cd go-example-logs-api-extension
$ GOOS=linux GOARCH=amd64 go build -o bin/extensions/go-example-logs-api-extension main.go
$ chmod +x bin/extensions/go-example-logs-api-extension
```

## Layer Setup Process
The extensions .zip file should contain a root directory called `extensions/`, where the extension executables are located. In this sample project we must include the `go-example-logs-api-extension` binary.

Creating zip package for the extension:
```bash
$ cd bin
$ zip -r extension.zip extensions/
```

Publish a new layer using the `extension.zip` and capture the produced layer arn in `layer_arn`. If you don't have jq command installed, you can run only the aws cli part and manually pass the layer arn to `aws lambda update-function-configuration`.
```bash
layer_arn=$(aws lambda publish-layer-version --layer-name "go-example-logs-api-extension" --region "<use your region>" --zip-file  "fileb://extension.zip" | jq -r '.LayerVersionArn')
```

Add the newly created layer version to a Lambda function.
```bash
aws lambda update-function-configuration --region <use your region> --function-name <your function name> --layers $layer_arn
```

## Function Invocation and Extension Execution
> Note: Your function role should have the AmazonS3FullAccess policy attached.

> Note: You need to add `LOGS_API_EXTENSION_S3_BUCKET` environment variable to your lambda function. The value of this variable will be used to create a bucket or use an existing bucket if it is created previously. The logs received from Logs API will be written in a file inside that bucket. For S3 bucket naming rules, see [AWS docs](https://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html).

After invoking the function and receiving the shutdown event, you should now see log messages from the example extension written to an S3 bucket with the following name format:

`<function-name>-<timestamp>-<UUID>.log` in to the bucket set with the environment variable above.