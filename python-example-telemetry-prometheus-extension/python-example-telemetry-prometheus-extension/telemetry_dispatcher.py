# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import os
import uuid
import boto3
import json
from prometheus_client import CollectorRegistry, Summary, Counter, push_to_gateway, delete_from_gateway
from prometheus_client.exposition import basic_auth_handler

# Dispatch metrics when the available batch size reaches this value 
DISPATCH_MIN_BATCH_SIZE = int(os.getenv("DISPATCH_MIN_BATCH_SIZE", 5))

# Grab the function name to use as a label on the metrics
FUNCTION_NAME = os.getenv("AWS_LAMBDA_FUNCTION_NAME")

# Generate a UUID to differentiate events from different lambda containers
UUID = str(uuid.uuid4())

# The URL of the PushGateway
GATEWAY=os.getenv("GATEWAY")

# A CollectorRegistry instance is necessary for pushing to the PushGateway
registry = CollectorRegistry()

# To secure the PushGateway, this example follows these recommended security practices:
#  - Authentication is enabled on the PushGateway, using the "basic_auth_users" configuration parameter
#  - Traffic to the PushGateway is encrypted using TLS
#  - The Lambda is VPC enabled and the PushGateway is in the same VPC

# Create a client for the Secrets Manager service for reading credentials for the PushGateway
# Requires the "secretsmanager:GetSecretValue" permissions on the Lambda Role
sm_client = boto3.client('secretsmanager')

# Retrieve the credentials for the PushGateway from Secrets Manager
secret_response = sm_client.get_secret_value(SecretId=os.getenv('AUTH_SECRET'))
secret = json.loads(secret_response['SecretString'])

# Using http basic auth to authenticate with the PushGateway
def gateway_auth_handler(url, method, timeout, headers, data):
    username = secret['username']
    password = secret['password']
    return basic_auth_handler(url, method, timeout, headers, data, username, password)

# Define a set of standard labels for each of the Summary metrics
metric_labels = ["request_id", "instance", "function"]

# Create Summary metrics for each of the metrics in the "platform.report" event type
report_metrics = {
    "durationMs": Summary('durationMs', 'Duration in ms', registry=registry, labelnames=metric_labels),
    "billedDurationMs": Summary('billedDurationMs', 'Billed Duration', registry=registry, labelnames=metric_labels),
    "memorySizeMB": Summary('memorySizeMB', 'Memory Size MB', registry=registry, labelnames=metric_labels),
    "maxMemoryUsedMB": Summary('maxMemoryUsedMB', 'Max Memory Used MB', registry=registry, labelnames=metric_labels),
    "initDurationMs": Summary('initDurationMs', 'Init Duration MS', registry=registry, labelnames=metric_labels),
}

# Define a Counter to count the total number of events logged
counter = Counter('events', 'Number of events', registry=registry)

# Define a map that will contain counters of each of the different types of events
event_types = {}

# Dispatch a batch of metrics to prometheus
def dispatch_to_prometheus(batch):
    for metric in batch:
        # Increment the count of all events
        counter.inc()

        # Get the event time and increment the associated Counter for that event type
        metric_type = metric['type']
        metric_name = metric_type.replace(".","_")

        if not metric_type in event_types:
            # If this is the first time we've seen this event type, create a new Counter
            event_types[metric_type] = Counter(metric_name, f'Number of {metric_type}', registry=registry, labelnames=["instance", "function"]).labels(instance=UUID, function=FUNCTION_NAME)
        event_types[metric_type].inc()

        # For the platform.report type, record the Summary data for each metric
        if metric_type == "platform.report":
            record = metric['record']
            for summary_name, summary_metric in report_metrics.items():
                if summary_name in record["metrics"]:
                    sum_obj = summary_metric.labels(request_id=record['requestId'], instance=UUID, function=FUNCTION_NAME)
                    sum_obj.observe(record["metrics"][summary_name])

    # Push all metrics to the PushGateway
    push_to_gateway(GATEWAY, job=FUNCTION_NAME, registry=registry, handler=gateway_auth_handler, grouping_key={"id":UUID})

# When a Shutdown event is received, trigger the PushGateway to delete the metrics (otherwise Prometheus will continue receiving the same metrics)
def finalize_telemetry():
    print ("[telementry_dispatcher] Finalizing telemetry data", flush=True)
    delete_from_gateway(GATEWAY, job=FUNCTION_NAME, grouping_key={"id":UUID})

# Get the next batch and send to the Prometheus PushGateway
def dispatch_telmetery(queue, force):
    if ((not queue.empty()) and (force or queue.qsize() >= DISPATCH_MIN_BATCH_SIZE)):
        print ("[telementry_dispatcher] Dispatch telemetry data", flush=True)

        batch = []
        while (not queue.empty()):
            batch.extend(queue.get_nowait())

        dispatch_to_prometheus(batch)
