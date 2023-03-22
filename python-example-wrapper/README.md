# Example Wrapper Script in Python
The provided code sample demonstrates how to get a wrapper script written in Python up and running.

There are two components to this sample:
* `wrapper_script`: This is a Python executable script that customizes the runtime startup by inserting additional parameters to the runtime process startup.
* `lambda_function.py`: This is a sample Python file that includes a Lambda function handler that validates and demonstrates the use of the extra arguments that were included by the wrapper script.

## Customize the wrapper script
You can modify or include other valid runtime parameters in the script as extra arguments.

```python
...
# the extra options we want to pass to the interpreter
extra_args = ["-X", "importtime"]
...
```

## Modify the script access permissions
You'll want to ensure that the script is executable by running the following command:

```bash
$ chmod +x wrapper_script
```

## Deploy a function to test your script
Create a Lambda function for the Python runtime that includes both the `wrapper_script` and `lambda_function.py` using `lambda_function.lambda_handler` as the function handler.

Add an environment variable to your function's configuration with key `AWS_LAMBDA_EXEC_WRAPPER` and a value of `/var/task/wrapper_script` (if you've included the wrapper script alongside your function code).

## Invoke the function
Invoke the function using a test eventand you should see the wrapper script in action reflected in the functions logs and invocation.
