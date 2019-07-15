package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/errs"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/callback"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/scheduler"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
	FAILED     = "FAILED"
)

const (
	create = "CREATE"
	delete = "DELETE"
	list   = "LIST"
	read   = "READ"
	update = "UPDATE"
)

const (
	InvalidRequest          = "InvalidRequest"
	AccessDenied            = "AccessDenied"
	InvalidCredentials      = "InvalidCredentials"
	NoOperationToPerform    = "NoOperationToPerform"
	NotUpdatable            = "NotUpdatable"
	NotFound                = "NotFound"
	NotReady                = "NotRead"
	Throttling              = "Throttling"
	ServiceLimitExceeded    = "ServiceLimitExceeded"
	ServiceTimeout          = "ServiceTimeout"
	ServiceException        = "ServiceException"
	NetworkFailure          = "NetworkFailure"
	InternalFailure         = "InternalFailure"
	AlreadyExists           = "AlreadyExists"
	GeneralServiceException = "GeneralServiceException"
)

// InvokeHandler is an interface that the custom resource must implement.
type InvokeHandler interface {
	CreateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*ProgressEvent, error)
	DeleteRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*ProgressEvent, error)
	ListRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*ProgressEvent, error)
	ReadRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*ProgressEvent, error)
	UpdateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*ProgressEvent, error)
}

type Wrapper struct {
	metpub         *metric.Publisher
	sch            *scheduler.CloudWatchScheduler
	cbak           *callback.CloudFormationCallbackAdapter
	customResource InvokeHandler
	wrapperCreds   *credentials.Credentials
	logger         *log.Logger
}

//initialiseRuntime initialises dependencies which are depending on credentials
//passed at function invoke and not available during construction
func (p *Wrapper) initialiseRuntime(req HandlerRequest) {

	u := url.URL{
		Scheme: "https",
		Host:   req.ResponseEndpoint,
	}

	//Set the caller credentials
	p.wrapperCreds = setWrapperCreds(req)

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

//HandleLambdaEvent is the main entry point for the lambda function.
// A response will be output on all paths, though CloudFormation will
// not block on invoking the handlers, but rather listen for callbacks
func (p *Wrapper) HandleLambdaEvent(ctx context.Context, request HandlerRequest) (lr HandlerResponse, e error) {

	//Handle all panics from the resource handler, log the error and return failed.
	defer func(event *HandlerRequest) {
		if r := recover(); r != nil {
			err := &errs.TerminalError{CustomerFacingErrorMessage: "Internal error"}

			// Log the Go stack trace for this panic'd goroutine.
			p.logger.Println(fmt.Sprintf("%s in a %s action on a %s: %s\n%s", "HandlerRequest panic", event.Action, event.ResourceType, r, debug.Stack()))
			lr = createProgressResponse(DefaultFailureHandler(err, InternalFailure), event.BearerToken)
			e = nil
		}
	}(&request)

	hr, err := p.processInvocation(ctx, request)

	if err != nil {
		// Exceptions are wrapped as a consistent error response to the caller (i.e;
		// CloudFormation)

		hr = DefaultFailureHandler(err, InternalFailure)

		if (reflect.DeepEqual(request, HandlerRequest{})) {
			if (reflect.DeepEqual(request.Data, RequestData{})) {
				hr.ResourceModel = request.Data.ResourceProperties
			}
			if perr := p.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				log.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}

		} else {
			if perr := p.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				log.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
		}
		//	switch e := err.(type) {
		//	case *errs.ValidationException :
		//		if perr := p.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
		//			log.Printf("%s : %s", "Publish error metric failed ", perr.Error())
		//		}
		//		p.logUnhandledError("An existing resource was found", request, e)
		//	default:
		//		if perr := p.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
		//			log.Printf("%s : %s", "Publish error metric failed ", perr.Error())
		//		}
		//		p.logUnhandledError("An unknown error occurred ", request, e)
		//	}
	}
	return createProgressResponse(hr, request.BearerToken), nil

}

