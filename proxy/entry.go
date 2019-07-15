package proxy

import (
	"github.com/aws/aws-lambda-go/lambda"
)

//Start in the entry point of the lambda function.
func Start(i InvokeHandler) {

	h := Wrapper{
		customResource: i,
	}

	//Start the lambda function.
	lambda.Start(h.HandleLambdaEvent)
}
