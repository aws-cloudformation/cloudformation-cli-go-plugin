# {{ type_name }}

Congratulations on starting development!

## Create your resource handler

1. Write the JSON schema describing your resource:

    `{{ schema_path.name }}`

2. Generate the resource model:

    `{{ executable }} generate`

3. Implement your resource handlers by adding code to `ResourceHandler`'s functions.

## Deploy your resource handler

1. Build your handler:

    `GOOS=linux go build -ldflags="-s -w" -o bin/handler cmd/main.go`

2. Deploy the CloudFormation stack:

    `aws cloudformation package --template-file template.yml --output-template-file packaged.yml --s3-bucket <your-s3-bucket>`

    `aws cloudformation deploy --template-file packaged.yml --stack-name <your-stack> --capabilities CAPABILITY_IAM`

3. Submit your resource:

    `cfn-cli submit -v`

4. Test your resource by creating a new template that contains a resource with `Type` `{{ type_name }}` and deploy it.
