package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/scheduler"
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

// InvokeHandler is an interface that the custom resource must implement.
type InvokeHandler interface {
	CreateRequest(request *ResourceHandlerRequest, context RequestContext) (*ProgressEvent, error)
	DeleteRequest(request *ResourceHandlerRequest, context RequestContext) (*ProgressEvent, error)
	ListRequest(request *ResourceHandlerRequest, context RequestContext) (*ProgressEvent, error)
	ReadRequest(request *ResourceHandlerRequest, context RequestContext) (*ProgressEvent, error)
	UpdateRequest(request *ResourceHandlerRequest, context RequestContext) (*ProgressEvent, error)
}

type ProcessInvocationInput struct {
	Cx     context.Context
	Req    HandlerRequest
	Metric *metric.Publisher
	Sched  *scheduler.CloudWatchScheduler
}

// CustomHandler is a wrapper that handles execution of the custom resource.
type CustomHandler struct {
	CustomResource InvokeHandler
}

//New is a factory function that returns a pointer ot a new CustomHandler
func New(input InvokeHandler) *CustomHandler {
	return &CustomHandler{
		CustomResource: input,
	}

}

//ProcessInvocation process the request information and invokes the handler.
func (c *CustomHandler) ProcessInvocation(in *ProcessInvocationInput) *ProgressEvent {

	//Set the request and context.
	req := in.Req

	//Set the lambda Context.
	lc, _ := lambdacontext.FromContext(in.Cx)

	//Set the Scheduler.
	sh := in.Sched

	//Set the Metric Publisher
	pub := in.Metric

	//If Action.CREATE, Action.DELETE, or Action.UPDATE validate if the request has properties
	validateResourceProps(req.Data.ResourceProperties, req.Action)

	pub.PublishInvocationMetric(time.Now(), req.Action)

	if (reflect.DeepEqual(in.Req, HandlerRequest{})) {

		err := errors.New("No request object received")
		log.Printf("No request object received : request value %v", req)
		pub.PublishExceptionMetric(time.Now(), req.Action, err)
		rpy := buildReply(Failed, InvalidRequest, err.Error(), in.Req.Context.CallbackContext, 0, in.Req.Data.ResourceProperties)

		return rpy
	}

	// transform the request object to pass to caller.
	resHanReq, err := Transform(req, resor)

	if err != nil {
		rpy := buildReply(Failed, InvalidRequest, err.Error(), in.Req.Context.CallbackContext, 0, in.Req.Data.ResourceProperties)

		return rpy
	}

	if err := c.checkReinvoke(in.Req.Context, in.Sched); err != nil {

		rpy := buildReply(Failed, InvalidRequest, err.Error(), in.Req.Context.CallbackContext, 0, in.Req.Data.ResourceProperties)

		return rpy
	}

	// for CUD actions, validate incoming model - any error is a terminal failure on the invocation
	//if req.Action == "CREATE" || req.Action == "Update" || req.Action == "Delete" {
	//	valdiate(&p.in.Req.Context)
	//}

	//start the timer
	st := time.Now()

	//todo: It would be better to make this call with a goroutine/channel and set a timeout.
	hr, err := c.invoke(resHanReq, &in.Req)

	//Stop the timer.
	elapsed := time.Since(st) * time.Millisecond

	pub.PublishDurationMetric(time.Now(), req.Action, elapsed.Seconds()*1e3)

	log.Printf("Handler Duration :%vms", elapsed.Seconds()*1e3)

	if err != nil {
		pub.PublishExceptionMetric(time.Now(), req.Action, err)
		rpy := buildReply(Failed, InvalidRequest, err.Error(), in.Req.Context.CallbackContext, 0, in.Req.Data.ResourceProperties)
		return rpy
	}

	if hr == nil {
		err := errors.New("Handler failed to provide a response")
		pub.PublishExceptionMetric(time.Now(), req.Action, err)

		rpy := buildReply(Failed, InvalidRequest, err.Error(), in.Req.Context.CallbackContext, 0, in.Req.Data.ResourceProperties)
		return rpy
	}

	// When the handler responses InProgress with a callback delay, we trigger a callback to re-invoke
	// the handler for the Resource type to implement stabilization checks and long-poll creation checks.
	if hr.OperationStatus == InProgress {
		req.Context.Invocation = req.Context.Invocation + 1
		req.Context.CallbackContext = hr.CallbackContext

		uID, err := scheduler.NewUUID()
		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), in.Req.Context.CallbackContext, 0, in.Req.Data.ResourceProperties)
			return rpy
		}

		rn := fmt.Sprintf("reinvoke-handler-%s", uID)
		tID := fmt.Sprintf("reinvoke-target-%s", uID)

		// record the CloudWatchEvents objects for cleanup on the callback
		req.Context.CloudWatchEventsRuleName = rn
		req.Context.CloudWatchEventsTargetID = tID

		rj, err := json.Marshal(req)

		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), in.Req.Context.CallbackContext, 0, in.Req.Data.ResourceProperties)
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

