package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/callback"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/scheduler"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
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

type Proxy struct {
	metpub         *metric.Publisher
	sch            *scheduler.CloudWatchScheduler
	cbak           *callback.CloudFormationCallbackAdapter
	customResource InvokeHandler
	proxyCreds     *credentials.Credentials
}

//HandleLambdaEvent is the main entry point for the lambda function.
// A response will be output on all paths, though CloudFormation will
// not block on invoking the handlers, but rather listen for callbacks
func (p *Proxy) HandleLambdaEvent(ctx context.Context, event HandlerRequest) (r HandlerResponse, e error) {

	defer func(event HandlerRequest) {
		if e := recover(); e != nil {
			r = createProgressResponse(p.Panics(event, e), event.BearerToken)
		}
	}(event)

	p.initialiseRuntime(event)

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

	res := p.ProcessInvocation(ctx, event)

	return createProgressResponse(res, event.BearerToken), nil

}

func (p *Proxy) initialiseRuntime(req HandlerRequest) {

	u := url.URL{
		Scheme: "https",
		Host:   req.ResponseEndpoint,
	}

	//Set the caller credentials
	p.proxyCreds = setproxyCreds(req)

	// If null, we are not running a test.
	if p.cbak == nil {
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

		p.cbak = callback.New(cloudformation.New(cfsess))
	}

	// If null, we are not running a test.
	if p.metpub == nil || p.sch == nil {
		//Create a Cloudwatch events and Cloudwatch AWS session.
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(req.Region),
			Credentials: credentials.NewStaticCredentials(req.Data.PlatformCredentials.AccessKeyID, req.Data.PlatformCredentials.SecretAccessKey, req.Data.PlatformCredentials.SessionToken),
			Endpoint:    aws.String(u.String()),
		})

		if err != nil {
			panic(err)
		}
		p.metpub = metric.New(cloudwatch.New(sess), req.ResourceType)
		p.sch = scheduler.New(cloudwatchevents.New(sess))

	}
}

//ProcessInvocation process the request information and invokes the handler.
func (p *Proxy) ProcessInvocation(cx context.Context, req HandlerRequest) (r *ProgressEvent) {

	hr := &ProgressEvent{}

	//Get the lambda Context.
	lc, _ := lambdacontext.FromContext(cx)

	//If Action.CREATE, Action.DELETE, or Action.UPDATE validate if the request has properties
	validateResourceProperties(req.Data.ResourceProperties, req.Action)

	// transform the request object to pass to caller.
	resHanReq := p.transform(req)

	p.checkReinvoke(req.Context)

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
			p.metpub.PublishInvocationMetric(time.Now(), req.Action)

			// Report the work is done.
			re := p.invoke(resHanReq, &req)

			ch <- re
		}()

		// Wait for the work to finish. If it takes too long move on.
		select {
		case d := <-ch:
			elapsed := time.Since(st)
			p.metpub.PublishDurationMetric(time.Now(), req.Action, elapsed.Seconds()*1e3)
			computeLocally = p.scheduleReinvocation(cx, &req, d, lc)
			hr = d

		case <-ctx.Done():
			//handler failed to respond; shut it down.
			elapsed := time.Since(st)
			p.metpub.PublishDurationMetric(time.Now(), req.Action, elapsed.Seconds()*1e3)
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

//transform the the request into a resource handler.
//Using reflection, finds the type of th custom resource,
//Unmarshalls DesiredResource and PreviousResourceState, sets the field in the
//CustomHandler and returns a ResourceHandlerRequest.
func (p *Proxy) transform(r HandlerRequest) *ResourceHandlerRequest {

	// Custom resource struct.
	v := reflect.ValueOf(p.customResource)

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
		//AwsAccountID:        r.AwsAccountID,
		//NextToken:           r.NextToken,
		//Region:              r.Region,
		//ResourceType:        r.ResourceType,
		//ResourceTypeVersion: r.ResourceTypeVersion,
	}
}

// If this invocation was triggered by a 're-invoke' CloudWatch Event, clean it up.
func (p *Proxy) checkReinvoke(context RequestContext) {

	if context.CloudWatchEventsRuleName != "" && context.CloudWatchEventsTargetID != "" {

		if err := p.sch.CleanupCloudWatchEvents(context.CloudWatchEventsRuleName, context.CloudWatchEventsTargetID); err != nil {

			panic(err)
		}
	}
}

func (p *Proxy) scheduleReinvocation(c context.Context, req *HandlerRequest, hr *ProgressEvent, l *lambdacontext.LambdaContext) bool {

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

	p.sch.RescheduleAfterMinutes(l.InvokedFunctionArn, hr.CallbackDelaySeconds, string(rj), time.Now(), uID, rn, tID)

	return false
}

//Helper to method to invoke th CustomResouce handler function.
func (p *Proxy) invoke(request *ResourceHandlerRequest, input *HandlerRequest) *ProgressEvent {

	switch input.Action {
	case create:
		return p.customResource.CreateRequest(request, input.Context.CallbackContext)

	case delete:
		return p.customResource.DeleteRequest(request, input.Context.CallbackContext)

	case list:
		return p.customResource.ListRequest(request, input.Context.CallbackContext)

	case read:
		return p.customResource.ReadRequest(request, input.Context.CallbackContext)

	case update:
		return p.customResource.UpdateRequest(request, input.Context.CallbackContext)

	default:
		return nil
	}
}

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func (p *Proxy) Panics(event HandlerRequest, r interface{}) *ProgressEvent {

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

	// Log the Go stack trace for this panic.
	//log.Printf("%s :\n%s", event.Data.ResourceProperties, debug.Stack())

	if (!reflect.DeepEqual(event.Data.PlatformCredentials, Credentials{})) {

		if perr := p.metpub.PublishExceptionMetric(time.Now(), event.Action, err); perr != nil {
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
func (p *Proxy) InjectCredentialsAndInvoke(req request.Request) error {

	req.Config.Credentials = p.proxyCreds
	err := req.Send()
	if err != nil {
		return err
	}

	return nil
}

func setproxyCreds(r HandlerRequest) *credentials.Credentials {
	return credentials.NewStaticCredentials(r.Data.CallerCredentials.AccessKeyID, r.Data.CallerCredentials.SecretAccessKey, r.Data.CallerCredentials.SessionToken)
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

//Valdiate the model against schemata.
func valdiateSchema(request *RequestContext, action string) {
	if action == "CREATE" || action == "DELETE" || action == "UPDATE" {

		//Todo: make call to validation api

	}
}

func validateResourceProperties(in json.RawMessage, action string) {
	//If Action.CREATE, Action.DELETE or Action.UPDATE, make sure that properties were passed in.

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
