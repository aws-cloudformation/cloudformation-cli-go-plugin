package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/scheduler"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

const (
	InProgress = "InProgress"
	Complete   = "Complete"
	Failed     = "Failed"
)

const (
	create = "CREATE"
	delete = "DELETE"
	list   = "LIST"
	read   = "READ"
	update = "UPDATE"
)

const (
	InvalidRequest       = "InvalidRequest"
	AccessDenied         = "AccessDenied"
	InvalidCredentials   = "InvalidCredentials"
	NoOperationToPerform = "NoOperationToPerform"
	NotUpdatable         = "NotUpdatable"
	NotFound             = "NotFound"
	NotReady             = "NotRead"
	Throttling           = "Throttling"
	ServiceLimitExceeded = "ServiceLimitExceeded"
	ServiceTimeout       = "ServiceTimeout"
	ServiceException     = "ServiceException"
	NetworkFailure       = "NetworkFailure"
	InternalFailure      = "InternalFailure"
)

type InvokeHandler interface {
	CreateRequest(request *ResourceHandlerRequest, context RequestContext) (*ProgressEvent, error)
	DeleteRequest(request *ResourceHandlerRequest, context RequestContext) (*ProgressEvent, error)
	ListRequest(request *ResourceHandlerRequest, context RequestContext) (*ProgressEvent, error)
	ReadRequest(request *ResourceHandlerRequest, context RequestContext) (*ProgressEvent, error)
	UpdateRequest(request *ResourceHandlerRequest, context RequestContext) (*ProgressEvent, error)
}

type Proxy struct {
	in ProcessInvocationInput
}

func New(input ProcessInvocationInput) *Proxy {
	return &Proxy{
		in: input,
	}

}

//processInvocation process the request information and invokes the handler.
func (p *Proxy) ProcessInvocation() *ProgressEvent {

	//Set the request and context.
	req := p.in.Req

	//Set the lambda Context.
	lc, _ := lambdacontext.FromContext(p.in.Cx)

	//Set the Scheduler.
	sh := p.in.Sched

	//Set the Metric Publisher
	pub := p.in.Metric

	pub.PublishInvocationMetric(time.Now(), req.Action)

	if (reflect.DeepEqual(p.in.Req, HandlerRequest{})) {

		err := errors.New("No request object received")
		log.Printf("No request object received : request value %v", req)
		pub.PublishExceptionMetric(time.Now(), req.Action, err)
		rpy := buildReply(Failed, InvalidRequest, err.Error(), p.in.Req.Context.CallbackContext, 0, p.in.Req.Data.ResourceProperties)

		return rpy
	}

	// transform the request object to pass to caller
	resHanReq, err := transform(req)

	if err := p.checkReinvoke(p.in.Req.Context); err != nil {

		rpy := buildReply(Failed, InvalidRequest, err.Error(), p.in.Req.Context.CallbackContext, 0, p.in.Req.Data.ResourceProperties)

		return rpy
	}

	// for CUD actions, validate incoming model - any error is a terminal failure on the invocation
	//if req.Action == "CREATE" || req.Action == "Update" || req.Action == "Delete" {
	//	valdiate(&p.in.Req.Context)
	//}

	//start the timer
	st := time.Now()

	//todo: It would be better to make this call with a goroutine/channel and set a timeout.
	hr, err := p.invoke(resHanReq, p.in.Req.Context)

	//stop the timer
	elapsed := time.Since(st) * time.Millisecond

	pub.PublishDurationMetric(time.Now(), req.Action, elapsed.Seconds()*1e3)

	log.Printf("Handler Duration :%vms", elapsed.Seconds()*1e3)

	if err != nil {
		pub.PublishExceptionMetric(time.Now(), req.Action, err)
		rpy := buildReply(Failed, InvalidRequest, err.Error(), p.in.Req.Context.CallbackContext, 0, p.in.Req.Data.ResourceProperties)
		return rpy
	}

	if hr == nil {
		err := errors.New("Handler failed to provide a response")
		pub.PublishExceptionMetric(time.Now(), req.Action, err)

		rpy := buildReply(Failed, InvalidRequest, err.Error(), p.in.Req.Context.CallbackContext, 0, p.in.Req.Data.ResourceProperties)
		return rpy
	}

	// When the handler responses InProgress with a callback delay, we trigger a callback to re-invoke
	// the handler for the Resource type to implement stabilization checks and long-poll creation checks
	if hr.ProgressStatus == InProgress {
		req.Context.Invocation = req.Context.Invocation + 1
		req.Context.CallbackContext = hr.CallbackContext

		uID, err := scheduler.NewUUID()
		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), p.in.Req.Context.CallbackContext, 0, p.in.Req.Data.ResourceProperties)
			return rpy
		}

		rn := fmt.Sprintf("reinvoke-handler-%s", uID)
		tID := fmt.Sprintf("reinvoke-target-%s", uID)

		// record the CloudWatchEvents objects for cleanup on the callback
		req.Context.CloudWatchEventsRuleName = rn
		req.Context.CloudWatchEventsTargetID = tID

		rj, err := json.Marshal(req)

		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), p.in.Req.Context.CallbackContext, 0, p.in.Req.Data.ResourceProperties)
			return rpy
		}

		sh.RescheduleAfterMinutes(lc.InvokedFunctionArn, hr.CallbackDelayMinutes, string(rj), time.Now(), uID, rn, tID)
	}

	// report the progress status when in non-terminal state (i.e; InProgress) back to configured endpoint
	//this.callbackAdapter.reportProgress(request.getBearerToken(),
	//	handlerResponse.getErrorCode(),
	//	handlerResponse.getStatus(),
	//	handlerResponse.getResourceModel(),
	//	handlerResponse.getMessage())
	//}

	// The wrapper will log any context to the configured CloudWatch log group
	log.Printf("Call back context: %v", hr.CallbackContext)

	return hr

}

