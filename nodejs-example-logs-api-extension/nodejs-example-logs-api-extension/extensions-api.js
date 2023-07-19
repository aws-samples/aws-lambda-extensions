const http = require('http');
const { basename } = require('path');

const baseUrl = `http://${process.env.AWS_LAMBDA_RUNTIME_API}/2020-01-01/extension`;

async function register() {
    const requestBody = JSON.stringify({
        events: ['INVOKE', 'SHUTDOWN'],
    });

    const options = {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Lambda-Extension-Name': basename(__dirname),
        },
    };

    return new Promise((resolve, reject) => {
        const req = http.request(`${baseUrl}/register`, options, (res) => {
            let data = '';
            res.on('data', (chunk) => {
                data += chunk;
            });

            res.on('end', () => {
                if (res.statusCode >= 200 && res.statusCode < 300) {
                    const extensionId = res.headers['lambda-extension-identifier'];
                    resolve(extensionId);
                } else {
                    console.error('register failed', data);
                    reject(new Error('Registration failed'));
                }
            });
        });

        req.on('error', (error) => {
            console.error('register error:', error);
            reject(error);
        });

        req.write(requestBody);
        req.end();
    });
}

async function next(extensionId) {
    const options = {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Lambda-Extension-Identifier': extensionId,
        },
    };

    return new Promise((resolve, reject) => {
        const req = http.request(`${baseUrl}/event/next`, options, (res) => {
            let data = '';
            res.on('data', (chunk) => {
                data += chunk;
            });

            res.on('end', () => {
                if (res.statusCode >= 200 && res.statusCode < 300) {
                    try {
                        const eventData = JSON.parse(data);
                        resolve(eventData);
                    } catch (error) {
                        console.error('next failed', error);
                        reject(error);
                    }
                } else {
                    console.error('next failed', data);
                    reject(new Error('Next event failed'));
                }
            });
        });

        req.on('error', (error) => {
            console.error('next error:', error);
            reject(error);
        });

        req.end();
    });
}

module.exports = {
    register,
    next,
};
