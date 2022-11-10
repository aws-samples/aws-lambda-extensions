console.log('Hello from function initalization');

exports.handler = async (event, context) => {
    console.log('Hello from function handler', {event});
}
