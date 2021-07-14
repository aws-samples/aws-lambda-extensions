# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import json
import os

def lambda_handler(event, context):
    print(f"Inside Lambda function handler")
    return {
        'statusCode': 200,
        'body': json.dumps('Hello from Lambda!')
    }
