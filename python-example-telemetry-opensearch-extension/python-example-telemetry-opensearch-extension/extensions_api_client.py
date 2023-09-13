# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import os
import sys
import requests
import json


LAMBDA_EXTENSION_NAME_HEADER_KEY = "Lambda-Extension-Name"
LAMBDA_EXTENSION_IDENTIFIER_HEADER_KEY = "Lambda-Extension-Identifier"
REGISTRATION_REQUEST_BASE_URL = "http://{0}/2020-01-01/extension".format(os.getenv("AWS_LAMBDA_RUNTIME_API"))

def register_extension(extension_name):
    print ("[extension_api_client.register_extension] Registering Extension using {0}".format(REGISTRATION_REQUEST_BASE_URL))
    
    try:
        registration_request_body =  {
            "events": 
            [
                "INVOKE", "SHUTDOWN"
            ]
        }
        registration_request_header = {
            "Content-Type": "application/json",
            LAMBDA_EXTENSION_NAME_HEADER_KEY: extension_name,
        }

        response = requests.post(
            "{0}/register".format(REGISTRATION_REQUEST_BASE_URL), 
            data = json.dumps(registration_request_body),
            headers= registration_request_header
        )

        if response.status_code == 200:
            extension_id = response.headers[LAMBDA_EXTENSION_IDENTIFIER_HEADER_KEY]
            print("[extension_api_client.register_extension] Registration success with extensionId {0}".format(extension_id), flush=True)
        else:
            print("[extension_api_client.register_extension] Error Registering extension: ", response.text, flush=True)
            # Fail the extension
            sys.exit(1)
        
        return extension_id

    except Exception as e:
        print("[extension_api_client.register_extension] Error registering extension: ",e, flush=True)
        raise Exception("Error setting AWS_LAMBDA_RUNTIME_API", e)


def next(extension_id):
    try:
        next_event_request_header = {
            "Content-Type": "application/json",
            LAMBDA_EXTENSION_IDENTIFIER_HEADER_KEY: extension_id,
        }

        response = requests.get(
            "{0}/event/next".format(REGISTRATION_REQUEST_BASE_URL), 
            headers= next_event_request_header
        )

        if response.status_code != 200:
            print("[extension_api_client.next] Failed receiving next event ", response.status_code, response.text, flush=True)
            #Fail extension with non-zero exit code
            sys.exit(1)

        event_data =  response.json()
        return event_data

    except Exception as e:
        print("[extension_api_client.next] Error registering extension.", e, flush=True)
        raise Exception("Error setting AWS_LAMBDA_RUNTIME_API", e)
