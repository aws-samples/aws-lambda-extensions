# Example Lambda Layer using AWS SAM and the Go example extension

## Quick Start

### Pre-reqs

You'll need AWS SAM and Go in your local machine.

### 1. Build

❯ sam build
```
Building function 'HelloWorldFunction'
Running CustomMakeBuilder:CopySource
Running CustomMakeBuilder:MakeBuild
Current Artifacts Directory : /Users/sancard/work/src/code.amazon.com/LambdaExtensionsShowcase/go-example-extension-sam-layer/.aws-sam/build/HelloWorldFunction
Building layer 'GoExampleExtensionLayer'
Running CustomMakeBuilder:CopySource
Running CustomMakeBuilder:MakeBuild
Current Artifacts Directory : /Users/sancard/work/src/code.amazon.com/LambdaExtensionsShowcase/go-example-extension-sam-layer/.aws-sam/build/GoExampleExtensionLayer

Build Succeeded

Built Artifacts  : .aws-sam/build
Built Template   : .aws-sam/build/template.yaml

Commands you can use next
=========================
[*] Invoke Function: sam local invoke
[*] Deploy: sam deploy --guided
```
### 2. Deploy

If first time, use `sam deploy --guided`, otherwise:

❯ sam deploy
```
	Deploying with following values
	===============================
	Stack name                 : go-example-extension-sam-layer
	Region                     : us-east-1
	Confirm changeset          : False
	Deployment s3 bucket       : aws-sam-cli-managed-default-samclisourcebucket-xxxxxxxx
	Capabilities               : ["CAPABILITY_IAM"]
	Parameter overrides        : {}

Initiating deployment
=====================

Waiting for changeset to be created..

CloudFormation stack changeset
---------------------------------------------------------------------------------------------------------------------------------------------------------
Operation                                           LogicalResourceId                                   ResourceType                                      
---------------------------------------------------------------------------------------------------------------------------------------------------------
+ Add                                               GoExampleExtensionLayer12347879e2                   AWS::Lambda::LayerVersion                         
+ Add                                               HelloWorldFunctionRole                              AWS::IAM::Role                                    
+ Add                                               HelloWorldFunction                                  AWS::Lambda::Function                             
---------------------------------------------------------------------------------------------------------------------------------------------------------

Changeset created successfully. arn:aws:cloudformation:us-east-1:123456789012:changeSet/samcli-deploy1234316323/abcde523-abcd-abcd-ad18-abcd4b2f2a67


2020-XX-XX XX:XX:XX - Waiting for stack create/update to complete

CloudFormation events from changeset
---------------------------------------------------------------------------------------------------------------------------------------------------------
ResourceStatus                         ResourceType                           LogicalResourceId                      ResourceStatusReason                 
---------------------------------------------------------------------------------------------------------------------------------------------------------
CREATE_IN_PROGRESS                     AWS::IAM::Role                         HelloWorldFunctionRole                 Resource creation Initiated          
CREATE_IN_PROGRESS                     AWS::IAM::Role                         HelloWorldFunctionRole                 -                                    
CREATE_IN_PROGRESS                     AWS::Lambda::LayerVersion              GoExampleExtensionLayer12347879e2      -                                    
CREATE_COMPLETE                        AWS::Lambda::LayerVersion              GoExampleExtensionLayer12347879e2      -                                    
CREATE_IN_PROGRESS                     AWS::Lambda::LayerVersion              GoExampleExtensionLayer12347879e2      Resource creation Initiated          
CREATE_COMPLETE                        AWS::IAM::Role                         HelloWorldFunctionRole                 -                                    
CREATE_IN_PROGRESS                     AWS::Lambda::Function                  HelloWorldFunction                     -                                    
CREATE_COMPLETE                        AWS::Lambda::Function                  HelloWorldFunction                     -                                    
CREATE_IN_PROGRESS                     AWS::Lambda::Function                  HelloWorldFunction                     Resource creation Initiated          
CREATE_COMPLETE                        AWS::CloudFormation::Stack             go-example-extension-sam-layer         -                                    
---------------------------------------------------------------------------------------------------------------------------------------------------------

CloudFormation outputs from deployed stack
----------------------------------------------------------------------------------------------------------------------------------------------------------
Outputs                                                                                                                                                  
----------------------------------------------------------------------------------------------------------------------------------------------------------
Key                 GoExampleExtensionLayer                                                                                                              
Description         Go Example Lambda Extension Layer Version ARN                                                                                        
Value               arn:aws:lambda:us-east-1:123456789012:layer:go-example-extension:1                                                                   

Key                 HelloWorldFunctionIamRole                                                                                                            
Description         Implicit IAM Role created for Hello World function                                                                                   
Value               arn:aws:iam::123456789012:role/go-example-extension-sam-la-HelloWorldFunctionRole-XXXXX2U1MWG0                                       

Key                 HelloWorldFunction                                                                                                                   
Description         First Lambda Function ARN                                                                                                            
Value               arn:aws:lambda:us-east-1:123456789012:function:go-example-extension-sam-layer-HelloWorldFunction-XXXXXX36QAAC5                       
----------------------------------------------------------------------------------------------------------------------------------------------------------

Successfully created/updated stack - go-example-extension-sam-layer in us-east-1
```

### 3. Destroy

Go to the CloudFormation console and delete the stack created or delete via AWS CLI.