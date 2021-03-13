# Example Logs API Extension in Python

The provided code sample demonstrates how to get a basic Logs API extension written in Python 3 up and running.

> Note: This extension requires the Python 3 runtime to be present in the Lambda execution environment of your function. This example code is not production ready. Use it with your own discretion after testing thoroughly.

In this example, we start by developing a simple extension and then add the ability to read logs from the Logs API. For more details on building an extension, please read the Extension API Developer Guide.

When the Lambda service sets up the execution environment, it runs the extension (logs_api_http_extension.py). This extension first registers as an extension and then subscribes to the Logs API to receive the logs via HTTP protocol. It starts an HTTP listener which receives the logs and processes them.

## Layer Setup Process

The extensions .zip file should contain a root directory called `extensions/`, where the extension executables are located and another root directory called `python-example-extension/`, where the core logic of the extension  and its dependencies are located.

Creating zip package for the extension:

```bash
cd python-example-logs-api-extension
chmod +x extensions/logs_api_http_extension.py
zip -r extension.zip ./extensions
```

Publish a new layer using the `extension.zip`. The output of the following command should provide you a layer arn.

```bash
aws lambda publish-layer-version \
 --layer-name "python-example-logs-api-extension" \
 --region <use your region> \
 --zip-file  "fileb://extension.zip"
```

Note the LayerVersionArn that is produced in the output.
e.g. `"LayerVersionArn": "arn:aws:lambda:<region>:123456789012:layer:python-example-logs-api-extension:1"`

Add the newly created layer version to a Python 3.8 runtime Lambda function.

```bash
aws lambda update-function-configuration --region <use your region> --function-name <your function name> --layers <LayerVersionArn from previous step>
```

## Function Invocation and Extension Execution

After invoking the function and receiving the shutdown event, you should now see log messages from the example extension written back to the extension logs. For example:

```
...
BATCH RECEIVED: [{'time': '2020-11-12T00:30:48.883Z', 'type': 'platform.start', 'record': {'requestId': '49ae0e5f-bc60-4521-81e3-6e41d6bcb55c', 'version': '$LATEST'}}, {'time': '2020-11-12T00:30:48.993Z', 'type': 'platform.logsSubscription', 'record': {'name': 'logs_api_http_extension.py', 'state': 'Subscribed', 'types': ['platform', 'function']}}, {'time': '2020-11-12T00:30:48.993Z', 'type': 'platform.extension', 'record': {'name': 'logs_api_http_extension.py', 'state': 'Ready', 'events': ['INVOKE', 'SHUTDOWN']}}, {'time': '2020-11-12T00:30:49.017Z', 'type': 'platform.end', 'record': {'requestId': '49ae0e5f-bc60-4521-81e3-6e41d6bcb55c'}}, {'time': '2020-11-12T00:30:49.017Z', 'type': 'platform.report', 'record': {'requestId': '49ae0e5f-bc60-4521-81e3-6e41d6bcb55c', 'metrics': {'durationMs': 15.74, 'billedDurationMs': 100, 'memorySizeMB': 128, 'maxMemoryUsedMB': 62, 'initDurationMs': 226.3}}}]
...
```
