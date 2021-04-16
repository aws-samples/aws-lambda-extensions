# Adaptive Batching Extension in Go

The provided code sample demonstrates how to use the Logs API extension written in Go to adaptively batch log data to a destination (S3). 

> This is a simple example extension to help you start investigating the Lambda Runtime Logs API. This example code is not production ready. Use it with your own discretion after testing thoroughly.

This sample extension: 
* Subscribes to receive platform and function logs
* Runs with a main and a helper goroutine: The main goroutine registers to ExtensionAPI and process its invoke and shutdown events (see nextEvent call). The helper goroutine:
    - starts a local HTTP server at the provided port (default 1234) that receives requests from Logs API
    - puts the logs in a synchronized queue (Producer) to be processed by the main goroutine (Consumer)
* The main go routine tracks the number of invokes, last time logs were shipped, and the size of the logs that have been accumulated since the last ship using a structure. Once one of these fields exceeds a set value the log shipping process begins and a new file with the logs is created in S3.

## Requirements

* [AWS SAM CLI ](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html) - **minimum version 0.48**.

## Installation instructions

1. [Create an AWS account](https://portal.aws.amazon.com/gp/aws/developer/registration/index.html) if you do not already have one and login.

2. Clone the repository onto your local development machine:
```bash
git clone https://github.com/aws-samples/aws-lambda-extensions.git
```
3. Enter the directory 
```bash
cd go-example-adaptive-batching/
```

4. Run the following command for AWS SAM to deploy the components as specified in the `template.yaml` file:
```bash
sam build
# If you don't have 'Python' or 'make' installed, you can use the option to build using a container which uses a python3.8 Docker container image
# sam build --use-container
sam deploy --stack-name adaptive-batching-extension --guided
```

Following the prompts:

* Accept the default Stack Name `adaptive-batching-extension`.
* Enter your preferred Region
* Accept the defaults for the remaining questions.

AWS SAM deploys the application stack which includes the Lambda function and an IAM Role. AWS SAM creates a layer for the runtime, a layer for the extension, and adds them to the function.

Note the outputted S3 bucket name. This is where the logs will be outputted. 

## Invoking the Lambda function
You can now invoke the Lambda function. Amend the Region and use the following command:
```bash
aws lambda invoke \
 --function-name "adaptive-batching-extension-demo-function" \
 --payload '{"payload": "hello"}' /tmp/invoke-result \
 --cli-binary-format raw-in-base64-out \
 --log-type Tail \
 --region <use your Region>
```
The function should return `"StatusCode": 200`

Browse to the [Amazon CloudWatch Console](https://console.aws.amazon.com/cloudwatch). Navigate to *Logs\Log Groups*. Select the log group **/aws/lambda/adaptive-batching-extension-demo-function**.

View the log stream to see the platform, function, and extensions each logging while they are processing.

The logging extension also receives the log stream directly from Lambda, and copies the logs to S3.

Browse to the [Amazon S3 Console](https://console.aws.amazon.com/S3). Navigate to the S3 bucket created as part of the SAM deployment. 

Downloading the file object containing the copied log stream. The log contains the same platform and function logs, but not the extension logs, as specified during the subscription.


## Environment Variables

This section details environment variables that can be used to modify the functionality of the extension. Some are required and are marked as such. 

* ADAPTIVE_BATCHING_EXTENSION_S3_BUCKET : (REQUIRED) The value of this variable will be used to create a bucket or use an existing bucket if it is created previously. The logs received from Logs API will be written in a file inside that bucket. For S3 bucket naming rules, see [AWS docs](https://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html). The SAM template will set this by default. 
* ADAPTIVE_BATCHING_EXTENSION_SHIP_RATE_BYTES : Logs are shipped to S3 once log size reaches the number of bytes defined here. For example a value of 1024 here would result in logs being shipped once the log size exceeds 1 kilobyte in size. Default value of 4096 bytes (4 kilobytes).
* ADAPTIVE_BATCHING_EXTENSION_SHIP_RATE_INVOKES : Logs are shipped to S3 once the number of invokes reaches the number defined here. For example a value of 10  would result in logs being shipped at least once every 10 invokes since the last time logs were shipped. Default value is 10 invokes. 
* ADAPTIVE_BATCHING_EXTENSION_SHIP_RATE_MILLISECONDS : Logs are shipped to S3 once the amount of time elapsed since the last time logs were shipped is exceeded. For example a value of 60,000 here would result in logs being shipped once every 60 seconds has passed. The default value is 10,0000 milliseconds. 
* ADAPTIVE_BATCHING_EXTENSION_LOG_TYPES: This is a JSON array the log types that can be requested from the Logs API. These are the supported log types `["platform", "function", "extension"]`. If not included or parsing errors occur, the log types default to `["platform", "function"]`. 

## Performance, maximums, and environment shutdown 

Logs are shipped to S3 based off of the rates defined in the environment variables or the default values provided. If any of the conditions are met, the logs queued will be shipped to S3. The rates defined are only checked when an invoke to lambda occurs. This means for ADAPTIVE_BATCHING_EXTENSION_SHIP_RATE_MILLISECONDS, the time elapsed may be exceeded and the metric will only be checked once an invocation occurs. So for a rate of 100 ms, if there are 200 ms gaps between lambda invocations, logs will be shipped every 200 ms, once invocations occur. 

Lambda extensions share Lambda function resources, like CPU, memory, and storage, with your function code. The extension here has a limit written into `agent/metrics.go` with MAX_SHIP_RATE_BYTES to prevent users from using too much memory. It is currently set to 100 megabytes, meaning the maximum rate that can be set for ADAPTIVE_BATCHING_EXTENSION_SHIP_RATE_BYTES is 100 megabytes. Any value set above will default to that maximum value of 100 megabytes. That maximum can be increased by modifying the constant in `agent/metrics/go`.

In the case of the Lambda environment shutting down, either from error or stagnation the extension will flush the log queue and upload the final log file to S3. For information about the Lambda environment shutdown phase, see [AWS docs](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-extensions-api.html#runtimes-lifecycle-shutdown).

## Outputs 

Multiple files will show up in the S3 bucket depending on how many logs are generated and the thresholds defined. Files are named as follows, `<function-name>-<timestamp>-<UUID>.log`. 



