package proxy

import (
	"context"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/scheduler"
	"github.com/aws/aws-lambda-go/lambda"
)

var CustomHandler InvokeHandler

type ProcessInvocationInput struct {
	Cx     context.Context
	Req    HandlerRequest
	Metric *metric.Publisher
	Sched  *scheduler.CloudWatchScheduler
}

func Start(i InvokeHandler) {
	CustomHandler = i
	lambda.Start(HandleLambdaEvent)
}

func StartWithOutLambda(i InvokeHandler) {
	CustomHandler = i
}
