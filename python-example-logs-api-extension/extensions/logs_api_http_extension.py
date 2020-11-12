#!/bin/sh
''''exec python -u -- "$0" ${1+"$@"} # '''
# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0
import os

from logs_api_http_extension.http_listener import http_server_init, RECEIVER_PORT
from logs_api_http_extension.logs_api_client import LogsAPIClient
from logs_api_http_extension.extensions_api_client import ExtensionsAPIClient
from queue import Queue

"""Here is the sample extension code that stitches all of this together.
    - The extension will run two threads. The "main" thread, will register to ExtensionAPI and process its invoke and
    shutdown events (see next call). The second "listener" thread will listen for HTTP Post events that deliver log batches.
    - The "listener" thread will place every log batch it receives in a synchronized queue; during each execution slice,
    the "main" thread will make sure to process any event in the queue before returning control by invoking next again.
    - Note that because of the asynchronous nature of the system, it is possible that logs for one invoke are
    processed during the next invoke slice. Likewise, it is possible that logs for the last invoke are processed during
    the SHUTDOWN event.

Note: 

1.  This is a simple example extension to make you help start investigating the Lambda Runtime Logs API.
    This code is not production ready, and it has never intended to be. Use it with your own discretion after you tested
    it thoroughly.  

2.  The extension code is starting with a shebang this is to bring Python runtime to the execution environment.
    This works if the lambda function is a python3.x function therefore it brings python3.x runtime with itself.
    It may not work for python 2.7 or other runtimes. 
    The recommended best practice is to compile your extension into an executable binary and not rely on the runtime.
  
3.  This file needs to be executable, so make sure you add execute permission to the file 
    `chmod +x logs_api_http_extension.py`

"""

class LogsAPIHTTPExtension():
    def __init__(self, agent_name, registration_body, subscription_body):
        print(f"Initializing LogsAPIExternalExtension {agent_name}")
        self.agent_name = agent_name
        self.queue = Queue()
        self.logs_api_client = LogsAPIClient()
        self.extensions_api_client = ExtensionsAPIClient()

        # Register early so Runtime could start in parallel
        self.agent_id = self.extensions_api_client.register(self.agent_name, registration_body)

        # Start listening before Logs API registration
        http_server_init(self.queue)
        self.logs_api_client.subscribe(self.agent_id, subscription_body)

    def run_forever(self):
        print(f"Serving LogsAPIHTTPExternalExtension {self.agent_name}")
        while True:
            resp = self.extensions_api_client.next(self.agent_id)
            # Process the received batches if any.
            while not self.queue.empty():
                batch = self.queue.get_nowait()
                # This line logs the events received to CloudWatch.
                # Replace it to send logs to elsewhere.
                # If you've subscribed to extension logs, e.g. "types": ["platform", "function", "extension"],
                # you'll receive the logs of this extension back from Logs API.
                # And if you log it again with the line below, it will create a cycle since you receive it back again.
                # Use `extension` log type if you'll egress it to another endpoint,
                # or make sure you've implemented a protocol to handle this case.
                print(f"BATCH RECEIVED: {batch}")

# Register for the INVOKE events from the EXTENSIONS API
_REGISTRATION_BODY = {
    "events": ["INVOKE", "SHUTDOWN"],
}

# Subscribe to platform logs and receive them on ${local_ip}:4243 via HTTP protocol.

TIMEOUT_MS = 1000 # Maximum time (in milliseconds) that a batch would be buffered.
MAX_BYTES = 262144 # Maximum size in bytes that the logs would be buffered in memory.
MAX_ITEMS = 10000 # Maximum number of events that would be buffered in memory.

_SUBSCRIPTION_BODY = {
    "destination":{
        "protocol": "HTTP",
        "URI": f"http://sandbox:{RECEIVER_PORT}",
    },
    "types": ["platform", "function"],
    "buffering": {
        "timeoutMs": TIMEOUT_MS,
        "maxBytes": MAX_BYTES,
        "maxItems": MAX_ITEMS
    }
}

def main():
    print(f"Starting Extensions {_REGISTRATION_BODY} {_SUBSCRIPTION_BODY}")
    # Note: Agent name has to be file name to register as an external extension
    ext = LogsAPIHTTPExtension(os.path.basename(__file__), _REGISTRATION_BODY, _SUBSCRIPTION_BODY)
    ext.run_forever()

if __name__ == "__main__":
    main()
