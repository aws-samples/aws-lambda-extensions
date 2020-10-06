# Example Wrapper Script in Ruby
The provided code sample demonstrates how to get a wrapper script written in Ruby up and running.

There are two components to this sample:
* `wrapper_script`: This is a Ruby executable script that customizes the runtime startup by inserting additional parameters to the runtime process startup.
* `lambda_function.rb`: This is a sample Ruby file that includes a Lambda function handler that validates and demonstrates the use of the extra arguments that were included by the wrapper script.

## Customize the wrapper script
You can modify or include other valid runtime parameters in the script as extra arguments.

```ruby
...
# the extra options we want to pass to the interpreter
extra_args=["--enable", "frozen-string-literal"]
...
```

## Modify that the script access permissions
You'll want to ensure that the script is executable by running the following command:

```bash
$ chmod +x wrapper_script
```

## Deploy a function to test is out
Create a Lambda function for the Ruby runtime that includes both the `wrapper_script` and `lambda_function.rb` using `lambda_function.lambda_handler` as the function handler.

Add an environment variable to your function's configuration with key `AWS_LAMBDA_EXEC_WRAPPER` and a value of `/var/task/wrapper_script` (if you've included the wrapper script alongside your function code).

## Invoke the function
Invoke the function using a test eventand you should see the wrapper script in action reflected in the functions logs and invocation.