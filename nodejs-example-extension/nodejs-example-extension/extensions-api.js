const fetch = require('node-fetch');
const {basename} = require('path');

const baseUrl = `http://${process.env.AWS_LAMBDA_RUNTIME_API}/2020-01-01/extension`;

async function register() {
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
            'Lambda-Extension-Name': basename(__dirname),
        }
    });

    if (!res.ok) {
        console.error('register failed', await res.text());
    }

    return res.headers.get('lambda-extension-identifier');
}

async function next(extensionId) {
    const res = await fetch(`${baseUrl}/event/next`, {
        method: 'get',
        headers: {
            'Content-Type': 'application/json',
            'Lambda-Extension-Identifier': extensionId,
        }
    });

    if (!res.ok) {
        console.error('next failed', await res.text());
        return null;
    }

    return await res.json();
}

async function error(extensionId, phase, err) {
    const errorType = `Extension.${err.name || 'UnknownError'}`;
    await fetch(`${baseUrl}/${phase}/error`, {
        method: 'post',
        body: JSON.stringify({
            errorMessage: err.message || `${err}`,
            errorType: errorType,
            stackTrace: [ err.stack ]
        }),
        headers: {
            'Content-Type': 'application/json',
            'Lambda-Extension-Identifier': extensionId,
            'Lambda-Extension-Function-Error-Type': errorType,
        }
    });

    if (!res.ok) {
        console.error(`${phase} error failed`, await res.text());
        throw new AggregateError([err, res.text()], `Failure reporting ${phase} error`);
    }

    throw err;
}

module.exports = {
    register,
    next,
    error
};
