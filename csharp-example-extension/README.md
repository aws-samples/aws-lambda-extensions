# Example Extension in C#.NET

## Project structure

Project files and folders:

- `dotnet-example-extension.sln` - Visual Studio solution file. This is an optional file which is not required for .NET CLI and VSCode editor.
- `dotnet-example-extension.csproj` - .NET Core project file, which is mandatory for .NET Core build process. It defines `<LangVersion>latest</LangVersion>`, so that the latest C# language features, like [async Main](https://docs.microsoft.com/en-us/dotnet/csharp/language-reference/proposals/csharp-7.1/async-main) could be used.
- `Program.cs` - main entry point for this extension.
- `ExtensionClient.cs` - Lambda Extension API client implementation.
- `ExtensionEvent.cs` - Event types enumerable, so that the rest of the code can work with enum values, rather than string constants.
- `extensions/dotnet-example-extension` - Bash script that must be deployed to `opt/extensions` folder as an executable file (see manual deployment steps below for details).

## Requirements

- [.NET Core SDK](https://dotnet.microsoft.com/download) - 3.1 or a later version with 3.1 targeting pack installed.
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

### Option 1: Use shell script to deploy the extension

- make sure that `csharp-example-extension` is your current folder and `run.sh` script has executable bit set - `ls -l run.sh` output should start with `-rwxr-xr-x` permissions mask.
- make `run.sh` executable with `chmod +x run.sh` if needed.
- assuming that demo C# Lambda function has already been deployed to the current AWS account and its name is `test-dotnet`, execute the following command to build the extension and deploy it to the current AWS account as a Lambda layer, named `dotnet-example-extension`:

```bash
./run.sh dotnet-example-extension test-dotnet
```

### Option 2: Step-by-step manual deployment

- Publish .NET Core project with `Release` configuration. This command (see below) will download all necessary NuGet packages, referenced by the project, build the binaries using `Release` configuration settings and publish the result to `bin/Release/netcoreapp3.1/publish` subfolder. Please, refer to [dotnet publish](https://docs.microsoft.com/en-us/dotnet/core/tools/dotnet-publish) documentation for details and additional command line options.

```bash
dotnet publish -c Release
```

- Make sure that `extensions/dotnet-example-extension` has executable bit set and set it if needed, otherwise Lambda initialization will fail with `PermissionDenied` error (see Troubleshooting section for details):

```bash
chmod +x extensions/dotnet-example-extension
```

- Change your current folder to the publish destination folder:

```bash
cd bin/Release/netcoreapp3.1/publish
```

- Move all publish results to `deploy.zip` archive recursively:

```bash
zip -rm ./deploy.zip *
```

- Define Lambda function and extension name variables, so that they can be easily referenced later. Make sure that `test-dotnet` Lambda function has already been published to the current AWS account:

```bash
EXTENSION_NAME="dotnet-example-extension"
LAMBDA_FUNCTION="test-dotnet"
```

- Publish extension archive as a new Lambda layer, limiting it to be compatible with .NET Core Lambda runtime only:

```bash
aws lambda publish-layer-version \
  --compatible-runtimes "dotnetcore3.1" \
  --layer-name "${EXTENSION_NAME}" \
  --zip-file "fileb://deploy.zip"
```

- Update `test-dotnet` Lambda function to use `dotnet-example-extension` layer

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

> EXTENSION Name: dotnet-example-extension  State: LaunchError  Events: []  Error Type: PermissionDenied

#### Resolution

Validate that execution bit has been properly set on the extension shell script.
For example `ls -la extensions` output should look like (notice `x` bit set):

> -rwxr-xr-x   1 user  group  289 Dec 22 15:16 dotnet-example-extension
