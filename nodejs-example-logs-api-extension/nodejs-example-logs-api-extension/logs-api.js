const http = require('http');

const baseUrl = `http://${process.env.AWS_LAMBDA_RUNTIME_API}/2020-08-15/logs`;

async function subscribe(extensionId, subscriptionBody, server) {
    const requestBody = JSON.stringify(subscriptionBody);

    const options = {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Lambda-Extension-Identifier': extensionId,
        },
    };

    return new Promise((resolve, reject) => {
        const req = http.request(baseUrl, options, (res) => {
            let data = '';
            res.on('data', (chunk) => {
                data += chunk;
            });

            res.on('end', () => {
                switch (res.statusCode) {
                    case 200:
                        console.info('logs subscription ok: ', data);
                        resolve();
                        break;
                    case 202:
                        console.warn('WARNING!!! Logs API is not supported! Is this extension running in a local sandbox?');
                        resolve();
                        break;
                    default:
                        console.error('logs subscription failed: ', data);
                        reject(new Error('Logs subscription failed'));
                        break;
                }
            });
        });

        req.on('error', (error) => {
            console.error('logs subscription error:', error);
            reject(error);
        });

        req.write(requestBody);
        req.end();
    });
}

module.exports = {
    subscribe,
};
