// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

const fetch = require('node-fetch');

const baseUrl = `http://${process.env.AWS_LAMBDA_RUNTIME_API}/2022-07-01/telemetry`;
const TIMEOUT_MS = 1000; // Maximum time (in milliseconds) that a batch is buffered.
const MAX_BYTES = 256 * 1024; // Maximum size in bytes that the logs are buffered in memory.
const MAX_ITEMS = 10000; // Maximum number of events that are buffered in memory.

async function subscribe(extensionId, listenerUri) {
    console.log('[telemetry-api:subscribe] Subscribing', { baseUrl, extensionId, listenerUri });

    const subscriptionBody = {
        schemaVersion: "2022-07-01",
        destination: {
            protocol: "HTTP",
            URI: listenerUri,
        },
        types: ['platform'],// 'function', 'extension'
        buffering: {
            timeoutMs: TIMEOUT_MS,
            maxBytes: MAX_BYTES,
            maxItems: MAX_ITEMS
        }
    };

    const res = await fetch(baseUrl, {
        method: 'put',
        body: JSON.stringify(subscriptionBody),
        headers: {
            'Content-Type': 'application/json',
            'Lambda-Extension-Identifier': extensionId,
        }
    });

    switch (res.status) {
        case 200:
            console.log('[telemetry-api:subscribe] Subscription success:', await res.text());
            break;
        case 202:
            console.warn('[telemetry-api:subscribe] Telemetry API not supported. Are you running the extension locally?');
            break;
        default:
            console.error('[telemetry-api:subscribe] Subscription failure:', await res.text());
            break;
    }
}

module.exports = {
    subscribe,
};