const fetch = require('node-fetch');

const baseUrl = `http://${process.env.AWS_LAMBDA_RUNTIME_API}/2020-08-15/logs`;

async function subscribe(extensionId, subscriptionBody, server) {
    const res = await fetch(baseUrl, {
        method: 'put',
        body: JSON.stringify(subscriptionBody),
        headers: {
            'Content-Type': 'application/json',
            'Lambda-Extension-Identifier': extensionId,
        }
    });

    switch(res.status) {
        case 200:
            console.info('logs subscription ok: ', await res.text());
            break;
        case 202:
            console.warn('WARNING!!! Logs API is not supported! Is this extension running in a local sandbox?');
            break;
        default:
            console.error('logs subscription failed: ', await res.text());
            break;
    }
}

module.exports = {
    subscribe,
};

