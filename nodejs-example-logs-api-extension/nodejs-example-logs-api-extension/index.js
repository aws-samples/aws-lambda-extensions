#!/usr/bin/env node
const { register, next } = require('./extensions-api');
const { subscribe } = require('./logs-api');
const { listen } = require('./http-listener');
const https = require('https');

/**

Note: 

- This is a simple example extension to help you start investigating the Lambda Runtime Logs API.
- It sends the Logs captured at runtime to elastic search without adding any latency as the extensions processing happens async in nature
- Because of the asynchronous nature of the system, it is possible that logs for one invoke are
processed during the next invoke slice. Likewise, it is possible that logs for the last invoke are processed during
the SHUTDOWN event.

*/

const EventType = {
  INVOKE: 'INVOKE',
  SHUTDOWN: 'SHUTDOWN',
};

function handleShutdown(event) {
  console.log('shutdown', { event });
  process.exit(0);
}

function handleInvoke(event) {
  console.log('invoke');
}

const LOCAL_DEBUGGING_IP = '0.0.0.0';
const RECEIVER_NAME = 'sandbox';

async function receiverAddress() {
  return process.env.AWS_SAM_LOCAL === 'true' ? LOCAL_DEBUGGING_IP : RECEIVER_NAME;
}

const ELASTICSEARCH_ENDPOINT = process.env.ELASTICSEARCH_ADD_DOC_ENDPOINT || `https://${process.env.ES_ENDPOINT}/my-index/_doc`; // Replace with your Elasticsearch insertion endpoint along with <index>/_doc

const RECEIVER_PORT = 4243;
const TIMEOUT_MS = 1000; // Maximum time (in milliseconds) that a batch is buffered.
const MAX_BYTES = 262144; // Maximum size in bytes that the logs are buffered in memory.
const MAX_ITEMS = 10000; // Maximum number of events that are buffered in memory.

const SUBSCRIPTION_BODY = {
  destination: {
      protocol: 'HTTP',
      URI: `http://${RECEIVER_NAME}:${RECEIVER_PORT}`,
  },
  types: ['platform', 'function'],
  buffering: {
      timeoutMs: TIMEOUT_MS,
      maxBytes: MAX_BYTES,
      maxItems: MAX_ITEMS,
  },
};

(async function main() {
    process.on('SIGINT', () => handleShutdown('SIGINT'));
    process.on('SIGTERM', () => handleShutdown('SIGTERM'));

    console.log('hello from logs api extension');

    console.log('register');
    const extensionId = await register();
    console.log('extensionId', extensionId);

    console.log('starting listener');
    // listen returns `logsQueue`, a mutable array that collects logs received from Logs API
    const { logsQueue, server } = listen(await receiverAddress(), RECEIVER_PORT);

    console.log('subscribing listener');
    // subscribing listener to the Logs API
    await subscribe(extensionId, SUBSCRIPTION_BODY, server);

    // function for processing collected logs
    function sendLogsToElasticsearch() {
        console.log(`collected ${logsQueue.length} log objects`);
        logsQueue.forEach(log => {
            try {
                const postData = JSON.stringify(log);
                console.log('postdata DEBUG:', postData);
                const options = {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Content-Length': Buffer.byteLength(postData),
                    },
                };

                const req = https.request(ELASTICSEARCH_ENDPOINT, options, (res) => {
                    console.log(res);
                    console.log(`Status code: ${res?.statusCode}`);
                });               

                req.on('error', (error) => {
                    console.error(`[${this.agent_name}] Error: ${error}`, flush = true);
                });

                req.write(postData);
                req.end();
                console.log('DEBUG: log processing & es exporting complete')

            } catch (error) {
                console.error(`[${this.agent_name}] Error: ${error}`, flush = true);
            }

        });
    }

    // execute extensions logic to process & export Logs [events captured in the queue]
    while (true) {
        console.log('next');
        const event = await next(extensionId);

        switch (event.eventType) {
            case EventType.SHUTDOWN:
                sendLogsToElasticsearch(); // upload remaining logs, during shutdown event
                handleShutdown(event);
                break;
            case EventType.INVOKE:
                handleInvoke(event);
                sendLogsToElasticsearch(); // send queued logs to Elasticsearch, during invoke event
                break;
            default:
                throw new Error('unknown event: ' + event.eventType);
        }
    }
})();
