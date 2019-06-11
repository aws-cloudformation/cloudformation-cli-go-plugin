package proxy

import (
	"context"
	"log"
	"net/url"
	"reflect"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/scheduler"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

//HandleLambdaEvent is the main entry point for the lambda function.
// A response will be output on all paths, though CloudFormation will
// not block on invoking the handlers, but rather listen for callbacks
func HandleLambdaEvent(ctx context.Context, event HandlerRequest) (Reponse, error) {

	if (reflect.DeepEqual(event, HandlerRequest{})) {
		log.Panicln("No request object received")
	}

	r := resor.ProcessInvocation(initialiseRuntime(ctx, event))

	return Reponse{
		Status:        r.OperationStatus,
		Message:       r.Message,
		ResourceModel: r.ResourceModel,
	}, nil
}

func initialiseRuntime(ct context.Context, req HandlerRequest) *ProcessInvocationInput {

	if req.ResponseEndpoint == "" {
		log.Panicln("Response endpoint was not provided.")
	}

	if req.Region == "" {
		log.Panicln("Region was not provided.")
	}

	u := url.URL{
		Scheme: "https",
		Host:   req.ResponseEndpoint,
	}

	//Create a Cloudformation AWS session.
	cfsess, err := session.NewSession(&aws.Config{
		Region:      aws.String(req.Region),
		Credentials: credentials.NewStaticCredentials(req.Data.platformCredentials.AccessKeyID, req.Data.platformCredentials.SecretAccessKey, req.Data.platformCredentials.SessionToken),
		Endpoint:    aws.String(u.String()),
		MaxRetries:  aws.Int(16),
	})

	if err != nil {
		log.Panicln("Sesson error: ", err)
	}

	//Create a Cloudwatch AWS session.
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(req.Region),
		Credentials: credentials.NewStaticCredentials(req.Data.platformCredentials.AccessKeyID, req.Data.platformCredentials.SecretAccessKey, req.Data.platformCredentials.SessionToken),
		Endpoint:    aws.String(u.String()),
	})

	if err != nil {
		log.Panicln("Sesson error: ", err)
	}

	return &ProcessInvocationInput{
		Cx:     ct,
		Req:    req,
		Metric: metric.New(cloudwatch.New(sess), req.ResourceType),
		Sched:  scheduler.New(cloudwatchevents.New(sess)),
	}
}
