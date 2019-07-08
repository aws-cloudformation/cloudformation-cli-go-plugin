package proxy

import (
	"context"
	"errors"
	"log"
	"net/url"
	"reflect"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/callback"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/scheduler"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

var proxyCreds *credentials.Credentials
var metpub *metric.Publisher
var sch *scheduler.CloudWatchScheduler
var cbak *callback.CloudFormationCallbackAdapter

//HandleLambdaEvent is the main entry point for the lambda function.
// A response will be output on all paths, though CloudFormation will
// not block on invoking the handlers, but rather listen for callbacks
func HandleLambdaEvent(ctx context.Context, event HandlerRequest) (r HandlerResponse, e error) {

	defer func(event HandlerRequest) {
		if e := recover(); e != nil {
			r = createProgressResponse(Panics(event, e), event.BearerToken)
		}
	}(event)

	initialiseRuntime(event)

	//Pre checks to ensure a stable request.
	if (reflect.DeepEqual(event, HandlerRequest{})) {
		panic("No request object received")
	}

	if event.ResponseEndpoint == "" {
		panic("No callback endpoint received")
	}

	if event.BearerToken == "" {
		panic("No BearerToken received")
	}

	if (reflect.DeepEqual(event.Data.PlatformCredentials, Credentials{})) {
		panic("Missing required platform credentials")
	}

	if event.Region == "" {
		panic("Region was not provided.")
	}

	res := resor.ProcessInvocation(ctx, event)

	return createProgressResponse(res, event.BearerToken), nil

}

func createProgressResponse(progressEvent *ProgressEvent, bearerToken string) HandlerResponse {

	return HandlerResponse{
		Message:         progressEvent.Message,
		OperationStatus: progressEvent.OperationStatus,
		ResourceModel:   progressEvent.ResourceModel,
		BearerToken:     bearerToken,
		ErrorCode:       progressEvent.HandlerErrorCode,
	}

}

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func Panics(event HandlerRequest, r interface{}) *ProgressEvent {

	var err error

	// find out exactly what the error was and set err
	switch x := r.(type) {
	case string:
		err = errors.New(x)
	case error:
		err = x
	default:
		err = errors.New("Unknown panic")
	}

	// Log the Go stack trace for this panic'd goroutine.
	//log.Printf("%s :\n%s", event.Data.ResourceProperties, debug.Stack())

	if (!reflect.DeepEqual(event.Data.PlatformCredentials, Credentials{})) {

		if perr := metpub.PublishExceptionMetric(time.Now(), event.Action, err); perr != nil {
			log.Printf("%s : %s", "Publish error metric failed ", perr.Error())
		}

	}

	//Return a a progress event.
	hr := &ProgressEvent{
		Message:              err.Error(),
		OperationStatus:      Failed,
		ResourceModel:        event.Data.ResourceProperties,
		CallbackDelaySeconds: 0,
		HandlerErrorCode:     InvalidRequest,
	}

	return hr

}

//SetproxyCreds sets the clients call credentials
func setproxyCreds(r HandlerRequest) {
	proxyCreds = credentials.NewStaticCredentials(r.Data.CallerCredentials.AccessKeyID, r.Data.CallerCredentials.SecretAccessKey, r.Data.CallerCredentials.SessionToken)
}

func initialiseRuntime(req HandlerRequest) {

	u := url.URL{
		Scheme: "https",
		Host:   req.ResponseEndpoint,
	}

	setproxyCreds(req)

	// If null, we are not running a test.
	if cbak == nil {
		//Create a Cloudformation AWS session.
		cfsess, err := session.NewSession(&aws.Config{
			Region:      aws.String(req.Region),
			Credentials: credentials.NewStaticCredentials(req.Data.PlatformCredentials.AccessKeyID, req.Data.PlatformCredentials.SecretAccessKey, req.Data.PlatformCredentials.SessionToken),
			Endpoint:    aws.String(u.String()),
			MaxRetries:  aws.Int(16),
		})

		if err != nil {
			panic(err)
		}

		cbak = callback.New(cloudformation.New(cfsess))
	}

	// If null, we are not running a test.
	if metpub == nil || sch == nil {
		//Create a Cloudwatch events and Cloudwatch AWS session.
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(req.Region),
			Credentials: credentials.NewStaticCredentials(req.Data.PlatformCredentials.AccessKeyID, req.Data.PlatformCredentials.SecretAccessKey, req.Data.PlatformCredentials.SessionToken),
			Endpoint:    aws.String(u.String()),
		})

		if err != nil {
			panic(err)
		}
		metpub = metric.New(cloudwatch.New(sess), req.ResourceType)
		sch = scheduler.New(cloudwatchevents.New(sess))

	}
}
