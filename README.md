## AWS Lambda Extensions
### Sample projects

Extensions are a new way for tools to more easily integrate deeply into the Lambda execution environment to control and participate in Lambda’s lifecycle. 

You can use extensions to integrate your Lambda functions with your preferred monitoring, observability, security, and governance tools. You can choose from a broad set of tools provided by AWS Lambda partners or you can create your own Lambda extensions.

Extensions use the Extensions API, a new HTTP interface, to register for lifecycle events and get greater control during function initialization, invocation, and shutdown. They can also use environment variables to add options and tools to the runtime, or use wrapper scripts to customize the runtime startup behavior.

Note: an internal extension runs in the runtime process, and shares the same lifecycle as the runtime. An external extension runs as a separate process in the execution environment. The extension runs in parallel with the function's runtime. It is initialized before the function is invoked and continues to run after the function invocation is complete

For more information, see [Using AWS Lambda extensions](https://docs.aws.amazon.com/lambda/latest/dg/using-extensions.html).

In this repository you'll find a number of different sample projects and demos to help you get started with building your own extension. These include:

* [AWS AppConfig extension demo](awsappconfig-extension-demo/)
* [Custom runtime extension demo](custom-runtime-extension-demo/)
* [Extension in Go](go-example-extension/)
* [Extension in Python](python-example-extension/)
* [Extension in Node.js](nodejs-example-extension/)
* [Inter-process communication extension in Go](go-example-ipc-extension/)
* [Crash uploader extension in Go](go-example-crash-uploader-extension/)
* [ElasticSearch extension in Python](python-example-elasticsearch-extension/)
* [Lambda layer extension using SAM](go-example-extension-sam-layer/)
* [Wrapper script in Bash](bash-example-wrapper/)
* [Wrapper script in Python](python-example-wrapper/)
* [Wrapper script in Ruby](ruby-example-wrapper/)


## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## License

This library is licensed under the MIT-0 License. See the LICENSE file.

