// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

const express = require('express');

const LISTENER_HOST = process.env.AWS_SAM_LOCAL === 'true' ? '0.0.0.0' : 'sandbox.localdomain';
const LISTENER_PORT = 4243;
const eventsQueue = [];

function start() {
    console.log('[telementry-listener:start] Starting a listener');
    const server = express();
    server.use(express.json({ limit: '512kb' }));

    // Logging or printing besides handling error cases below is not recommended 
    // if you have subscribed to receive extension logs. Otherwise, logging here will 
    // cause Telemetry API to send new entries for the printed lines which might create a loop
    server.post('/', (req, res) => {
        if (req.body.length && req.body.length > 0) {
            eventsQueue.push(...req.body);
        }
        console.log('[telementry-listener:post] received', req.body.length, 'total', eventsQueue.length);
        res.send('OK');
    });

    const listenerUrl = `http://${LISTENER_HOST}:${LISTENER_PORT}`;
    server.listen(LISTENER_PORT, LISTENER_HOST, () => {
        console.log(`[telemetry-listener:start] listening at ${listenerUrl}`);
    });
    return listenerUrl;
}

module.exports = {
    start,
    eventsQueue
};

