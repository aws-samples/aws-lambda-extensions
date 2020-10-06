#!/usr/bin/env python3
# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import boto3
import datetime
import json
import os
import requests
import signal
import sys
from pathlib import Path
from requests_aws4auth import AWS4Auth


# global variables
# extension name has to match the file's parent directory name)
LAMBDA_EXTENSION_NAME = Path(__file__).parent.name


# global variables: runtime
AWS_LAMBDA_RUNTIME_API = os.environ['AWS_LAMBDA_RUNTIME_API']
AWS_LAMBDA_FUNCTION_NAME = os.environ['AWS_LAMBDA_FUNCTION_NAME']
AWS_LAMBDA_FUNCTION_VERSION = os.environ['AWS_LAMBDA_FUNCTION_VERSION']
AWS_REGION = os.environ['AWS_REGION']


# global variables: elasticsearch
# initialize and set later from environment variables
ES_ENDPOINT = "localhost"
ES_INDEX = "test"


# custom elasticsearch code
# def get_awsauth():
#     # get credentials, can be done via iam role, here shown using a service account in cognito
#     service = "es"
#     credentials = boto3.Session().get_credentials()
#     awsauth = AWS4Auth(credentials.access_key, credentials.secret_key, AWS_REGION, service, session_token=credentials.token)
#     return awsauth


def send_elasticsearch(payload):
    # optional code for signed elasticsearch requests, awsauth should be done outside the loop
    # aws4auth = get_awsauth()
    url = f"https://{ES_ENDPOINT}/{ES_INDEX}/_doc"
    print(f"[{LAMBDA_EXTENSION_NAME}] Attempting POST to: {url}", flush=True)
    try:
        # if using self-signed certificates for dev/test, set verify=False
        # response = requests.post(endpoint, auth=aws4auth json=payload, verify=True)
        response = requests.post(url, json=payload, verify=True)
        print(f"[{LAMBDA_EXTENSION_NAME}] Response: {response.text}", flush=True)
    except requests.exceptions.ConnectionError as e:
        # TODO: handle this with exponential backoff and circuit breaker
        print(f"[{LAMBDA_EXTENSION_NAME}] ConnectionrError: {e}", flush=True)
        sys.exit(1)


# custom extension code
def execute_custom_processing(payload):
    # perform custom per-event processing here
    print(f"[{LAMBDA_EXTENSION_NAME}] Sending payload: {json.dumps(payload)}", flush=True)
    send_elasticsearch(payload)


# boiler plate code
def handle_signal(signal, frame):
    # if needed pass this signal down to child processes
    print(f"[{LAMBDA_EXTENSION_NAME}] Received signal={signal}. Exiting.", flush=True)
    sys.exit(0)


def register_extension():
    global ES_ENDPOINT, ES_INDEX
    print(f"[{LAMBDA_EXTENSION_NAME}] Registering...", flush=True)
    headers = {
        'Lambda-Extension-Name': LAMBDA_EXTENSION_NAME,
    }
    payload = {
        'events': [
            'INVOKE',
            'SHUTDOWN'
        ],
    }
    response = requests.post(
        url=f"http://{AWS_LAMBDA_RUNTIME_API}/2020-01-01/extension/register",
        json=payload,
        headers=headers
    )
    ext_id = response.headers['Lambda-Extension-Identifier']
    ES_ENDPOINT = os.environ('ES_ENDPOINT')
    ES_INDEX = os.environ('ES_INDEX')
    print(f"[{LAMBDA_EXTENSION_NAME}] Registered with ID: {ext_id}", flush=True)
    print(f"[{LAMBDA_EXTENSION_NAME}] Elasticsearch endpoint: {ES_ENDPOINT}", flush=True)
    print(f"[{LAMBDA_EXTENSION_NAME}] Elasticsearch index: {ES_INDEX}", flush=True)

    return ext_id


def process_events(ext_id):
    headers = {
        'Lambda-Extension-Identifier': ext_id
    }
    while True:
        print(f"[{LAMBDA_EXTENSION_NAME}] Waiting for event...", flush=True)
        start = datetime.datetime.now().timestamp()*1000
        response = requests.get(
            url=f"http://{AWS_LAMBDA_RUNTIME_API}/2020-01-01/extension/event/next",
            headers=headers,
            timeout=None
        )
        end = datetime.datetime.now().timestamp()*1000
        event = json.loads(response.text)
        if event['eventType'] == 'SHUTDOWN':
            print(f"[{LAMBDA_EXTENSION_NAME}] Received SHUTDOWN event. Exiting.", flush=True)
            sys.exit(0)
        else:
            payload = {
                "functionName": AWS_LAMBDA_FUNCTION_NAME,
                "functionVersion": AWS_LAMBDA_FUNCTION_VERSION,
                "requestId": event["requestId"],
                "waitDuration": end-start
            }
            execute_custom_processing(payload)


def main():
    # handle signals
    signal.signal(signal.SIGINT, handle_signal)
    signal.signal(signal.SIGTERM, handle_signal)

    # execute extensions logic
    extension_id = register_extension()
    process_events(extension_id)


if __name__ == "__main__":
    main()
