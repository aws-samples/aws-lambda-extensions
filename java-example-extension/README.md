# Example Extension in Java 11

The project source includes function code and supporting resources:

- `src/main` - A Java extension.
- `pom.xml` - A Maven build file.
- `run.sh` - Shell script that will build the extension, create zip file, publish the lambda layer and update the lambda function with the latest version of this extension
- `zip.sh` - Shell script that builds the code and packages it extension.zip

Use the following instructions to setup and deploy the sample extension.

# Requirements

- [Corretto 11](https://docs.aws.amazon.com/corretto/latest/corretto-11-ug/downloads-list.html)
- [Maven 3](https://maven.apache.org/docs/history.html)
- The Bash shell. For Linux and macOS, this is included by default. In Windows 10, you can install the [Windows Subsystem for Linux](https://docs.microsoft.com/en-us/windows/wsl/install-win10) to get a Windows-integrated version of Ubuntu and Bash.
- [The AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) v1.17 or newer.
- Since this is an external java extension we would require a lambda function with java 11 runtime published to test this extension. For more information on publishing a blank java lambda function please [refer here](https://github.com/awsdocs/aws-lambda-developer-guide/tree/master/sample-apps/blank-java)

If you use the AWS CLI v2, add the following to your [configuration file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) (`~/.aws/config`):

```
cli_binary_format=raw-in-base64-out
```

This setting enables the AWS CLI v2 to load JSON events from a file, matching the v1 behavior.

# Setup

Download or clone this repository.

    $ git clone https://github.com/aws-samples/aws-lambda-extensions.git
    $ cd java-example-extension

To build a Lambda layer that contains the extension dependencies, run `mvn clean install`. Packaging dependencies in a layer reduces the size of the deployment package that you upload when you modify your code.

    java-example-extension$ mvn clean install

# Deployment instruction
You can deploy the extension by either following step-by-step instruction or using the shell script

## Step-by-step instruction
1. Build and create extension.zip
* Change the permission on the executable - Run the following command to change the permission of zip.sh

  ```
  java-example-extension$ chmod +x zip.sh
  ```

* Execute zip.sh - Run the following command to build the project and create extension.zip

  ```
  java-example-extension$ ./zip.sh
  ```

2. Set the environment variables - Run the following command to update environment variables so they can point to actual artifacts

  ```
  EXTENSION_NAME=<<ExtensionName>>
  LAMBDA_FUNCTION=<<LambdaFunctionName>>
  ```
   
2. Deploy the extension - Run the following command to deploy the extension as a lambda layer 

  ```
  java-example-extension$ aws lambda publish-layer-version \
  --layer-name "${EXTENSION_NAME}" \
  --zip-file "fileb://extension.zip"
  ```

3. Update the lambda function with the layer - Run the following command to update the lambda function to point to the latest version of the lambda 
layer that we uploaded in the previous step

  ```
  java-example-extension$ aws lambda update-function-configuration \
  --function-name ${LAMBDA_FUNCTION} \
  --layers $(aws lambda list-layer-versions --layer-name ${EXTENSION_NAME} --max-items 1 \
  --no-paginate --query 'LayerVersions[0].LayerVersionArn' --output text)
  ```

## Deploy using the shell script
To deploy the extension using the bash script do the following:
* Change the permission of the executable - Run the following command to change the permission of run.sh

  ```
  java-example-extension$ chmod +x run.sh
  ```

* Execute run.sh - Run the following command `run.sh <<extension-name>> <<function-name>>` with the extension name and function name as parameters.

  ```
  java-example-extension$ ./run.sh java-extension blank-java
  ```   

This script uses AWS CLI to perform the following:
- Builds the lambda extension
- Creates deployment zip file named as "extension.zip"
- Push the latest extension.zip to lambda layer
- Update the lambda function with the latest version of the lambda layer published in the previous step

# Testing
* To invoke the function, running the following command using AWS CLI

 ```
  java-example-extension$ aws lambda invoke \
  --function-name "${LAMBDA_FUNCTION}" \
  --payload '{"payload": "hello"}' /tmp/invoke-result \
  --cli-binary-format raw-in-base64-out --log-type Tail
  ```                  
