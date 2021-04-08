# Fluent Bit Lambda Extensions in Go

The provided code sample demonstrates how to get Fluent Bit set up for Lambda Extensions. 

Note: 
1. Supported Fluent Bit: v1.6
2. Supported AWS Fluent Bit Outputs: CloudWatch Logs, Kinesis Firehose and Kinesis Data Streams.

This extension runs a goroutine that will do the following:
* Starts the Extension and registers to ExtensionAPI,
* Invokes Fluent Bit binary using user-specified configurations,
* Subscribes the extension to the Logs API

## Prerequisites

1. You need to compile a Fluent Bit binary. The binary should be placed inside the `fluent-bit/` directory. Here are the steps:
   1. Clone Fluent Bit on a linux machine (Amazon Linux is recommended): https://github.com/fluent/fluent-bit
   2. Run the following:
   ```
      cd fluent-bit/build
      cmake -DFLB_IN_SYSTEMD=Off ../
      make
   ```
      
   **Note:** you must use cmake version 3, and on some platforms its command may be "cmake3" instead of "cmake"


2. Configuration files: In `fluent-bit/` is a typical set of configuration files for Fluent Bit:
    
    - fluent-bit.conf: This is the main configuration file. 
    - Output files: There are three example output files. You will need to change the `@INCLUDE` under `fluent-bit.conf` to use a specific one. See [here](https://docs.fluentbit.io/manual/pipeline/outputs) for a list of other configurable outputs.
         - output-cw.conf: CloudWatch Logs
         - output-firehose.conf: Kinesis Firehose
         - output-kinesis.conf: Kinesis Data Streams
    - parsers.conf: This is a specific parser configuration file for ingesting Lambda logs.

## Compile package and dependencies

To run this example, you will need to ensure that your build architecture matches that of the Lambda execution environment by compiling with `GOOS=linux` and `GOARCH=amd64` if you are not running in a Linux environment.

Building and saving package into a `bin/extensions` directory:
```bash
$ cd aws-lambda-extensions-fluent-bit
$ GOOS=linux GOARCH=amd64 go build -o bin/extensions/fluent-bit-extension main.go
$ chmod +x bin/extensions/fluent-bit-extension
```

## Layer Setup Process
The extensions .zip file should contain a root directory called `extensions/`, where the extension executables are located. In this sample project we must include the `fluent-bit-extension` binary, as well as the Fluent Bit binary and the configuration files under `fluent-bit/`.

Creating zip package for the extension:
```bash
$ cp -r fluent-bit/ bin/extensions/fluent-bit
$ cd bin
$ zip -r extension.zip extensions/
```

Publish a new layer using the `extension.zip` and capture the produced layer arn in `layer_arn`. If you don't have jq command installed, you can run only the aws cli part and manually pass the layer arn to `aws lambda update-function-configuration`.
```bash
layer_arn=$(aws lambda publish-layer-version --layer-name "fluent-bit-extension" --region "<use your region>" --zip-file  "fileb://extension.zip" | jq -r '.LayerVersionArn')
```

Add the newly created layer version to a Lambda function.
```bash
aws lambda update-function-configuration --region <use your region> --function-name <your function name> --layers $layer_arn
```

## Function Invocation and Extension Execution

> Note: Your function role should have permission to write to AWS resources for any AWS outputs (i.e. CloudWatch, Kinesis, or Firehose).

> Note: Within your Lambda function, set environment variable "PORT" as the port you want to use for the HTTP listener on Fluent Bit. Otherwise the value defaults to 1234.

After invoking the function and receiving the shutdown event, you should now see log messages from Fluent Bit sent to the output of your choice. 

Using the CloudWatch Logs output provided as an example configuration, several events will be written in a log stream `lambda-logs-lambda.http` inside the log group `fluent-bit-lambda-logs`. The events will look something like this:

```
   {
      'time': '2020-11-12T00:30:48.883Z', 
      'record': {
         'requestId': '49ae0e5f-bc60-4521-81e3-6e41d6bcb55c', 
         'version': '$LATEST'
      }
   }
```

