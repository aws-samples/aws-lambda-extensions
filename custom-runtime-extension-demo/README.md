# Custom Runtime Extension(s) Demo

This is a demo of AWS Lambda extensions. The demo uses the custom runtime `provided.al2` with one or more extensions delivered as Lambda Layers.

The runtime, function, and extension(s) each log output to Amazon CloudWatch Logs. This shows how you can generate logs before, during, and after function invocation. The extension(s) register and then run a loop, sleeping for 5 seconds to simulate processing and then wait for a function invocation.

You create three .zip files for this demo to create a custom runtime with extensions. One is for the runtime, one for the function code, and one for the extension(s).

There are two deployment options to create the demo function. Either creating each component individually using the [AWS CLI V2](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html) or deploying all components together using the [AWS Serverless Application Model (AWS SAM)](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)


There are two deployment options to create the demo function. 

1.  Deploying all components together using the [AWS Serverless Application Model (AWS SAM)](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)
2. Creating each component individually using the [AWS CLI V2](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html) or 

## Requirements

Depending on which deployment method you choose, ensure you have the correct installation:
* [AWS SAM CLI ](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html) - **minimum version 0.48**.
* [AWS CLI V2](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html) configured with Administrator permission.

## Installation instructions

1. [Create an AWS account](https://portal.aws.amazon.com/gp/aws/developer/registration/index.html) if you do not already have one and login.

2. Clone the repo onto your local development machine:
```bash
git clone https://github.com/aws-samples/aws-lambda-extensions
cd aws-lambda-extensions/custom-runtime-extension-demo
```
Now select, either one of the deployment options
## Deployment option 1: All components using AWS SAM
1. Run the following command for AWS SAM to deploy the components as specified in the `template.yml` file:
```bash
sam deploy --stack-name custom-runtime-extension-demo --guided
```

During the prompts:

* Accept the default Stack Name `custom-runtime-extension-demo`.
* Enter your preferred Region
* Accept the defaults for the remaining questions.

AWS SAM deploys the application stack which includes the Lambda function and an IAM Role. AWS SAM creates a layer for the runtime, a layer for the extension, and adds them to the function.

Skip deployment option 2 and continue with **Invoke the Lambda function**

## Deployment Option 2: Each component individually using AWS-CLI-V2
### 1.  Create the custom `runtime.zip` file
Create a `runtime.zip` file with the executable custom shell script named `bootstrap` in the root directory.

```
$ chmod +x runtime/bootstrap
$ zip -j runtime.zip runtime/bootstrap
$ zipinfo runtime.zip
Archive:  runtime.zip
Zip file size: 1181 bytes, number of entries: 1
-rwxrwxrwx  3.0 unx     1544 tx defN 20-Jul-06 12:41 bootstrap
1 file, 1707 bytes uncompressed, 863 bytes compressed:  49.4%
```

### 2.  Create the `function.zip` file
Create a `function.zip` file with the `function.sh` file in the root directory.

```
$ chmod +x functionsrc/function.sh
$ zip -j function.zip function/function.sh
$ zipinfo function.zip
Archive:  function.zip
Zip file size: 1181 bytes, number of entries: 1
-rwxrwxrwx  3.0 unx      163 tx defN 20-Sep-03 13:04 function.sh
1 file, 1707 bytes uncompressed, 863 bytes compressed:  49.4%
```

### 3.  Create the `extension(s).zip` file
Create an `extensions.zip` file containing a root directory called extensions" in which the extension executables or scripts are located.

Example for single extension:
```bash
$ cd extensionssrc
$ chmod +x extensions/extension1.sh
$ zip ../extensions.zip extensions/extension1.sh
$ cd ..
$ zipinfo extensions.zip 
Archive:  extensions.zip
Zip file size: 1029 bytes, number of entries: 1
-rwxrwxrwx  3.0 unx     1834 tx defN 20-Sep-07 15:38 extensions/extension1.sh
1 file, 1834 bytes uncompressed, 831 bytes compressed:  54.7%
```
Example for multiple extensions:
```bash
$ cd extensionssrc
$ chmod +x extensions/extension1.sh
$ chmod +x extensions/extension2.sh
$ zip ../extensions.zip extensions/extension1.sh extensions/extension2.sh
$ cd ..
$ zipinfo extensions.zip 
Archive:  extensions.zip
Zip file size: 2036 bytes, number of entries: 2
-rwxrwxrwx  3.0 unx     1834 tx defN 20-Sep-07 15:38 extensions/extension1.sh
-rwxrwxrwx  3.0 unx     1834 tx defN 20-Sep-07 15:38 extensions/extension2.sh
2 files, 3668 bytes uncompressed, 1662 bytes compressed:  54.7%```
```
### 4. Create a Lambda layer containing the custom runtime
Publish a new layer using `runtime.zip`. The output of the following command provides a layer arn.
```bash
aws lambda publish-layer-version \
 --layer-name "custom-runtime-layer" \
 --region <use your Region> \
 --zip-file  "fileb://runtime.zip"
```
Note the *LayerVersionArn* that is produced in the output.

eg. `"LayerVersionArn":"arn:aws:lambda:<region>:123456789012:layer:custom-runtime-layer:1"`

### 5. Create a Lambda layer containing the extension(s)

Publish a new layer using `extensions.zip`. The output of the following command provides a layer arn.
```bash
aws lambda publish-layer-version \
 --layer-name "extensions-layer" \
 --region <use your Region> \
 --zip-file  "fileb://extensions.zip"
```
Note the *LayerVersionArn* that is produced in the output.

eg. `"LayerVersionArn":"arn:aws:lambda:<region>:123456789012:layer:extensions-layer:1"`

### 6. Create a Lambda execution role.

Create a Lambda IAM execution role and attach a role policy. 
Note the Arn that is produced in the output during the `create-role` step.

eg. `"Arn": "arn:aws:iam::123456789123:role/custom-runtime-extension-demo-role"`
```bash
aws iam create-role --role-name custom-runtime-extension-demo-role --assume-role-policy-document '{"Version": "2012-10-17","Statement": [{ "Effect": "Allow", "Principal": {"Service": "lambda.amazonaws.com"}, "Action": "sts:AssumeRole"}]}'
aws iam attach-role-policy --role-name custom-runtime-extension-demo-role --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
```

### 7. Create a Lambda function adding the custom runtime and extensions layers

Create a new Lambda function using the custom runtime `runtime.zip`. Amend the region and attach the layers and role arns created in the previous steps:
```bash
aws lambda create-function \
 --function-name "custom-runtime-extension-demo-function" \
 --runtime "provided.al2" \
 --region <use your Region> \
 --role <use the Lambda IAM role arn created previously> \
 --layers <use the runtime and extensions layers previously created in the format "arn:aws:lambda:<Region>:1234567890123:layer:custom-runtime-layer:1" "arn:aws:lambda:<Region>:123456789012:layer:extensions-layer:1"> \
 --timeout 120 \
 --handler "function.handler" \
 --zip-file "fileb://function.zip"
```
## Invoke the Lambda function
You can now invoke the Lambda function. Amend the Region and use the following command:
```bash
aws lambda invoke \
 --function-name "custom-runtime-extension-demo-function" \
 --payload '{"payload": "hello"}' /tmp/invoke-result \
 --cli-binary-format raw-in-base64-out \
 --log-type Tail \
 --region <use your Region>
```
The function should return `"StatusCode": 200`

Browse to the [Amazon CloudWatch Console](https://console.aws.amazon.com/cloudwatch). Navigate to *Logs\Log Groups*. Select the log group **/aws/lambda/custom-runtime-extension-demo-function**.

View the log stream to see the extension, runtime, and function each logging while they are processing.