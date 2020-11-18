#!/usr/bin/env node
const {register, next} = require('./extensions-api');
const secretCaches = require('./secrets');

const EventType = {
    SHUTDOWN: 'SHUTDOWN',
};

function handleShutdown(event) {
    console.log("Shutting down the container");
    process.exit(0);
}

(async function main() {
    process.on('SIGINT', () => handleShutdown('SIGINT'));
    process.on('SIGTERM', () => handleShutdown('SIGTERM'));

    const extensionId = await register();
    await secretCaches.cacheSecrets();
    await secretCaches.startHttpServer();

    // execute extensions logic
    while (true) {
        const event = await next(extensionId);
        switch (event.eventType) {
            case EventType.SHUTDOWN:
                handleShutdown(event);
                break;
            default:
                throw new Error('unknown event: ' + event.eventType);
        }
    }
})();