// If this invocation was triggered by a 're-invoke' CloudWatch Event, clean it up
func (p *Proxy) checkReinvoke(context RequestContext) error {
	//Set the Scheduler.
	sh := p.in.Sched

	if context.CloudWatchEventsRuleName != "" && context.CloudWatchEventsTargetID != "" {

		if err := sh.CleanupCloudWatchEvents(context.CloudWatchEventsRuleName, context.CloudWatchEventsTargetID); err != nil {

			return err
		}
	}

	return nil
}

func buildReply(status string, code string, message string, context json.RawMessage, minutes int, model json.RawMessage) *ProgressEvent {

	p := ProgressEvent{
		ProgressStatus:       status,
		HandlerErrorCode:     code,
		Message:              message,
		CallbackContext:      context,
		CallbackDelayMinutes: minutes,
		ResourceModel:        model,
	}

	return &p
}

//Transform the the request into a resource handler
func transform(r HandlerRequest) (*ResourceHandlerRequest, error) {

	v := reflect.ValueOf(CustomHandler)

	desiredResource := reflect.New(v.Type())
	if r.Data.ResourceProperties != nil {
		if err := json.Unmarshal([]byte(r.Data.ResourceProperties), desiredResource.Interface()); err != nil {
			return nil, err
		}
	}
	previousResource := reflect.New(v.Type())

	if r.Data.PreviousResourceProperties != nil {
		if err := json.Unmarshal([]byte(r.Data.PreviousResourceProperties), previousResource.Interface()); err != nil {
			return nil, err
		}

	}

	return &ResourceHandlerRequest{
		AwsAccountID:          r.AwsAccountID,
		NextToken:             r.NextToken,
		Region:                r.Region,
		ResourceType:          r.ResourceType,
		ResourceTypeVersion:   r.ResourceTypeVersion,
		Cred:                  r.Data.Creds,
		DesiredResourceState:  desiredResource.Elem().Interface(),
		PreviousResourceState: previousResource.Elem().Interface(),
	}, nil
}

func valdiate(request *RequestContext) {
	//// for CUD actions, validate incoming model - any error is a terminal failure on the invocation
}

func (p *Proxy) invoke(request *ResourceHandlerRequest, context RequestContext) (*ProgressEvent, error) {

	switch p.in.Req.Action {
	case create:
		r, err := CustomHandler.CreateRequest(request, context)

		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), p.in.Req.Context.CallbackContext, 0, p.in.Req.Data.ResourceProperties)

			return rpy, nil
		}
		return r, nil
	case delete:
		r, err := CustomHandler.DeleteRequest(request, context)

		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), p.in.Req.Context.CallbackContext, 0, p.in.Req.Data.ResourceProperties)

			return rpy, nil
		}
		return r, nil

	case list:
		r, err := CustomHandler.ListRequest(request, context)

		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), p.in.Req.Context.CallbackContext, 0, p.in.Req.Data.ResourceProperties)

			return rpy, nil
		}
		return r, nil

	case read:
		r, err := CustomHandler.ReadRequest(request, context)

		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), p.in.Req.Context.CallbackContext, 0, p.in.Req.Data.ResourceProperties)

			return rpy, nil
		}
		return r, nil

	case update:
		r, err := CustomHandler.UpdateRequest(request, context)

		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), p.in.Req.Context.CallbackContext, 0, p.in.Req.Data.ResourceProperties)

			return rpy, nil
		}
		return r, nil

	}

	//We should never reach this point; however, return a new error/
	e := errors.New("failed to parse the Action")

	return nil, e

}
