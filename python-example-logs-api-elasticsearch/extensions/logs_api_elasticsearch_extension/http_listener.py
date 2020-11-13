# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import json
import sys
from http.server import BaseHTTPRequestHandler, HTTPServer
from threading import Event, Thread

""" Demonstrates code to set up an HTTP listener and receive log events
"""

RECEIVER_IP = "0.0.0.0"
RECEIVER_PORT = 4243

def http_server_init(queue):
    def handler(*args):
        LogsHandler(queue, *args)
    print(f"Initializing HTTP Server on {RECEIVER_IP}:{RECEIVER_PORT}")
    server = HTTPServer((RECEIVER_IP, RECEIVER_PORT), handler)

    # Ensure that the server thread is scheduled so that the server binds to the port
    # and starts to listening before subscribe for the logs and ask for the next event.
    started_event = Event()
    server_thread = Thread(target=serve, daemon=True, args=(started_event, server,))
    server_thread.start()
    rc = started_event.wait(timeout = 9)
    if rc is not True:
        raise Exception("server_thread has timedout before starting")


# Server implementation
class LogsHandler(BaseHTTPRequestHandler):
    def __init__(self, queue, *args):
        self.queue = queue
        BaseHTTPRequestHandler.__init__(self, *args)

    def do_POST(self):
        try:
            cl = self.headers.get("Content-Length")
            if cl:
                data_len = int(cl)
            else:
                data_len = 0
            content = self.rfile.read(data_len)
            self.send_response(200)
            self.end_headers()
            batch = json.loads(content.decode("utf-8"))
            self.queue.put(batch)

        except Exception as e:
            print(f"Error processing message: {e}")

# Server thread
def serve(started_event, server):
    # Notify that this thread is up and running
    started_event.set()
    try:
        print(f"Serving HTTP Server on {RECEIVER_IP}:{RECEIVER_PORT}")
        server.serve_forever()
    except:
        print(f"Error in HTTP server {sys.exc_info()[0]}")
    finally:
        if server:
            server.shutdown()
