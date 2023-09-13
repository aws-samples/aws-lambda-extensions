# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import os
from re import T
import requests
import json

TELEMETRY_API_URL = "http://{0}/2022-07-01/telemetry".format(os.getenv("AWS_LAMBDA_RUNTIME_API"))
LAMBDA_EXTENSION_IDENTIFIER_HEADER_KEY = "Lambda-Extension-Identifier"

TIMEOUT_MS = 1000; # Maximum time (in milliseconds) that a batch is buffered.
MAX_BYTES = 256*1024; # Maximum size in bytes that the logs are buffered in memory.
MAX_ITEMS = 10000; # Maximum number of events that are buffered in memory.

def subscibe_listener(extension_id, listener_url):
    print ("[telemetry_api_client.subscibe_listener] Subscribing Extension to receive telemetry data. ExtenionsId: {0}, listener url: {1}, telemetry api url: {2}".format(extension_id, listener_url, TELEMETRY_API_URL))
    
    try:
        subscription_request_body = {
            "schemaVersion": "2022-07-01",
            "destination": {
                "protocol": "HTTP",
                "URI": listener_url,
            },
            "types": ["platform", "function", "extension"],
            "buffering": {
                "timeoutMs": TIMEOUT_MS,
                "maxBytes": MAX_BYTES,
                "maxItems": MAX_ITEMS
            }
        };

        subscription_request_headers = {
            "Content-Type": "application/json",
            LAMBDA_EXTENSION_IDENTIFIER_HEADER_KEY: extension_id,
        }

        response = requests.put(
            TELEMETRY_API_URL, 
            data = json.dumps(subscription_request_body),
            headers= subscription_request_headers
        )

        if response.status_code == 200:
            print("[telemetry_api_client.subscibe_listener] Extension successfully subscribed to telemetry api", response.text, flush=True)
        elif response.status_code == 202:
            print("[telemetry_api_client.subscibe_listener] Telemetry API not supported. Are you running the extension locally?", flush=True)
        else:
            print("[telemetry_api_client.subscibe_listener] Subsciption to telmetry API failed. ", "status code: ", response.status_code, "response text: ", response.text, flush=True)
        return extension_id

    except Exception as e:
        print("Error registering extension.", e, flush=True)
        raise Exception("Error setting AWS_LAMBDA_RUNTIME_API", e)



