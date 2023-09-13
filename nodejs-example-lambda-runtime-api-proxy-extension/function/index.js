// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

exports.handler = async (event, context) => {
    console.log('[handler] incoming event', JSON.stringify(event));
    return {
        message: 'Hello from function handler'
    }
}
