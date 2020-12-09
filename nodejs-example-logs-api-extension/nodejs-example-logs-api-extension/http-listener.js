const http = require('http')

function listen(port) {
    const logsQueue = [];
    // init HTTP server for the Logs API subscription
    const server = http.createServer(function(request, response) {
        if (request.method == 'POST') {
            var body = '';
            request.on('data', function(data) {
                body += data;
            });
            request.on('end', function() {
                console.log('Logs listener received: ' + body);
                try {
                    let batch = JSON.parse(body);
                    if (batch.length > 0) {
                        logsQueue.push( ...batch );
                    }
                } catch(e) { 
                    console.log("failed to parse logs"); 
                }
                response.writeHead(200, {})
                response.end("OK")
            });
        } else {
            console.log('GET');
            response.writeHead(200, {});
            response.end("OK");
        }
    });
    server.listen(port, "sandbox");
    console.log(`Listening for logs at http://sandbox:${port}`);
    return { logsQueue };
}

module.exports = {
    listen,
};