//ProcessInvocation process the request information and invokes the handler.
func (p *Wrapper) processInvocation(cx context.Context, req HandlerRequest) (pr *ProgressEvent, e error) {

	//Handle all panics from processInvocation, log the error and return a failed Wrapper event.
	defer func(event *HandlerRequest) {
		if r := recover(); r != nil {
			err := &errs.TerminalError{CustomerFacingErrorMessage: "Internal error"}

			// Log the Go stack trace for this panic'd goroutine.
			p.logger.Println(fmt.Sprintf("%s in a %s action on a %s: %s\n%s", "processInvocation panic", event.Action, event.ResourceType, r, debug.Stack()))
			pr = nil
			e = err
		}
	}(&req)

	//Pre checks to ensure a stable request.
	if (reflect.DeepEqual(req, HandlerRequest{})) {
		return nil, &errs.TerminalError{CustomerFacingErrorMessage: "Invalid request object received"}
	}

	if (reflect.DeepEqual(req.Data, RequestData{})) {
		return nil, &errs.TerminalError{CustomerFacingErrorMessage: "Invalid resource properties object received"}
	}

	if req.Action == "CREATE" || req.Action == "DELETE" || req.Action == "UPDATE" {
		if err := validateResourceProperties(req.Data.ResourceProperties); err != nil {
			return nil, err
		}
	}

	if req.ResponseEndpoint == "" {
		return nil, &errs.TerminalError{CustomerFacingErrorMessage: "No callback endpoint received"}
	}

	if (reflect.DeepEqual(req.Data.PlatformCredentials, Credentials{})) {
		return nil, &errs.TerminalError{CustomerFacingErrorMessage: "Missing required platform credentials"}
	}

	p.initialiseRuntime(req)

	hr := &ProgressEvent{}

	//Get the lambda Context.
	lc, _ := lambdacontext.FromContext(cx)

	// transform the request object to pass to caller.
	resHanReq := p.transform(req)

	p.checkReinvoke(req.Context)

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
			p.metpub.PublishInvocationMetric(time.Now(), req.Action)

			// Report the work is done.
			re := p.wrapInvocationAndHandleErrors(resHanReq, &req)

			ch <- re
		}()

		// Wait for the work to finish. If it takes too long move on.
		select {
		case d := <-ch:
			elapsed := time.Since(st)
			p.metpub.PublishDurationMetric(time.Now(), req.Action, elapsed.Seconds()*1e3)
			hr, computeLocally = p.scheduleReinvocation(cx, &req, d, lc)

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

	return hr, nil

}

// If this invocation was triggered by a 're-invoke' CloudWatch Event, clean it up.
func (p *Wrapper) checkReinvoke(context RequestContext) {

	if context.CloudWatchEventsRuleName != "" && context.CloudWatchEventsTargetID != "" {

		if err := p.sch.CleanupCloudWatchEvents(context.CloudWatchEventsRuleName, context.CloudWatchEventsTargetID); err != nil {

			panic(err)
		}
	}
}

func validateResourceProperties(in json.RawMessage) error {
	dst := new(bytes.Buffer)
	err := json.Compact(dst, []byte(in))
	if err != nil {
		return &errs.TerminalError{"Invalid resource properties object received"}
	}
	if dst.String() == "{}" {
		return &errs.TerminalError{"Invalid resource properties object received"}
	}

	return nil
}

func setWrapperCreds(r HandlerRequest) *credentials.Credentials {
	return credentials.NewStaticCredentials(r.Data.CallerCredentials.AccessKeyID, r.Data.CallerCredentials.SecretAccessKey, r.Data.CallerCredentials.SessionToken)
}

func (p *Wrapper) logUnhandledError(errorDescription string, request *HandlerRequest, e error) {
	p.logger.Printf("%s in a %s action on a %s: %s\n%s", errorDescription, request.Action, request.ResourceType, e.Error(), debug.Stack())
}

