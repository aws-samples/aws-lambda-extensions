// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
exports.handler = function(event, context, callback) {

    const https = require('http')
    const options = {
        hostname: 'localhost',
        port: 3000,
        path: '/dynamodb/' + process.env.databaseName + "-pKey1-sKey1",
        method: 'GET'
    }

    const req = https.request(options, res => {
        res.on('data', d => {
            console.log("Finally got some response here: "+d);
            return d;
        })
    })

    req.on('error', error => {
        console.error(error)
    })

    req.end()
};