package proxy

import (
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/callback"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/scheduler"
	"github.com/aws/aws-lambda-go/lambda"
)

var resor *CustomHandler

//Start in the entry point of the proxy. It creates a new CustomHandler and starts the lambda function.
func Start(i InvokeHandler) {

	//create a new CustomHandler
	resor = New(i)
	lambda.Start(HandleLambdaEvent)
}

//StartWithOutLambda starts the proxy without a Lambda funtion to assist in running test. It creates a new CustomHandler and injects aws dependencies.
func StartWithOutLambda(i InvokeHandler, m *metric.Publisher, s *scheduler.CloudWatchScheduler, c *callback.CloudFormationCallbackAdapter) {
	metpub = m
	sch = s
	cbak = c
	resor = New(i)
}
