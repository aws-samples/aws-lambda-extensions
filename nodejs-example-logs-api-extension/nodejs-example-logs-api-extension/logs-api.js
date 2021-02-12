const fetch = require('node-fetch');
const terminator = require('http-terminator');

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

    let terminate = false;
    switch(res.status) {
        case 200:
            console.info('logs subscription ok: ', await res.text());
            break;
        case 202:
            console.warn('WARNING!!! Logs API is not supported! Is this extension running in a local sandbox?');
            terminate = true;
            break;
        default:
            console.error('logs subscription failed: ', await res.text());
            terminate = true
            break;
    }

    if(terminate) {
        console.info('terminating http listener...');
        const httpTerminator = terminator.createHttpTerminator({server});
        await httpTerminator.terminate();
    }
}

module.exports = {
    subscribe,
};

