# Example Extension in Python
The provided code sample demonstrates how to get a basic extension written in Python 3 up and running.

> Note: This extension requires the Python 3 runtime to be present in the Lambda execution environment of your function.

There are two components to this sample:
* `extensions/`: This sub-directory should be extracted to /opt/extensions where the Lambda platform will scan for executables to launch extensions
* `python-example-extension/`: This sub-directory should be extracted to /opt/python-example-extension which is referenced by the `extensions/python-example-extension` executable and includes a Python executable along with all of its necessary dependencies.

## Prep Python Dependencies
Install the extension dependencies locally, which will be mounted along with the extension code.

```bash
$ cd python-example-extension
$ chmod +x extension.py
$ pip3 install -r requirements.txt -t .
$ cd ..
```

## Layer Setup Process
The extensions .zip file should contain a root directory called `extensions/`, where the extension executables are located and another root directory called `python-example-extension/`, where the core logic of the extension  and its dependencies are located.

Creating zip package for the extension:
```bash
$ chmod +x extensions/python-example-extension
$ zip -r extension.zip .
```

Ensure that you have aws-cli v2 for the commands below.
Publish a new layer using the `extension.zip`. The output of the following command should provide you a layer arn.
```bash
aws lambda publish-layer-version \
 --layer-name "python-example-extension" \
 --region <use your region> \
 --zip-file  "fileb://extension.zip"
```
Note the LayerVersionArn that is produced in the output.
eg. `"LayerVersionArn": "arn:aws:lambda:<region>:123456789012:layer:<layerName>:1"`

Add the newly created layer version to a Python 3.8 runtime Lambda function.


## Function Invocation and Extension Execution

When invoking the function, you should now see log messages from the example extension similar to the following:
```
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    EXTENSION Name: python-example-extension State: Ready Events: [INVOKE,SHUTDOWN]
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    START RequestId: 9ca08945-de9b-46ec-adc6-3fe9ef0d2e8d Version: $LATEST
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    python-example-extension launching extension
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [python-example-extension] Registering...
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [python-example-extension] Registered with ID: 6ec8756c-4830-458b-9dda-156e5dda1cc1
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [python-example-extension] Waiting for event...
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [python-example-extension] Received event: {"eventType": "INVOKE", "deadlineMs": 1596217195507, "requestId": "c6702194-cf88-4407-9040-4d524ce0af3b", "invokedFunctionArn": "", "tracing": {"type": "X-Amzn-Trace-Id", "value": ""}}
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    [python-example-extension] Waiting for event...
    ...
    ...
    Function logs...
    ...
    ...
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    END RequestId: 9ca08945-de9b-46ec-adc6-3fe9ef0d2e8d
    XXXX-XX-XXTXX:XX:XX.XXX-XX:XX    REPORT RequestId: 9ca08945-de9b-46ec-adc6-3fe9ef0d2e8d Duration: 80.36 ms Billed Duration: 100 ms Memory Size: 128 MB Max Memory Used: 67 MB Init Duration: 297.83 ms
```