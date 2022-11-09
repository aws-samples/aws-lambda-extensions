// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

const fetch = require('node-fetch');
const {basename} = require('path');

const baseUrl = `http://${process.env.AWS_LAMBDA_RUNTIME_API}/2020-01-01/extension`;

async function register() {
    console.info('[extensions-api:register] Registering using baseUrl', baseUrl);
    const res = await fetch(`${baseUrl}/register`, {
        method: 'post',
        body: JSON.stringify({
            'events': [
                'INVOKE',
                'SHUTDOWN'
            ],
        }),
        headers: {
            'Content-Type': 'application/json',
            // The extension name must match the file name of the extension itself that's in /opt/extensions/
            'Lambda-Extension-Name': basename(__dirname),
        }
    });

    if (!res.ok) {
        console.error('[extensions-api:register] Registration failed:', await res.text());
    } else {
        const extensionId = res.headers.get('lambda-extension-identifier');
        console.info('[extensions-api:register] Registration success with extensionId', extensionId);
        return extensionId;
    }
}

async function next(extensionId) {
    console.info('[extensions-api:next] Waiting for next event');
    const res = await fetch(`${baseUrl}/event/next`, {
        method: 'get',
        headers: {
            'Content-Type': 'application/json',
            'Lambda-Extension-Identifier': extensionId,
        }
    });

    if (!res.ok) {
        console.error('[extensions-api:next] Failed receiving next event', await res.text());
        return null;
    } else {
        const event = await res.json();
        console.info('[extensions-api:next] Next event received');
        return event;
    }
}

module.exports = {
    register,
    next,
};