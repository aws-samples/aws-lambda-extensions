console.log("Hello from function initalization");

exports.handler = async (event, context) => {
  for (let i = 0; i < 1000; i++) {
    console.log(`I'm log ${i}`);
  }
  const response = {
    statusCode: 200,
    body: "hello, world",
  };
  return response;
};
