package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/aws-cloudformation-rpdk-go-plugin/cmd/rpdk/internal/createhandler"
	"github.com/aws-cloudformation-rpdk-go-plugin/cmd/rpdk/internal/deleteHandler"
	"github.com/aws-cloudformation-rpdk-go-plugin/cmd/rpdk/internal/listhandler"
	"github.com/aws-cloudformation-rpdk-go-plugin/cmd/rpdk/internal/readhandler"
	"github.com/aws-cloudformation-rpdk-go-plugin/cmd/rpdk/internal/updatehandler"
	"github.com/aws-cloudformation-rpdk-go-plugin/internal/metric"
	"github.com/aws-cloudformation-rpdk-go-plugin/internal/platform/proxy"
	"github.com/aws-cloudformation-rpdk-go-plugin/internal/scheduler"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

//Reponse is used to return the response of invoking the Lambda function to the caller.
//status: indicates whether the handler has reached a terminal state or is still computing and requires more time to complete
//message: The handler can (and should) specify a contextual information message which can be shown to callers to indicate the nature of a progress transition or callback delay; for example a message indicating "propagating to edge"
//resourceModel: The output resource instance populated by a READ/LIST for synchronous results and by CREATE/UPDATE/DELETE for final response validation/confirmation.
type Reponse struct {
	Status        string
	Message       string
	ResourceModel proxy.Model
}

type ProcessInvocationInput struct {
	Cx      context.Context
	Req     proxy.HandlerRequest
	Handles map[string]proxy.InvokeHandler
	Metric  *metric.Publisher
	Sched   *scheduler.CloudWatchScheduler
}

//HandleLambdaEvent is the main entry point for the lambda function.
// A response will be output on all paths, though CloudFormation will
// not block on invoking the handlers, but rather listen for callbacks
func HandleLambdaEvent(ctx context.Context, event proxy.HandlerRequest) (Reponse, error) {

	//Open an AWS session.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		log.Fatal("Sesson error: ", err)
	}

	if (reflect.DeepEqual(event, proxy.HandlerRequest{})) {
		log.Panicln("No request object received")
	}
	//Map of handlers
	hl := map[string]proxy.InvokeHandler{
		"Create": createhandler.New(event.Data.ResourceProperties),
		"Delete": deletehandler.New(event.Data.ResourceProperties),
		"List":   listhandler.New(event.Data.ResourceProperties),
		"Read":   readhandler.New(event.Data.ResourceProperties),
		"Update": updatehandler.New(event.Data.ResourceProperties),
	}

	p := ProcessInvocationInput{
		Cx:      ctx,
		Req:     event,
		Handles: hl,
		Metric:  metric.New(cloudwatch.New(sess), event.ResourceType),
		Sched:   scheduler.New(cloudwatchevents.New(sess)),
	}

	r, err := processInvocation(p)

	//if error, return response to caller
	if err != nil {
		debug.PrintStack()
		log.Println(err)

		m := metric.New(cloudwatch.New(sess), event.ResourceType)
		m.PublishExceptionMetric(time.Now(), event.Action, err)

		pe := Reponse{
			Status:        proxy.Failed,
			Message:       err.Error(),
			ResourceModel: event.Data.ResourceProperties,
		}

		return pe, nil

	}

	return Reponse{
		Status:        r.ProgressStatus,
		Message:       r.Message,
		ResourceModel: r.ResourceModel,
	}, nil
}

//processInvocation process the request information and invokes the handler.
func processInvocation(in ProcessInvocationInput) (*proxy.ProgressEvent, error) {

	//Set the request and context.
	request := in.Req
	ctx := in.Cx

	//Set the lambda Context.
	lc, _ := lambdacontext.FromContext(ctx)

	//Load map of handlers.
	h := in.Handles

	//Set the Scheduler.
	sh := in.Sched

	//Set the Metric Publisher
	pub := in.Metric

	if (reflect.DeepEqual(request, proxy.HandlerRequest{})) {
		e := "No request object received"

		log.Println(e)
		return nil, errors.New(e)
	}

	// transform the request object to pass to caller
	resHanReq := proxy.Transform(request)

	reqCon := request.Context

	// If this invocation was triggered by a 're-invoke' CloudWatch Event, clean it up
	if reqCon.CloudWatchEventsRuleName != "" && reqCon.CloudWatchEventsTargetID != "" {

		if err := sh.CleanupCloudWatchEvents(reqCon.CloudWatchEventsRuleName, reqCon.CloudWatchEventsTargetID); err != nil {
			log.Println(err)
			pub.PublishExceptionMetric(time.Now(), request.Action, err)

			pe := proxy.ProgressEvent{
				ProgressStatus:   proxy.Failed,
				HandlerErrorCode: proxy.InvalidRequest,
				Message:          err.Error(),
				ResourceModel:    request.Data.ResourceProperties,
			}

			return &pe, nil

		}
	}

	pub.PublishInvocationMetric(time.Now(), request.Action)

	// for CUD actions, validate incoming model - any error is a terminal failure on the invocation
	if request.Action == "CREATE" || request.Action == "Update" || request.Action == "Delete" {
		valdiate(&reqCon)
	}
	st := time.Now()

	//todo: It would be better to make this call with a goroutine/channel and set a timeout.
	hr, err := proxy.Invoke(h[request.Action], resHanReq, reqCon)

	if err != nil {
		log.Println(err)
		pub.PublishExceptionMetric(time.Now(), request.Action, err)

		return &proxy.ProgressEvent{
			ProgressStatus:   proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          err.Error(),
			CallbackContext:  reqCon.CallbackContext,
			ResourceModel:    request.Data.ResourceProperties,
		}, err
	}
	et := time.Now()
	dur := st.Sub(et)

	fmt.Print(dur)

	pub.PublishDurationMetric(time.Now(), request.Action, int64(dur/time.Millisecond))

	if hr != nil {
		log.Printf("Handler returned %s", hr.ProgressStatus)
	} else {
		err := errors.New("Handler failed to provide a response")
		pub.PublishExceptionMetric(time.Now(), request.Action, err)
		return &proxy.ProgressEvent{
			ProgressStatus:   proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
			CallbackContext:  reqCon.CallbackContext,
			ResourceModel:    request.Data.ResourceProperties,
		}, err
	}

	// When the handler responses InProgress with a callback delay, we trigger a callback to re-invoke
	// the handler for the Resource type to implement stabilization checks and long-poll creation checks
	if hr.ProgressStatus == proxy.InProgress {
		c := proxy.RequestContext{
			Invocation:      request.Context.Invocation + 1,
			CallbackContext: hr.CallbackContext,
		}

		sh.RescheduleAfterMinutes(lc.InvokedFunctionArn, hr.CallbackDelayMinutes, &c, time.Now())

		// report the progress status when in non-terminal state (i.e; InProgress) back to configured endpoint
		//this.callbackAdapter.reportProgress(request.getBearerToken(),
		//	handlerResponse.getErrorCode(),
		//	handlerResponse.getStatus(),
		//	handlerResponse.getResourceModel(),
		//	handlerResponse.getMessage())
	}

	// The wrapper will log any context to the configured CloudWatch log group
	log.Print(hr.CallbackContext)

	return hr, nil
}

func valdiate(request *proxy.RequestContext) {
	//// for CUD actions, validate incoming model - any error is a terminal failure on the invocation
}

func main() {

	lambda.Start(HandleLambdaEvent)

}
