# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import os
import json
import boto3
from opensearchpy import OpenSearch, helpers

DISPATCH_MIN_BATCH_SIZE = int(os.getenv("DISPATCH_MIN_BATCH_SIZE", 5))

# Create a client for the Secrets Manager service for reading credentials for OpenSearch
# Requires the "secretsmanager:GetSecretValue" permissions on the Lambda Role
sm_client = boto3.client('secretsmanager')

# URL information for the OpenSearch server
url = os.getenv("URL")

# Read the authentication information from Secrets Manager
secret_response = sm_client.get_secret_value(SecretId=os.getenv('AUTH_SECRET'))
secret = json.loads(secret_response['SecretString'])
auth=(secret['username'], secret['password'])

# The indices for the various types of events
indices = {
    "platform": os.getenv("PLATFORM_INDEX"),
    "function": os.getenv("FUNCTION_INDEX"),
    "extension": os.getenv("EXTENSION_INDEX")
}

# Initialization, called when the extension first starts
def init_opensearch():
    # Initialize OpenSearch client
    client = OpenSearch(url, http_auth=auth)

    # Check for the telemetry index, and create if it doesn't exist
    # It needs to be created to ensure the "record" object is dynamic
    if not client.indices.exists(client.indices.exists(indices["platform"])):
        index_settings = {
                            "settings": {
                                "number_of_shards": 3
                            },
                            "mappings": {
                                "properties": {
                                "time": {"type": "date"},
                                "type": {"type": "keyword"},
                                "record": { "dynamic": True, "type": "object" }
                                }
                            }
                        }
        client.indices.create(index=indices["platform"], ignore=400, body=index_settings)
    return client

# Dispatch events to OpenSearch
def dispatch_to_opensearch(batch):
    # Each action specifies the destination index in the _index field, and the event is in the doc field
    actions = [{"_index": indices[doc['type'].split('.')[0]], "doc": doc} 
                    for doc in batch if doc['type'].split('.')[0] in indices.keys()]

    try:
        # Use the bulk helper to send a batch of events
        resp = helpers.bulk(client, actions)
    except helpers.errors.BulkIndexError as bulk_err:
        print(f'[telementry_dispatcher] Unable to submit telemetry data batch: {bulk_err}')

# Remove the next batch from the queue and dispatch to OpenSearch
def dispatch_telmetery(queue, force):
    if ((not queue.empty()) and (force or queue.qsize() >= DISPATCH_MIN_BATCH_SIZE)):
        print ("[telementry_dispatcher] Dispatch telemetry data")

        batch = []
        while (not queue.empty()):
            batch.extend(queue.get_nowait())

        dispatch_to_opensearch(batch)

# Initialize the OpenSearch client
client = init_opensearch()
