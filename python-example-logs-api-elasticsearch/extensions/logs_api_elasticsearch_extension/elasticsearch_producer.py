import json
import requests
import sys

class ElasticsearchProducer():
    def __init__(self, agent_name, endpoint, index):
        self.agent_name = agent_name
        self.endpoint = endpoint
        self.index = index

    def send(self, payload):
        # optional code for signed elasticsearch requests, awsauth should be done outside the loop
        # aws4auth = get_awsauth()
        url = f"https://{self.endpoint}/{self.index}/_doc"
        # print(f"[{self.agent_name}] Attempting POST to: {url}", flush=True)
        try:
            if isinstance(payload["record"], str):
                converted = payload
                converted["record"] = json.loads(payload["record"].replace("'",'"').rstrip())
            else:
                converted = payload
            # if using self-signed certificates for dev/test, set verify=False
            # response = requests.post(endpoint, auth=aws4auth json=payload, verify=True)
            response = requests.post(url, json=converted, verify=True)
            # print(f"[{self.agent_name}] Response: {response.text}", flush=True)
        except requests.exceptions.ConnectionError as e:
            # TODO: handle this with exponential backoff and circuit breaker
            print(f"[{self.agent_name}] ConnectionrError: {e}", flush=True)
            sys.exit(1)
