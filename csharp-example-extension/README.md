# Example Extension in C#.NET

The provided code sample demonstrates how to get a basic extension written in C# up and running.

## Project structure

Project files and folders:

- `csharp-example-extension.csproj` - .NET Core project file, which is mandatory for .NET Core build process. It defines `<LangVersion>latest</LangVersion>`, so that the latest C# language features, like [async Main](https://docs.microsoft.com/en-us/dotnet/csharp/language-reference/proposals/csharp-7.1/async-main) could be used.
- `Program.cs` - main entry point for this extension.
- `ExtensionClient.cs` - Lambda Extension API client implementation.
- `ExtensionEvent.cs` - Event types enumerable, so that the rest of the code can work with enum values, rather than string constants.
- `extensions/csharp-example-extension` - Bash script that must be deployed to `opt/extensions` folder as an executable file (see manual deployment steps below for details). This script will be used for .NET runtime dependent deployment.
- `extensions/csharp-example-extension-self-contained` - Bash script that must be deployed to `opt/extensions` folder as an executable file (see manual deployment steps below for details). This script will be used for self-contained deployment.

## Requirements

- [.NET Core SDK](https://dotnet.microsoft.com/download) - 5.0 or a later version with 3.1 targeting pack installed (if .NET Core 3.1 dependent deployment is used).
- [AWS CLI v2](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html) - version 2.1.14 or newer.
- Since this is an external .NET Core extension, we would require a lambda function with .NET Core 3.1 runtime published to test this extension. For more information on publishing a blank C# lambda function please [refer here](https://github.com/awsdocs/aws-lambda-developer-guide/tree/master/sample-apps/blank-csharp).
- Bash shell with `zip` command support. Linux and macOS operating systems have that shell available by default (although the macOS has recently switched to ZSH default shell, Bash is still available for Catalina and Big Sur anyway). Windows 10 users can install [Windows Subsystem for Linux](https://docs.microsoft.com/en-us/windows/wsl/install-win10) or just use Git Bash, which is included with [Git for Windows](https://gitforwindows.org/) installation package.

The following configuration setting must be set in [AWS CLI configuration file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) (`~/.aws/config`), so that AWS CLI v2 will use backward compatibility with the AWS CLI version 1 behavior where binary values must be passed literally:

```ini
cli_binary_format=raw-in-base64-out
```

## Setup

### Clone the repository

```bash
git clone https://github.com/aws-samples/aws-lambda-extensions.git
cd aws-lambda-extensions/csharp-example-extension
```

### Deployment options

`csharp-example-extension.csproj` has all necessary configuration for deploying this extension as a self-contained executable or a .NET runtime dependent extension.
Self-contained executable does not require any runtime to be pre-installed for Lambda function, thus it is compatible with any Lambda functions (.NET, Java, NodeJS, etc.). All necessary .NET runtime libraries are packaged together with the extension, thus the result package is much larger, than the .NET runtime dependent package (10MB+ vs 10KB+).

### Automated deployment: Use shell script to deploy the extension

- make sure that `csharp-example-extension` is your current folder and `run.sh` script has executable bit set - `ls -l run.sh` output should start with `-rwxr-xr-x` permissions mask.
- make `run.sh` executable with `chmod +x run.sh` if needed.
- assuming that demo C# Lambda function has already been deployed to the current AWS account and its name is `test-dotnet`, execute the following command to build the extension and deploy it to the current AWS account as a Lambda layer, named `csharp-example-extension`:

#### .NET Core 3.1 dependent extension

> IMPORTANT: You will be able to attach this extension to .NET Core 3.1 Lambda functions only!!!

```bash
./run.sh -e csharp-example-extension -f test-dotnet
```

#### Self-contained extension

This type of deployment is compatible with any Lambda function runtime.

```bash
./run.sh -e csharp-example-extension -f test-dotnet -s
```

### Manual deployment

- Make sure that all scripts in `extensions/*` folder has executable bit set and set it if needed, otherwise Lambda initialization will fail with `PermissionDenied` error (see Troubleshooting section for details):

```bash
chmod +x extensions/*
```

- Publish .NET Core project with `Release` configuration and targeting a specific framework. `csharp-example-extension.csproj` file contains all necessary configuration for building and packaging the result if needed. This command (see below) will download all necessary NuGet packages, referenced by the project, build the binaries using `Release` configuration settings and publish the result to `bin/publish` folder. Please, refer to [dotnet publish](https://docs.microsoft.com/en-us/dotnet/core/tools/dotnet-publish) documentation for details and additional command line options.

#### Build runtime dependent extension (compatible with .NET Core 3.1 Lambda only!!!)

```bash
dotnet publish -c Release -f netcoreapp3.1 -o bin/publish
```

#### Build self-contained extension using .NET Core 5.0 runtime

`-p:Platform=x64` switch will enable conditional build configuration in `csharp-example-extension.csproj` to package extension together with `linux-x64` runtime and wrap everything into a single bundle. This bundle doesn't require a custom shell script and will be deployed directly to `extensions` folder.

```bash
dotnet publish -c Release -f net5.0 -p:Platform=x64 -o bin/publish
```

- Change your current folder to the publish destination folder:

```bash
cd bin/publish
```

- Move all publish results to `deploy.zip` archive recursively:

```bash
zip -rm ./deploy.zip *
```

- Define Lambda function and extension name variables, so that they can be easily referenced later. Make sure that `test-dotnet` Lambda function has already been published to the current AWS account (it can be any other, non-.NET function im case of the self-contained extension):

```bash
EXTENSION_NAME="csharp-example-extension"
LAMBDA_FUNCTION="test-dotnet"
```

- Publish extension archive as a new Lambda layer:

```bash
aws lambda publish-layer-version \
  --layer-name "${EXTENSION_NAME}" \
  --zip-file "fileb://deploy.zip"
```

- Update `test-dotnet` Lambda function to use `csharp-example-extension` layer

```bash
aws lambda update-function-configuration \
  --function-name ${LAMBDA_FUNCTION} --layers $(aws lambda list-layer-versions --layer-name ${EXTENSION_NAME} \
  --max-items 1 --no-paginate --query 'LayerVersions[0].LayerVersionArn' \
  --output text)
```

## Testing

You can use AWS Console or AWS CLI to invoke your test function (`test-dotnet`):

```bash
aws lambda invoke \
  --function-name "${LAMBDA_FUNCTION}" \
  --payload '{"payload": "hello"}' /tmp/invoke-result \
  --cli-binary-format raw-in-base64-out --log-type Tail
```

All logs can be fund in `/aws/lambda/test-dotnet` Cloudwatch group - both Lambda extension and Lambda function log messages are reported to that group.

## Troubleshooting

### Execution bit not set on the launch script

#### Symptoms

> EXTENSION Name: csharp-example-extension  State: LaunchError  Events: []  Error Type: PermissionDenied

#### Resolution

Validate that execution bit has been properly set on the extension shell script.
For example `ls -la extensions` output should look like (notice `x` bit set):

> -rwxr-xr-x   1 user  group  289 Dec 22 15:16 csharp-example-extension

Self-contained deployments must make sure that `dotnet publish` output has proper executable flag set on the bundle file.