//WrapInvocationAndHandleErrors invokes the handler implementation for the request, and handles certain classes of errors and correctly map those to
//the appropriate HandlerErrorCode Also wraps the invocation in last-mile
//timing metrics.
func (p *Wrapper) wrapInvocationAndHandleErrors(input *ResourceHandlerRequest, request *HandlerRequest) (progressEvent *ProgressEvent) {

	//Handle all panics from the resource handler, log the error and return a failed Wrapper event.
	defer func(event *HandlerRequest) {
		if r := recover(); r != nil {
			err := fmt.Errorf("%s in a %s action on a %s: %s\n%s", "Handler panic", event.Action, event.ResourceType, r, debug.Stack())

			// Log the Go stack trace for this panic'd goroutine.
			p.logger.Println(err.Error())
			if perr := p.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				log.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			progressEvent = DefaultFailureHandler(err, InternalFailure)
		}
	}(request)

	var e *ProgressEvent
	var err error
	switch request.Action {
	case create:
		e, err = p.customResource.CreateRequest(input, request.Context.CallbackContext)

	case delete:
		e, err = p.customResource.DeleteRequest(input, request.Context.CallbackContext)

	case list:
		e, err = p.customResource.ListRequest(input, request.Context.CallbackContext)

	case read:
		e, err = p.customResource.ReadRequest(input, request.Context.CallbackContext)

	case update:
		e, err = p.customResource.UpdateRequest(input, request.Context.CallbackContext)
	}

	if err == nil && e == nil {

		err = &errs.TerminalError{CustomerFacingErrorMessage: "Handler returned null"}
	}

	if err != nil {

		if aerr, ok := err.(awserr.Error); ok {
			if perr := p.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				p.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			p.logUnhandledError("A downstream service error occurred", request, err)
			return DefaultFailureHandler(aerr, GeneralServiceException)
		}
		switch err := err.(type) {
		case *errs.ResourceAlreadyExistsError:
			if perr := p.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				p.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			p.logUnhandledError("An existing resource was found", request, err)
			return DefaultFailureHandler(err, AlreadyExists)
		case *errs.ResourceNotFoundError:
			if perr := p.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				p.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			p.logUnhandledError("A requested resource was not found", request, err)
			return DefaultFailureHandler(err, NotFound)
		case *errs.TerminalError:
			if perr := p.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				p.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			p.logUnhandledError(err.CustomerFacingErrorMessage, request, err)
			return DefaultFailureHandler(err, InternalFailure)
		default:
			if perr := p.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				p.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			p.logUnhandledError("An unknown error occurred ", request, err)
			return DefaultFailureHandler(err, InternalFailure)
		}
	}
	return e
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

//ScheduleReinvocation manages scheduling of handler re-invocations.
func (p *Wrapper) scheduleReinvocation(c context.Context, req *HandlerRequest, hr *ProgressEvent, l *lambdacontext.LambdaContext) (handlerResult *ProgressEvent, result bool) {

	if hr.OperationStatus != InProgress {
		// no reinvoke required
		return hr, false
	}

	req.Context.Invocation = req.Context.Invocation + 1

	cbcx, err := json.Marshal(hr.CallbackContext)

	if err != nil {
		p.logger.Printf("Failed to schedule re-invoke, caused by %s", err.Error())
		hr.Message = err.Error()
		hr.OperationStatus = "FAILED"
		hr.HandlerErrorCode = "InternalFailure"
		return hr, false
	}

	req.Context.CallbackContext = json.RawMessage(cbcx)

	uID, err := scheduler.NewUUID()
	if err != nil {
		p.logger.Printf("Failed to schedule re-invoke, caused by %s", err.Error())
		hr.Message = err.Error()
		hr.OperationStatus = "FAILED"
		hr.HandlerErrorCode = "InternalFailure"
		return hr, false
	}

	rn := fmt.Sprintf("reinvoke-handler-%s", uID)
	tID := fmt.Sprintf("reinvoke-target-%s", uID)

	// record the CloudWatchEvents objects for cleanup on the callback
	req.Context.CloudWatchEventsRuleName = rn
	req.Context.CloudWatchEventsTargetID = tID

	rcx, err := json.Marshal(req.Context)

	if err != nil {
		p.logger.Printf("Failed to schedule re-invoke, caused by %s", err.Error())
		hr.Message = err.Error()
		hr.OperationStatus = "FAILED"
		hr.HandlerErrorCode = "InternalFailure"
		return hr, false
	}

	//when a handler requests a sub-minute callback delay, and if the lambda
	//invocation
	//has enough runtime (with 20% buffer), we can reschedule from a thread wait
	//otherwise we re-invoke through CloudWatchEvents which have a granularity of
	//minutes.
	deadline, _ := c.Deadline()
	secondsUnitDeadline := time.Until(deadline).Seconds()

	if hr.CallbackDelaySeconds < 60 && secondsUnitDeadline > float64(hr.CallbackDelaySeconds)*1.2 {

		p.logger.Printf("Scheduling re-invoke locally after %v seconds, with Context %s", hr.CallbackDelaySeconds, string(rcx))

		time.Sleep(time.Duration(hr.CallbackDelaySeconds) * time.Second)

		return hr, true
	}

	p.logger.Printf("Scheduling re-invoke with Context %s", string(rcx))

	rj, err := json.Marshal(req)

	if err := p.sch.RescheduleAfterMinutes(l.InvokedFunctionArn, hr.CallbackDelaySeconds, string(rj), time.Now(), uID, rn, tID); err != nil {
		p.logger.Printf("Failed to schedule re-invoke, caused by %s", err.Error())
		hr.Message = err.Error()
		hr.OperationStatus = "FAILED"
		hr.HandlerErrorCode = "InternalFailure"

	}
	return hr, false
}

//transform the the request into a resource handler.
//Using reflection, finds the type of th custom resource,
//Unmarshalls DesiredResource and PreviousResourceState, sets the field in the
//CustomHandler and returns a ResourceHandlerRequest.
func (p *Wrapper) transform(r HandlerRequest) *ResourceHandlerRequest {

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
		ClientRequestToken:        r.BearerToken,
		LogicalResourceIdentifier: r.Data.LogicalResourceID,
	}
}

//null-safe logger redirect
func (p *Wrapper) log(message string) {
	if p.logger != nil {
		p.logger.Println(message)
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
//    err := Wrapper.InjectCredentialsAndInvoke(req)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (p *Wrapper) InjectCredentialsAndInvoke(req request.Request) error {

	req.Config.Credentials = p.wrapperCreds
	err := req.Send()
	if err != nil {
		return err
	}

	return nil
}
