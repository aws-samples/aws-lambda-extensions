import os
import requests
import json

DISPATCH_POST_URI = os.getenv("DISPATCH_POST_URI")
DISPATCH_MIN_BATCH_SIZE = int(os.getenv("DISPATCH_MIN_BATCH_SIZE"));

def dispatch_telmetery(queue, force):
    while ((not queue.empty()) and (force or queue.qsize() >= DISPATCH_MIN_BATCH_SIZE)):
        print ("[telementry_dispatcher] Dispatch telemetry data")
        batch = queue.get_nowait()

        if DISPATCH_POST_URI is None:
            print ('[telementry_dispatcher:dispatch] dispatchPostUri not found. Discarding log events from the queue')
        else:
            # Modify the below line to dispatch/send the telemetry data to the desired choice of observability tool.
            response = requests.post(
                DISPATCH_POST_URI, 
                data = json.dumps(batch),
                headers= {'Content-Type': 'application/json'}
            )
            #print(f"BATCH RECEIVED: {batch}", flush=True)