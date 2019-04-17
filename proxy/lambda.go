package proxy

import (
	"context"
	"log"
	"reflect"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/scheduler"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

//HandleLambdaEvent is the main entry point for the lambda function.
// A response will be output on all paths, though CloudFormation will
// not block on invoking the handlers, but rather listen for callbacks
func HandleLambdaEvent(ctx context.Context, event HandlerRequest) (Reponse, error) {

	//Open an AWS session.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		log.Fatal("Sesson error: ", err)
	}

	if (reflect.DeepEqual(event, HandlerRequest{})) {
		log.Panicln("No request object received")
	}

	p := ProcessInvocationInput{
		Cx:     ctx,
		Req:    event,
		Metric: metric.New(cloudwatch.New(sess), event.ResourceType),
		Sched:  scheduler.New(cloudwatchevents.New(sess)),
	}

	pxy := New(p)

	r := pxy.ProcessInvocation()

	return Reponse{
		Status:        r.ProgressStatus,
		Message:       r.Message,
		ResourceModel: r.ResourceModel,
	}, nil
}
