# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import json
import sys
from http.server import BaseHTTPRequestHandler, HTTPServer
from threading import Event, Thread
import socket

# Demonstrates code to set up an HTTP listener and receive log events

RECEIVER_NAME = "sandbox"
RECEIVER_IP = "0.0.0.0"
RECEIVER_PORT = 4243

def listenerAddress():
    try:
        socket.gethostbyname(RECEIVER_NAME)
        return RECEIVER_NAME
    except:
        return RECEIVER_IP

def http_server_init(queue):
    def handler(*args):
        LogsHandler(queue, *args)
    
    listenerName = listenerAddress()
    server = HTTPServer((listenerName, RECEIVER_PORT), handler)

    # Ensure that the server thread is scheduled so that the server binds to the port
    # and starts to listening before subscribing to the LogsAPI and asking for the next event.
    started_event = Event()
    server_thread = Thread(target=serve, daemon=True, args=(started_event, server,listenerName,))
    server_thread.start()
    rc = started_event.wait(timeout = 9)
    if rc is not True:
        raise Exception("server_thread has timed out before starting")


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
def serve(started_event, server, listenerName):
    # Notify that this thread is up and running
    started_event.set()
    try:
        print(f"extension.http_listener: Running HTTP Server on {listenerName}:{RECEIVER_PORT}")
        server.serve_forever()
    except:
        print(f"extension.http_listener: Error in HTTP server {sys.exc_info()[0]}", flush=True)
    finally:
        if server:
            server.shutdown()
