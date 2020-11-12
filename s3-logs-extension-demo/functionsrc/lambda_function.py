# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import json
import os

def lambda_handler(event, context):
    print(f"Function: Logging something which logging extension will send to S3")
    return {
        'statusCode': 200,
        'body': json.dumps('Hello from Lambda!')
    }
