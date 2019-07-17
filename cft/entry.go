package cft

import (
	"github.com/aws/aws-lambda-go/lambda"
)

//Start in the entry point of the lambda function.
//(A type that implements InvokeHandler must be passed in.)
func Start(i InvokeHandler) {

	h := Wrapper{
		customResource: i,
	}

	//Start the lambda function.
	lambda.Start(h.HandleLambdaEvent)
}
