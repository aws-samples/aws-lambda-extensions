# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import json
import sys
import urllib.request

class ElasticsearchProducer():
    def __init__(self, agent_name, endpoint, index):
        self.agent_name = agent_name
        self.endpoint = endpoint
        self.index = index

    def send(self, payload):
        url = f"https://{self.endpoint}/{self.index}/_doc"
        try:
            if isinstance(payload["record"], str):
                converted = payload
                converted["record"] = json.loads(payload["record"].replace("'",'"').rstrip())
            else:
                converted = payload
            req = urllib.request.Request(url)
            req.method = "POST"
            req.add_header("Content-Type", "application/json")
            req.data = json.dumps(converted).encode("utf-8")
            resp = urllib.request.urlopen(req)
        except urllib.request.HTTPError as e:
            print(f"[{self.agent_name}] HTTPError: {e}", flush=True)
            sys.exit(1)
