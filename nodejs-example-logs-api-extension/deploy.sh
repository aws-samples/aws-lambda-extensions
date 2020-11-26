cd nodejs-example-logs-api-extension
chmod +x index.js
npm install
cd ..

chmod +x extensions/nodejs-example-logs-api-extension
zip -r extension.zip .

aws lambda publish-layer-version \
 --layer-name "nodejs-example-logs-api-extension" \
 --region $YOUR_AWS_REGION \
 --zip-file  "fileb://extension.zip"
