#!/usr/bin/env node
const yaml = require('js-yaml');
const fs = require('fs');
const AWS = require('aws-sdk');
const express = require("express");
const app = express();
const file = "/var/task/config.yaml";
const PORT = 8080;

async function cacheSecrets() {
    fs.open(file, 'r', (err, fd) => {
        if (err) {
            if (err.code === 'ENOENT') {
                console.log('Config.yaml doesnt exist so no parsing required');
                return;
            }
        }
        readFile();
    });
}

var secretsCache = {};
var cacheLastUpdated;

async function readFile() {
    // Read the file
    try {
        const awsSecretManager = new AWS.SecretsManager();
        var fileContents = fs.readFileSync(file, 'utf8');
        var data = yaml.safeLoad(fileContents);

        if (data !== null) {
            var secretManagers = data.SecretManagers;
            for (var i = 0; i < secretManagers.length; i++) {
                var secrets = secretManagers[i].secrets;
                for (var j = 0; j < secrets.length; j++) {
                    var secretName = secrets[j];
                    try {
                        // Read secrets from SecretManager
                        const secretResponse = await awsSecretManager.getSecretValue({SecretId: secretName}).promise();
                        secretsCache[secretName] = secretResponse.SecretString;
                    } catch (e) {
                        console.log("Error while getting secret name " + secretName);
                    }
                }
            }

            // Read timeout from environment variable and set expiration timestamp
            var timeOut = parseInt(process.env.CACHE_TIMEOUT || 10);
            var s = new Date();
            s.setMinutes(s.getMinutes() + timeOut);
            cacheLastUpdated = s;
        }
    } catch (e) {
        console.error(e);
    }
}

async function processPayload(req, res) {
    var now = new Date();
    if (now > new Date()) {
        await readFile();
        console.log("Cache update is complete")
    }

    var secretName = req.params['name'];
    var secretValue = secretsCache[secretName];
    res.setHeader("Content-Type", "application/json");
    res.status(200);
    res.end(secretValue);
}

async function startHttpServer() {
    app.get("/cache/:name", function (req, res) {
        return processPayload(req, res);
    });

    app.listen(PORT, function (error) {
        if (error) throw error
        console.log("Server created Successfully on PORT", PORT)
    });
}


module.exports = {
    cacheSecrets,
    startHttpServer
};