// If this invocation was triggered by a 're-invoke' CloudWatch Event, clean it up.
func (c *CustomHandler) checkReinvoke(context RequestContext, sh *scheduler.CloudWatchScheduler) error {

	if context.CloudWatchEventsRuleName != "" && context.CloudWatchEventsTargetID != "" {

		if err := sh.CleanupCloudWatchEvents(context.CloudWatchEventsRuleName, context.CloudWatchEventsTargetID); err != nil {

			return err
		}
	}

	return nil
}

//Transform the the request into a resource handler.
//Using reflection, finds the type of th custom resource,
//Unmarshalls DesiredResource and PreviousResourceState, sets the field in the
//CustomHandler and returns a ResourceHandlerRequest.
func Transform(r HandlerRequest, handler *CustomHandler) (*ResourceHandlerRequest, error) {

	// Custom resource struct.
	v := reflect.ValueOf(handler.CustomResource)

	// Custom resource DesiredResourceState struct.
	dv := v.Elem().FieldByName("DesiredResourceState")

	//Check if the field is found and that it's a strut value.
	if !dv.IsValid() || dv.Kind() != reflect.Struct {
		err := errors.New("Unable to find DesiredResource in Config object")
		return nil, err
	}

	// Custom resource PreviousResourceState struct.
	pv := v.Elem().FieldByName("PreviousResourceState")

	//Check if the field is found and that it's a strut value.
	if !pv.IsValid() || pv.Kind() != reflect.Struct {
		err := errors.New("Unable to find PreviousResource in Config object")
		return nil, err
	}

	//Create new resource.
	dr := reflect.New(dv.Type())

	//Try to unmarshhal the into the strut field.
	if r.Data.ResourceProperties != nil {
		if err := json.Unmarshal([]byte(r.Data.ResourceProperties), dr.Interface()); err != nil {
			return nil, err
		}
	}

	//Set the resource
	dv.Set(dr.Elem())

	//create new resource
	pr := reflect.New(pv.Type())

	//Try to unmarshhal the into the strut field.
	if r.Data.PreviousResourceProperties != nil {
		if err := json.Unmarshal([]byte(r.Data.PreviousResourceProperties), pr.Interface()); err != nil {
			return nil, err
		}
	}

	//Set the resource
	pv.Set(pr.Elem())

	return &ResourceHandlerRequest{
		AwsAccountID:        r.AwsAccountID,
		NextToken:           r.NextToken,
		Region:              r.Region,
		ResourceType:        r.ResourceType,
		ResourceTypeVersion: r.ResourceTypeVersion,
	}, nil
}

//Helper to method to invoke th CustomResouce handler function.
func (c *CustomHandler) invoke(request *ResourceHandlerRequest, input *HandlerRequest) (*ProgressEvent, error) {
	switch input.Action {
	case create:
		r, err := c.CustomResource.CreateRequest(request, input.Context)

		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), input.Context.CallbackContext, 0, input.Data.ResourceProperties)

			return rpy, nil
		}
		return r, nil
	case delete:
		r, err := c.CustomResource.DeleteRequest(request, input.Context)

		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), input.Context.CallbackContext, 0, input.Data.ResourceProperties)

			return rpy, nil
		}
		return r, nil

	case list:
		r, err := c.CustomResource.ListRequest(request, input.Context)

		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), input.Context.CallbackContext, 0, input.Data.ResourceProperties)

			return rpy, nil
		}
		return r, nil

	case read:
		r, err := c.CustomResource.ReadRequest(request, input.Context)

		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), input.Context.CallbackContext, 0, input.Data.ResourceProperties)

			return rpy, nil
		}
		return r, nil

	case update:
		r, err := c.CustomResource.UpdateRequest(request, input.Context)

		if err != nil {
			rpy := buildReply(Failed, InvalidRequest, err.Error(), input.Context.CallbackContext, 0, input.Data.ResourceProperties)

			return rpy, nil
		}
		return r, nil

	}

	//We should never reach this point; however, return a new error.
	e := errors.New("failed to parse the Action")

	return nil, e

}

//Valdiate the model against schemata.
//// for CUD actions, validate incoming model - any error is a terminal failure on the invocation.
func valdiate(request *RequestContext) {

}

func validateResourceProps(in json.RawMessage, action string) {
	//Action.CREATE, Action.DELETE, Action.UPDATE

	if action == "CREATE" || action == "DELETE" || action == "UPDATE" {

		dst := new(bytes.Buffer)

		err := json.Compact(dst, []byte(in))

		if err != nil {
			log.Panic("Invalid resource properties object received")
		}

		fmt.Println(dst)
		if dst.String() == "{}" {
			log.Panic("Invalid resource properties object received")
		}
	}
}

//BuildReply: Helper method to return a a ProgressEvent.
func buildReply(status string, code string, message string, context interface{}, minutes int, model interface{}) *ProgressEvent {

	p := ProgressEvent{
		OperationStatus:      status,
		HandlerErrorCode:     code,
		Message:              message,
		CallbackContext:      context,
		CallbackDelayMinutes: minutes,
		ResourceModel:        model,
	}

	return &p
}
