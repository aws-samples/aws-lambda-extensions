# AWS AppConfig Extension Demo
This is a demo of [AWS Lambda](https://aws.amazon.com/lambda/) extensions using [AWS AppConfig](https://docs.aws.amazon.com/appconfig/latest/userguide/what-is-appconfig.html) as explained in the blog post "Introducing Lambda Extensions".

AWS AppConfig has an available extension to further integrate Lambda and AWS AppConfig. The extension runs a separate local process to retrieve configuration data from the AWS AppConfig service.

Using the AWS AppConfig extension, a function can fetch configuration data faster using a local call rather than over the network. You can dynamically change a function’s external configuration settings during invocations, without having to redeploy the function. As AWS AppConfig has robust validation features, all configuration changes can be tested safely before rolling out to one or more Lambda functions. 

The demo uses the [AWS Serverless Application Model (AWS SAM)](https://aws.amazon.com/serverless/sam/) to deploy two Lambda functions which include the AWS AppConfig extension layer.

As extensions share the same permissions as Lambda functions, the SAM template creates function execution roles that allow access to retrieve the AWS AppConfig configuration.

AWS SAM creates an AWS AppConfig application, environment, and configuration profile, storing a `loglevel` value, initially set to `normal`.

## Requirements

* AWS CLI already configured with Administrator permission
* [AWS SAM CLI installed](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html) - **minimum version 0.48**.

## Installation instructions

1. [Create an AWS account](https://portal.aws.amazon.com/gp/aws/developer/registration/index.html) if you do not already have one and login.

2. Clone the repo onto your local development machine:
```bash
git clone https://github.com/aws-samples/aws-lambda-extensions
```
3. Get the latest Amazon Resource Name (ARN) of the AppConfig Extension for your Region from the [AWS AppConfig Extensions Blog Post](). It is in the format `arn:aws:lambda:us-east-1:027255383542:layer:AWSAppConfigExtension:7`

4. Deploy the SAM template from the command line:
```bash
cd awsappconfig-extension-demo
sam deploy --stack-name awsappconfig-extension-demo --guided
```

During the prompts:

* Accept the default Stack Name `awsappconfig-extension-demo`.
* Enter your preferred Region
* Update the default AppConfigARN if the ARN retrieved in Step 3 is newer.
* Accept the default AppConfigProfile configuration function environment variable.
* Accept the defaults for the remaining questions.

SAM deploys the application stack

5. From the [AWS Lambda Management Console](https://console.aws.amazon.com/lambda), choose each Lambda function prefixed with `awsappconfig-extension-demo`.

Create a test event with an event payload for each function
```json
{
  "showextensions": "function1"
}
```
6. Invoke both functions using the test events. You should receive a response showing a Cold Start and LogLevel set to normal.

```json
{
  "event": {
    "extensions": "function1"
  },
  "ColdStart": true,
  "LogLevel": "normal"
}
```
7. Invoke each function again using the same test event. You should receive a response showing the invocation is not a Cold Start and LogLevel set to normal.

```json
{
  "event": {
    "extensions": "function1"
  },
  "ColdStart": false,
  "LogLevel": "normal"
}
```

8. From the [AWS AppConfig Management Console](https://console.aws.amazon.com/systems-manager/appconfig/applications), select **DemoExtensionApplication**.
Navigate to the *Configuration profiles* tab. Select the created profile, **LoggingLevel**. Create a new hosted configuration version 2 as JSON setting the `loglevel` to `verbose`
```JSON
{
  "loglevel": "verbose"
}
```
9. Deploy the configuration by chosing **Start deployment**.
* For *Environment*, select **Production**
* For *Hosted configuration version*, select version **2**
* For *Deployment strategy*, select **AllNow**
 
Once the deployment is complete, AWS AppConfig updates the configuration value for both functions. The function configuration itself is not changed.

Running another test invocation for both functions returns the updated value of `verbose` still without a cold start.
```json
{
  "event": {
    "extensions": "function2"
  },
  "ColdStart": false,
  "LogLevel": "verbose"
}
```
AWS AppConfig has updated a dynamic external configuration setting for multiple Lambda functions without having to redeploy the function configuration.

## Cleanup instructions

From the AWS AppConfig Console, delete the created *Hosted configuration version* **2**

From the [AWS Cloudformation Console](https://console.aws.amazon.com/cloudformation), select **Stacks**.

Select the **awsappconfig-extension-demo** stack, and choose **Delete**

The Lambda functions and AppConfig resources are deleted.

