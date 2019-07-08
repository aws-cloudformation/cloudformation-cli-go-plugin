package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/scheduler"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/request"
)

const (
	InProgress = "IN_PROGRESS"
	Complete   = "SUCCESS"
	Failed     = "FAILED"
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
	CreateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent
	DeleteRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent
	ListRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent
	ReadRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent
	UpdateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent
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
func (c *CustomHandler) ProcessInvocation(cx context.Context, req HandlerRequest) (r *ProgressEvent) {

	hr := &ProgressEvent{}

	//Set the lambda Context.
	lc, _ := lambdacontext.FromContext(cx)

	//If Action.CREATE, Action.DELETE, or Action.UPDATE validate if the request has properties
	validateResourceProps(req.Data.ResourceProperties, req.Action)

	// transform the request object to pass to caller.
	resHanReq := Transform(req, resor)

	checkReinvoke(req.Context)

	//valdiate()

	// Set a duration.
	//TODO: set Duration value to Constant
	duration := 60 * time.Second
	computeLocally := true
	for {
		// Create a context that is both manually cancellable and will signal
		// a cancel at the specified duration.
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		defer cancel()

		// Create a channel to received a signal that work is done.
		ch := make(chan *ProgressEvent, 1)

		//start the timer
		st := time.Now()

		// Ask the goroutine to do some work for us.
		go func() {
			//Publish invocation metric
			metpub.PublishInvocationMetric(time.Now(), req.Action)

			// Report the work is done.
			re := c.invoke(resHanReq, &req)

			ch <- re
		}()

		// Wait for the work to finish. If it takes too long move on.
		select {
		case d := <-ch:
			elapsed := time.Since(st)
			metpub.PublishDurationMetric(time.Now(), req.Action, elapsed.Seconds()*1e3)
			computeLocally = scheduleReinvocation(cx, &req, d, lc)
			hr = d

		case <-ctx.Done():
			//handler failed to respond; shut it down.
			elapsed := time.Since(st)
			metpub.PublishDurationMetric(time.Now(), req.Action, elapsed.Seconds()*1e3)
			panic("Handler failed to respond")
		}

		// report the progress status when in non-terminal state (i.e; InProgress) back to configured endpoint
		//cbak.ReportProgress(req.BearerToken, hr.HandlerErrorCode, hr.OperationStatus, hr.ResourceModel, hr.Message)

		if !computeLocally {
			break
		}

	}

	return hr

}

//Managed scheduling of handler re-invocations.
func scheduleReinvocation(c context.Context, req *HandlerRequest, hr *ProgressEvent, l *lambdacontext.LambdaContext) bool {

	if hr.OperationStatus != InProgress {
		// no reinvoke required
		return false
	}

	req.Context.Invocation = req.Context.Invocation + 1

	cbcx, err := json.Marshal(hr.CallbackContext)

	if err != nil {
		panic(err)
	}

	req.Context.CallbackContext = json.RawMessage(cbcx)

	uID, err := scheduler.NewUUID()
	if err != nil {
		panic(err)
	}

	rn := fmt.Sprintf("reinvoke-handler-%s", uID)
	tID := fmt.Sprintf("reinvoke-target-%s", uID)

	// record the CloudWatchEvents objects for cleanup on the callback
	req.Context.CloudWatchEventsRuleName = rn
	req.Context.CloudWatchEventsTargetID = tID

	rcx, err := json.Marshal(req.Context)

	if err != nil {
		panic(err)
	}

	//when a handler requests a sub-minute callback delay, and if the lambda
	//invocation
	//has enough runtime (with 20% buffer), we can reschedule from a thread wait
	//otherwise we re-invoke through CloudWatchEvents which have a granularity of
	//minutes.

	deadline, _ := c.Deadline()
	secondsUnitDeadline := time.Until(deadline).Seconds()

	if hr.CallbackDelaySeconds < 60 && secondsUnitDeadline > float64(hr.CallbackDelaySeconds)*1.2 {

		log.Printf("Scheduling re-invoke locally after %v seconds, with Context %s", hr.CallbackDelaySeconds, string(rcx))

		time.Sleep(time.Duration(hr.CallbackDelaySeconds) * time.Second)

		return true
	}

	log.Printf("Scheduling re-invoke with Context {%s}", string(rcx))

	rj, err := json.Marshal(req)

	sch.RescheduleAfterMinutes(l.InvokedFunctionArn, hr.CallbackDelaySeconds, string(rj), time.Now(), uID, rn, tID)

	return false
}

//Helper to method to invoke th CustomResouce handler function.
func (c *CustomHandler) invoke(request *ResourceHandlerRequest, input *HandlerRequest) *ProgressEvent {
	switch input.Action {
	case create:

		return c.CustomResource.CreateRequest(request, input.Context.CallbackContext)

	case delete:

		return c.CustomResource.DeleteRequest(request, input.Context.CallbackContext)

	case list:

		return c.CustomResource.ListRequest(request, input.Context.CallbackContext)

	case read:

		return c.CustomResource.ReadRequest(request, input.Context.CallbackContext)

	case update:

		return c.CustomResource.UpdateRequest(request, input.Context.CallbackContext)
	}
	//We should never reach this point; however, return a new error.

	return nil

}

// If this invocation was triggered by a 're-invoke' CloudWatch Event, clean it up.
func checkReinvoke(context RequestContext) {

	if context.CloudWatchEventsRuleName != "" && context.CloudWatchEventsTargetID != "" {

		if err := sch.CleanupCloudWatchEvents(context.CloudWatchEventsRuleName, context.CloudWatchEventsTargetID); err != nil {

			panic(err)
		}
	}
}

//Transform the the request into a resource handler.
//Using reflection, finds the type of th custom resource,
//Unmarshalls DesiredResource and PreviousResourceState, sets the field in the
//CustomHandler and returns a ResourceHandlerRequest.
func Transform(r HandlerRequest, handler *CustomHandler) *ResourceHandlerRequest {

	// Custom resource struct.
	v := reflect.ValueOf(handler.CustomResource)

	// Custom resource DesiredResourceState struct.
	dv := v.Elem().FieldByName("DesiredResourceState")

	//Check if the field is found and that it's a strut value.
	if !dv.IsValid() || dv.Kind() != reflect.Struct {
		panic("Unable to find DesiredResource in Config object")
	}

	// Custom resource PreviousResourceState struct.
	pv := v.Elem().FieldByName("PreviousResourceState")

	//Check if the field is found and that it's a strut value.
	if !pv.IsValid() || pv.Kind() != reflect.Struct {
		panic("Unable to find PreviousResource in Config object")
	}

	//Create new resource.
	dr := reflect.New(dv.Type())

	//Try to unmarshhal the into the strut field.
	if r.Data.ResourceProperties != nil {
		if err := json.Unmarshal([]byte(r.Data.ResourceProperties), dr.Interface()); err != nil {
			panic(err)
		}
	}

	//Set the resource.
	dv.Set(dr.Elem())

	//Create new resource.
	pr := reflect.New(pv.Type())

	//Try to unmarshhal the into the strut field.
	if r.Data.PreviousResourceProperties != nil {
		if err := json.Unmarshal([]byte(r.Data.PreviousResourceProperties), pr.Interface()); err != nil {
			panic(err)
		}
	}

	//Set the resource.
	pv.Set(pr.Elem())
	return &ResourceHandlerRequest{
		AwsAccountID:        r.AwsAccountID,
		NextToken:           r.NextToken,
		Region:              r.Region,
		ResourceType:        r.ResourceType,
		ResourceTypeVersion: r.ResourceTypeVersion,
	}
}

//Valdiate the model against schemata.
//// for CUD actions, validate incoming model - any error is a terminal failure on the invocation.
func valdiate(request *RequestContext, action string) {
	if action == "CREATE" || action == "DELETE" || action == "UPDATE" {

		//Todo: make call to validation api

	}
}

func validateResourceProps(in json.RawMessage, action string) {
	//Action.CREATE, Action.DELETE, Action.UPDATE

	if action == "CREATE" || action == "DELETE" || action == "UPDATE" {

		dst := new(bytes.Buffer)

		err := json.Compact(dst, []byte(in))

		if err != nil {
			panic("Invalid resource properties object received")
		}

		if dst.String() == "{}" {
			panic("Invalid resource properties object received")
		}

	}
}

// InjectCredentialsAndInvoke consumes a "aws/request.Request" representing the
// client's request for a service action and injects caller credentials. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
//
//// This method is useful when you want to inject credentials
// into the SDK's request.
//
//
//    // Example sending a request using the GetBucketReplicationRequest method.
//    req, resp := client.GetBucketReplicationRequest(params)
//    err := proxy.InjectCredentialsAndInvoke(req)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func InjectCredentialsAndInvoke(req request.Request) error {

	req.Config.Credentials = proxyCreds
	err := req.Send()
	if err != nil {
		return err
	}

	return nil
}

//BuildReply: Helper method to return a a ProgressEvent.
func buildReply(status string, code string, message string, seconds int, model interface{}) *ProgressEvent {

	p := ProgressEvent{
		OperationStatus:      status,
		HandlerErrorCode:     code,
		Message:              message,
		CallbackDelaySeconds: seconds,
		ResourceModel:        model,
	}

	return &p
}
