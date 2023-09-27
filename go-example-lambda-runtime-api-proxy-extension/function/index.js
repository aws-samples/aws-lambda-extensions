exports.handler = async (event, context) => {
    console.log('[handler] incoming event', JSON.stringify(event))
    return {
        message: 'Hello from function handler!',
    }
}
