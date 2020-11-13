# Example Logs API Extension in Python for Elasticsearch
This extension builds on the python-example-logs-api-extension and adds a library for communicating with Elasticsearch, wherein the LogsAPIHTTPExtension.run_forever() method is extended to send items in the received patch to Elasticsearch. It assumes those log lines are JSON objects or JSON serialized strings.
