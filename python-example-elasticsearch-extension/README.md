# Example Logs API Extension in Python for Elasticsearch
The provided code sample demonstrates how to get a basic Logs API extension for Elasticsearch written in Python 3 up and running.

> Note: This extension requires the Python 3 runtime to be present in the Lambda execution environment of your function. This example code is not production ready. Use it with your own discretion after testing thoroughly.

In this example, we start by developing a simple extension and then add the ability to read logs from the Logs API. For more details on building an extension, please read the Extension API Developer Guide.

When the Lambda service sets up the execution environment, it runs the extension (logs_api_elasticsearch_extension.py). This extension first registers as an extension and then subscribes to the Logs API to receive the logs via HTTP protocol. It starts an HTTP listener which receives the logs and processes them.

## Layer Setup Process
The extensions .zip file should contain a root directory called `extensions/`, where the extension executables are located. The dependencies for the extension (logs_api_elasticsearch_extension.py) are found in the logs_api_elasticsearch_extension directory.

Creating zip package for the extension:
```bash
$ cd python-example-elasticsearch-extension
$ chmod +x extensions/logs_api_elasticsearch_extension.py
$ zip -r extension.zip .
```

Publish a new layer using the `extension.zip`. The output of the following command should provides you a layer arn.
```bash
aws lambda publish-layer-version \
 --layer-name "python-example-elasticsearch-extension" \
 --region <use your region> \
 --zip-file  "fileb://extension.zip"
```
Note the LayerVersionArn that is produced in the output.
e.g. `"LayerVersionArn": "arn:aws:lambda:<region>:123456789012:layer:python-example-elasticsearch-extension:1"`

Add the newly created layer version to a Python 3.8 runtime Lambda function.
```bash
aws lambda update-function-configuration --region <use your region> --function-name <your function name> --layers <LayerVersionArn from previous step>
```
