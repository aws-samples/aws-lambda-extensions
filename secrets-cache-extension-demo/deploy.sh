# shellcheck disable=SC2164
cd nodejs-secrets-aws-lambda-extension
chmod +x *.js
npm install
cd ..

chmod +x extensions/nodejs-secrets-aws-lambda-extension
rm -Rf extension.zip
zip -r extension.zip .

aws lambda publish-layer-version \
 --layer-name "Secrets-Lambda-Extension-Layer" \
 --zip-file  "fileb://extension.zip"

aws lambda update-function-configuration \
 --function-name Secrets-Extension-Lambda-Test \
 --layers $(aws lambda list-layer-versions --layer-name Secrets-Lambda-Extension-Layer  \
--max-items 1 --no-paginate --query 'LayerVersions[0].LayerVersionArn' \
--output text)

# aws logs describe-log-groups --query 'logGroups[?starts_with(logGroupName,`/aws/lambda/Test`)].logGroupName' \
# --output table | awk '{print $2}' | grep -v ^$ | while read x; do aws logs delete-log-group --log-group-name $x; done
