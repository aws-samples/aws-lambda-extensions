'use strict';

exports.lambdaHandler = async (event, context) => {
    console.log('Inside Lambda function');

    for (var i = 0; i < 5; i++) {
        console.log(JSON.stringify({
            message: `Hello World ${i}`,
        }));
    }

    return {
        'statusCode': 200,
        'body': JSON.stringify({
            message: 'hello world',
        })
    };
};