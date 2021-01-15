const AWS = require("aws-sdk");
const response = require("cfn-response");
const docClient = new AWS.DynamoDB.DocumentClient();
exports.handler = function(event, context) {
    console.log(JSON.stringify(event,null,2));
    let params = {
        TableName: event.ResourceProperties.DBName,
        Item:{
        "pKey": "pKey1",
        "sKey": "sKey1",
        "Data": "Sample Data"
        }
    };
    docClient.put(params, function(err, data) { 
        if (err) {
        response.send(event, context, "FAILED", {});
        } else {
        response.send(event, context, "SUCCESS", {});
        }
    });
};
