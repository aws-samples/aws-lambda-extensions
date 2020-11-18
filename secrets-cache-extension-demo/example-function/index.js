exports.handler = function(event, context, callback) {

    const https = require('http')
    const options = {
        hostname: 'localhost',
        port: 8080,
        path: '/cache/secret_now',
        method: 'GET'
    }

    const req = https.request(options, res => {
        res.on('data', d => {
            console.log("Response from cache: "+d);
            return d;
        })
    })

    req.on('error', error => {
        console.error(error)
    })

    req.end()
};
