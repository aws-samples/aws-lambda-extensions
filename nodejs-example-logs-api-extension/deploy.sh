cd nodejs-example-logs-api-extension
chmod +x index.js
npm install
cd -

chmod +x extensions/nodejs-example-logs-api-extension

archive="extension.zip"
if [ -f $archive ] ; then
    rm $archive
fi

zip -r $archive ./nodejs-example-logs-api-extension
zip -r $archive ./extensions

aws lambda publish-layer-version \
 --layer-name "nodejs-example-logs-api-extension" \
 --region $YOUR_AWS_REGION \
 --zip-file  "fileb://extension.zip"
