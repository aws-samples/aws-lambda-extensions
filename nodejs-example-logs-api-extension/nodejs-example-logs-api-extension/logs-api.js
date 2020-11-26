const fetch = require('node-fetch');

const baseUrl = `http://${process.env.AWS_LAMBDA_RUNTIME_API}/2020-08-15/logs`;

async function subscribe(extensionId, subscriptionBody) {
    const res = await fetch(baseUrl, {
        method: 'put',
        body: JSON.stringify(subscriptionBody),
        headers: {
            'Content-Type': 'application/json',
            'Lambda-Extension-Identifier': extensionId,
        }
    });

    if (!res.ok) {
        console.error('logs subscription failed', await res.text());
    } else {
        console.error('logs subscription ok', await res.text());
    }
}

module.exports = {
    subscribe,
};
