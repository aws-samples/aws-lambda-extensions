# S3-Logs extension demo using container image format

This is a demo of the logging functionality available with [AWS Lambda](https://aws.amazon.com/lambda/) Extensions to send logs directly from Lambda to [Amazon Simple Storage Service (S3)](https://aws.amazon.com/s3/).

This example packages the extension and function as separate container images. See the s3-logs-extension-demo-zip-archive version to use the functionality without container images.

For more information on the extensions logs functionality, see the blog post [Using AWS Lambda extensions to send logs to custom destinations](https://aws.amazon.com/blogs/compute/using-aws-lambda-extensions-to-send-logs-to-custom-destinations/)

> This is a simple example extension to help you start investigating the Lambda Runtime Logs API. This code is not production ready, and it has never intended to be. Use it with your own discretion after testing thoroughly.  

The demo includes a Lambda function with an extension delivered as a container image.

The extension uses the Extensions API to register for INVOKE and SHUTDOWN events. The extension, using the Logs API, then subscribes to receive platform and function logs, but not extension logs. The extension runs a local HTTP endpoint listening for HTTP POST events. Lambda delivers log batches to this endpoint. The code can be amended (see the comments) to handle each log record in the batch. This can be used to process, filter, and route individual log records to any preferred destination

The example creates an S3 bucket to store the logs. A Lambda function is configured with an environment variable to specify the S3 bucket name. Lambda streams the logs to the extension. The extension copies the logs to the S3 bucket.

The extension uses the Python runtime from the execution environment to show the functionality. The recommended best practice is to compile your extension into an executable binary and not rely on the runtime.

The demo builds a separate container image for the extension.

The [AWS Serverless Application Model (AWS SAM)](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html) is used to build the function as another container image, which includes the previously created extension container image.

## Requirements

* [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html) - **minimum version 0.48**.
* [Docker CLI](https://docs.docker.com/get-docker/)

[Create an AWS account](https://portal.aws.amazon.com/gp/aws/developer/registration/index.html) if you do not already have one and login.

Clone the repo onto your local development machine:
```bash
git clone https://github.com/aws-samples/aws-lambda-extensions
cd s3-logs-extension-demo-container-image
```

## Installation instructions

### Create the extension container image
1. Create an [Amazon Elastic Container Registry (ECR)](https://aws.amazon.com/ecr/) repository for the extension.
```bash
aws ecr create-repository --repository-name log-extension-image
```
Note the `repositoryUri` created.

2. Run the following command to build the container image containing the extension as specified in the `Dockerfile` file.
```bash
cd .\extension\
docker build -t log-extension-image:latest  .
```
3. Tag, login, and push the extension container image to an existing ECR repository.

Replace the `repositoryUri` with the one previous created.

Replace the `AccountID` with the your AWS Account ID.
5. 
```
docker tag log-extension-image:latest <repositoryUri>:latest
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <AccountID>.dkr.ecr.us-east-1.amazonaws.com
docker push <repositoryUri>:latest
cd ..
```
### Create the Lambda function container image

1. Create an [Amazon Elastic Container Registry (ECR)](https://aws.amazon.com/ecr/) repository for the extension
```bash
aws ecr create-repository --repository-name log-extension-function
```

Note the `repositoryUri` created.

There is no need to run `pip install` for the function as it does not require any dependencies. The function Dockerfile contains the following lines, which add the previously created extension image. Replace the first FROM line with the `repositoryUri` previously created for the extension layer.

```
FROM 123456789012.dkr.ecr.us-east-1.amazonaws.com/log-extension-image:latest AS layer
FROM public.ecr.aws/lambda/python:3.8
# Layer code
WORKDIR /opt
COPY --from=layer /opt/ .

# Function code
WORKDIR /var/task
COPY app.py .

CMD ["app.lambda_handler"]
```

The AWS SAM template specified in the `template.yml` file builds the Lambda function as a container image.

Update the `/function/Dockerfile` with the extension container image location previously created.

1. Run the following commands

```
cd function
sam build
sam deploy --stack-name s3-logs-extension-demo --guided

```

During the prompts:

* Accept the default Stack Name `s3-logs-extension-demo`.
* Enter your preferred Region
* Enter the function container image ECR repository such as: `123456789012.dkr.ecr.us-east-1.amazonaws.com/log-function-image`
* Accept the defaults for the remaining questions.

AWS SAM deploys the application stack which includes the Lambda function and an IAM Role. 

Note the outputted S3 Bucket Name.

## Invoke the Lambda function
You can now invoke the Lambda function. Amend the Region and use the following command:
```bash
aws lambda invoke \
 --function-name "logs-extension-demo-function-container-image" \
 --payload '{"payload": "hello"}' /tmp/invoke-result \
 --cli-binary-format raw-in-base64-out \
 --log-type Tail \
 --region <use your Region>
```
The function should return `"StatusCode": 200`

Browse to the [Amazon CloudWatch Console](https://console.aws.amazon.com/cloudwatch). Navigate to *Logs\Log Groups*. Select the log group **/aws/lambda/logs-extension-demo-function**.

View the log stream to see the platform, function, and extensions each logging while they are processing.

The logging extension also receives the log stream directly from Lambda, and copies the logs to S3.

Browse to the [Amazon S3 Console](https://console.aws.amazon.com/S3). Navigate to the S3 bucket created as part of the SAM deployment. 

Downloading the file object containing the copied log stream. The log contains the same platform and function logs, but not the extension logs, as specified during the subscription.