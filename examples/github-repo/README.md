# Example::GitHub::Repo

Manages a public GitHub repository.

## Build
You can create a build of the resource for local testing by running `make`, it will create
a build of your resource that can be used with AWS SAM CLI.

### Local testing

```bash
$ sam local invoke -e ./events/create.json TypeFunction
```

Will executed your function without calling any AWS resources, unless your code interacts
with AWS.

### Deploy
Using `make deploy` will create a production version of the resource, which will interact
with CloudFormation and CloudWatch Logs.

## Run

| Name | Description | Required |
|:-----|:------------|:---------|
| Name | Repository Name | ✅ |
| Owner | Organization/User where the repo will be created | ✅ |
| Description | Description of the repository | |
| Homepage | Home page of the project | |
| OauthToken | Personal Access Token | ✅ |

Every input property of a resource is usable as an output.