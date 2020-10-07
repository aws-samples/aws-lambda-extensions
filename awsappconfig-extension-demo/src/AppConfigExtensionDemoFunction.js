 // Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

const http = require('http');
const params = process.env.APPCONFIG_PROFILE.split('/')
const AppConfigApplication = params [0]
const AppConfigEnvironment = params [1]
const AppConfigConfiguration = params [2]
let coldstart = true;

function getConfiguration(application, environment, configuration) {
    return new Promise((resolve, reject) => {
        const req = http.get(`http://localhost:2772/applications/${application}/environments/${environment}/configurations/${configuration}`, (res) => {
            if (res.statusCode < 200 || res.statusCode >= 300) {
                return reject(new Error('statusCode=' + res.statusCode));
            }
            var body = [];
            res.on('data', function(chunk) {
                body.push(chunk);
            });
            res.on('end', function() {
                resolve(Buffer.concat(body).toString());
            });
        });
        req.on('error', (e) => {
            reject(e.message);
        });
        req.end();
    });
}

exports.handler = async (event) => {
  try {
    const configData = await getConfiguration(AppConfigApplication, AppConfigEnvironment, AppConfigConfiguration);    
    const parsedConfigData = JSON.parse(configData);
    let LogLevel = parsedConfigData.loglevel
    return {
      'event' : event,
      'ColdStart' : coldstart,
      'LogLevel': LogLevel
      }
  } catch (err) {
      console.error(err)
      return err
  } finally {
    coldstart = false;
  }
}; 
