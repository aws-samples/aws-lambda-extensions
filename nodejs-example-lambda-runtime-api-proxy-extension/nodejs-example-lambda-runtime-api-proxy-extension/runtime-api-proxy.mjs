// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

// Read about Lambda Runtime API here
// https://docs.aws.amazon.com/lambda/latest/dg/runtimes-api.html

import express from 'express'
const RUNTIME_API_ENDPOINT = process.env.LRAP_RUNTIME_API_ENDPOINT || process.env.AWS_LAMBDA_RUNTIME_API;
const LISTENER_PORT = process.env.LRAP_LISTENER_PORT || 9009;
const RUNTIME_API_URL = `http://${RUNTIME_API_ENDPOINT}/2018-06-01/runtime`;

export class RuntimeApiProxy {
    async start() {
        console.info(`[LRAP:RuntimeApiProxy] start RUNTIME_API_ENDPOINT=${RUNTIME_API_ENDPOINT} LISTENER_PORT=${LISTENER_PORT}`)
        const listener = express()
        listener.use(express.json())
        listener.use(this.logIncomingRequest)
        listener.get('/2018-06-01/runtime/invocation/next', this.handleNext);
        listener.post('/2018-06-01/runtime/invocation/:requestId/response', this.handleResponse);
        listener.post('/2018-06-01/runtime/init/error', this.handleInitError);
        listener.post('/2018-06-01/runtime/invocation/:requestId/error', this.handleInvokeError);
        listener.use((_, res) => res.status(404).send());
        listener.listen(LISTENER_PORT)
    }

    async handleNext(_, res){
        console.log('[LRAP:RuntimeProxy] handleNext')
        
        // Getting the next event from Lambda Runtime API
        const nextEvent = await fetch(`${RUNTIME_API_URL}/invocation/next`);
        
        // Extracting the event payload
        const eventPayload = await nextEvent.json();
        
        // Updating the event payload
        eventPayload['lrap-processed']=true;

        // Copying headers 
        nextEvent.headers.forEach((value, key)=>{
            res.set(key, value);
        });

        return res.send(eventPayload)
    }

    async handleResponse(req, res) {
        const requestId = req.params.requestId
        console.log(`[LRAP:RuntimeProxy] handleResponse requestid=${requestId}`)

        // Extracting the handler response
        const responseJson = req.body;

        // Updating the handler response
        responseJson['lrap-processed']=true;

        // Posting the updated response to Lambda Runtime API
        const resp = await fetch(`${RUNTIME_API_URL}/invocation/${requestId}/response`, {
                method: 'POST',
                body: JSON.stringify(responseJson),
            },
        )

        console.log('[LRAP:RuntimeProxy] handleResponse posted')
        return res.status(resp.status).json(await resp.json())
    }

    async handleInitError(req, res) {
        console.log(`[LRAP:RuntimeProxy] handleInitError`)

        const resp = await fetch(`${RUNTIME_API_URL}/init/error`, {
            method: 'POST',
            headers: req.headers,
            body: JSON.stringify(req.body),
        })

        console.log('[LRAP:RuntimeProxy] handleInitError posted')
        return res.status(resp.status).json(await resp.json())
    }

    async handleInvokeError(req, res) {
        const requestId = req.params.requestId
        console.log(`[LRAP:RuntimeProxy] handleInvokeError requestid=${requestId}`)
        
        const resp = await fetch(`${RUNTIME_API_URL}/invocation/${requestId}/error`, {
            method: 'POST',
            headers: req.headers,
            body: JSON.stringify(req.body),
        });

        console.log('[LRAP:RuntimeProxy] handleInvokeError posted')
        return res.status(resp.status).json(await resp.json());
    }

    logIncomingRequest(req, _, next) {
        console.log(`[LRAP:RuntimeProxy] logIncomingRequest method=${req.method} url=${req.originalUrl}`);
        next();
    }
}
