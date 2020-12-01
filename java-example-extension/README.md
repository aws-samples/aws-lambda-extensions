# Example Extension in Java 11

The project source includes function code and supporting resources:

- `src/main` - A Java extension.
- `pom.xml` - A Maven build file.
- `run.sh` - Shell script that will build the extension, create zip file, publish the lambda layer and update the lambda function with the latest version of this extension 

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

# Deploy

To deploy the extension, run `run.sh <<extension-name>> <<function-name>>` with the extension name and function name as parameters.

    java-example-extension$ ./run.sh java-extension blank-java   

This script uses AWS CLI to perform the following:
- Build the lambda extension
- Creates deployment zip file "extension.zip"
- Push the latest artifact in the form of zip as lambda layer
- Update the lambda function with the latest version of the lambda layer published

# Test
To invoke the function, running the following command using AWS CLI.

    blank-java$ aws lambda invoke \
                  --function-name "blank-java" \
                  --payload '{"payload": "hello"}' /tmp/invoke-result \
                  --cli-binary-format raw-in-base64-out \
                  --log-type Tail                  