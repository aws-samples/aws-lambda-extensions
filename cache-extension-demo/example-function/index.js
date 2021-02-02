// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

const https = require('http');

exports.handler = function(event, context, callback) {

    const options = {
        hostname: 'localhost',
        port: 4000,
        path: '/dynamodb?name=DynamoDbTable-pKey1-sKey1',
        method: 'GET'
    };

    const req = https.request(options, res => {
        res.on('data', d => {
            console.log("Retrieved data from the cache: "+d);
            return d;
        });
    });

    req.on('error', error => {
        console.error(error);
    });

    req.end();
};