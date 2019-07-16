package cft

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

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cft/errs"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cft/internal/callback"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cft/internal/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cft/internal/scheduler"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cft/proxy"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

//A list of vaild Cloudformation actions
const (
	create = "CREATE"
	delete = "DELETE"
	list   = "LIST"
	read   = "READ"
	update = "UPDATE"
)

// InvokeHandler is an interface that the custom resource must implement.
type InvokeHandler interface {
	CreateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error)
	DeleteRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error)
	ListRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error)
	ReadRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error)
	UpdateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error)
}

//Wrapper contains the dependencies off the Lambda function.
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
func (w *Wrapper) initialiseRuntime(req HandlerRequest) {

	u := url.URL{
		Scheme: "https",
		Host:   req.ResponseEndpoint,
	}

	//Set the caller credentials
	w.wrapperCreds = setWrapperCreds(req)

	// If null, we are not running a test.
	if w.cbak == nil {
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

		w.cbak = callback.New(cloudformation.New(cfsess))
	}

	// If null, we are not running a test.
	if w.metpub == nil || w.sch == nil {
		//Create a Cloudwatch events and Cloudwatch AWS session.
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(req.Region),
			Credentials: credentials.NewStaticCredentials(req.Data.PlatformCredentials.AccessKeyID, req.Data.PlatformCredentials.SecretAccessKey, req.Data.PlatformCredentials.SessionToken),
			Endpoint:    aws.String(u.String()),
		})

		if err != nil {
			panic(err)
		}
		w.metpub = metric.New(cloudwatch.New(sess), req.ResourceType)
		w.sch = scheduler.New(cloudwatchevents.New(sess))

	}
}

//HandleLambdaEvent is the main entry point for the lambda function.
// A response will be output on all paths, though CloudFormation will
// not block on invoking the handlers, but rather listen for callbacks
func (w *Wrapper) HandleLambdaEvent(ctx context.Context, request HandlerRequest) (lr HandlerResponse, e error) {

	//Handle all panics from the resource handler, log the error and return failed.
	defer func(event *HandlerRequest) {
		if r := recover(); r != nil {
			err := &errs.TerminalError{CustomerFacingErrorMessage: "Internal error"}

			// Log the Go stack trace for this panic'd goroutine.
			w.logger.Println(fmt.Sprintf("%s in a %s action on a %s: %s\n%s", "HandlerRequest panic", event.Action, event.ResourceType, r, debug.Stack()))
			lr = createProgressResponse(proxy.DefaultFailureHandler(err, proxy.InternalFailure), event.BearerToken)
			e = nil
		}
	}(&request)

	hr, err := w.processInvocation(ctx, request)

	if err != nil {
		// Exceptions are wrapped as a consistent error response to the caller (i.e;
		// CloudFormation)

		hr = proxy.DefaultFailureHandler(err, proxy.InternalFailure)

		if (reflect.DeepEqual(request, HandlerRequest{})) {
			if (reflect.DeepEqual(request.Data, RequestData{})) {
				hr.ResourceModel = request.Data.ResourceProperties
			}
			if perr := w.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				log.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}

		} else {
			if perr := w.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				log.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
		}
	}
	return createProgressResponse(hr, request.BearerToken), nil

}

//processInvocation process the request information and invokes the handler.
func (w *Wrapper) processInvocation(cx context.Context, req HandlerRequest) (pr *proxy.ProgressEvent, e error) {

	//Handle all panics from processInvocation, log the error and return a failed Wrapper event.
	defer func(event *HandlerRequest) {
		if r := recover(); r != nil {
			err := &errs.TerminalError{CustomerFacingErrorMessage: "Internal error"}

			// Log the Go stack trace for this panic'd goroutine.
			w.logger.Println(fmt.Sprintf("%s in a %s action on a %s: %s\n%s", "processInvocation panic", event.Action, event.ResourceType, r, debug.Stack()))
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

	w.initialiseRuntime(req)

	hr := &proxy.ProgressEvent{}

	//Get and set the lambda Context.
	lc, _ := lambdacontext.FromContext(cx)

	// transform the request object to pass to caller.
	resHanReq := w.transform(req)

	w.checkReinvoke(req.Context)

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
		ch := make(chan *proxy.ProgressEvent, 1)

		//start the timer
		st := time.Now()

		// Ask the goroutine to do some work for us.
		go func() {
			w.metpub.PublishInvocationMetric(time.Now(), req.Action)

			// Report the work is done.
			re := w.wrapInvocationAndHandleErrors(resHanReq, &req)

			ch <- re
		}()

		// Wait for the work to finish. If it takes too long move on.
		select {
		case d := <-ch:
			elapsed := time.Since(st)
			w.metpub.PublishDurationMetric(time.Now(), req.Action, elapsed.Seconds()*1e3)
			hr, computeLocally = w.scheduleReinvocation(cx, &req, d, lc)

		case <-ctx.Done():
			//handler failed to respond; shut it down.
			elapsed := time.Since(st)
			w.metpub.PublishDurationMetric(time.Now(), req.Action, elapsed.Seconds()*1e3)
			panic("Handler failed to respond")
		}

		// Report the progress status when in non-terminal state (i.e; InProgress) back to configured endpoint.
		w.cbak.ReportProgress(req.BearerToken, hr.HandlerErrorCode, hr.OperationStatus, hr.ResourceModel, hr.Message)

		if !computeLocally {
			break
		}

	}

	return hr, nil

}

// If this invocation was triggered by a 're-invoke' CloudWatch Event, clean it up.
func (w *Wrapper) checkReinvoke(context RequestContext) {

	if context.CloudWatchEventsRuleName != "" && context.CloudWatchEventsTargetID != "" {

		if err := w.sch.CleanupCloudWatchEvents(context.CloudWatchEventsRuleName, context.CloudWatchEventsTargetID); err != nil {

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

func (w *Wrapper) logUnhandledError(errorDescription string, request *HandlerRequest, e error) {
	w.logger.Printf("%s in a %s action on a %s: %s\n%s", errorDescription, request.Action, request.ResourceType, e.Error(), debug.Stack())
}

//WrapInvocationAndHandleErrors invokes the handler implementation for the request, and handles certain classes of errors and correctly map those to
//the appropriate HandlerErrorCode Also wraps the invocation in last-mile
//timing metrics.
func (w *Wrapper) wrapInvocationAndHandleErrors(input *ResourceHandlerRequest, request *HandlerRequest) (progressEvent *proxy.ProgressEvent) {

	//Handle all panics from the resource handler, log the error and return a failed Wrapper event.
	defer func(event *HandlerRequest) {
		if r := recover(); r != nil {
			err := fmt.Errorf("%s in a %s action on a %s: %s\n%s", "Handler panic", event.Action, event.ResourceType, r, debug.Stack())

			// Log the Go stack trace for this panic'd goroutine.
			w.logger.Println(err.Error())
			if perr := w.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				log.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			progressEvent = proxy.DefaultFailureHandler(err, proxy.InternalFailure)
		}
	}(request)

	var e *proxy.ProgressEvent
	var err error
	switch request.Action {
	case create:
		e, err = w.customResource.CreateRequest(input, request.Context.CallbackContext)

	case delete:
		e, err = w.customResource.DeleteRequest(input, request.Context.CallbackContext)

	case list:
		e, err = w.customResource.ListRequest(input, request.Context.CallbackContext)

	case read:
		e, err = w.customResource.ReadRequest(input, request.Context.CallbackContext)

	case update:
		e, err = w.customResource.UpdateRequest(input, request.Context.CallbackContext)
	}

	if err == nil && e == nil {

		err = &errs.TerminalError{CustomerFacingErrorMessage: "Handler returned null"}
	}

	if err != nil {

		if aerr, ok := err.(awserr.Error); ok {
			if perr := w.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				w.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			w.logUnhandledError("A downstream service error occurred", request, err)
			return proxy.DefaultFailureHandler(aerr, proxy.GeneralServiceException)
		}
		switch err := err.(type) {
		case *errs.ResourceAlreadyExistsError:
			if perr := w.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				w.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			w.logUnhandledError("An existing resource was found", request, err)
			return proxy.DefaultFailureHandler(err, proxy.AlreadyExists)
		case *errs.ResourceNotFoundError:
			if perr := w.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				w.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			w.logUnhandledError("A requested resource was not found", request, err)
			return proxy.DefaultFailureHandler(err, proxy.NotFound)
		case *errs.TerminalError:
			if perr := w.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				w.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			w.logUnhandledError(err.CustomerFacingErrorMessage, request, err)
			return proxy.DefaultFailureHandler(err, proxy.InternalFailure)
		default:
			if perr := w.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				w.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			w.logUnhandledError("An unknown error occurred ", request, err)
			return proxy.DefaultFailureHandler(err, proxy.InternalFailure)
		}
	}
	return e
}

func createProgressResponse(progressEvent *proxy.ProgressEvent, bearerToken string) HandlerResponse {

	return HandlerResponse{
		Message:         progressEvent.Message,
		OperationStatus: progressEvent.OperationStatus,
		ResourceModel:   progressEvent.ResourceModel,
		BearerToken:     bearerToken,
		ErrorCode:       progressEvent.HandlerErrorCode,
	}

}

//scheduleReinvocation manages scheduling of handler re-invocations.
func (w *Wrapper) scheduleReinvocation(c context.Context, req *HandlerRequest, hr *proxy.ProgressEvent, l *lambdacontext.LambdaContext) (handlerResult *proxy.ProgressEvent, result bool) {

	if hr.OperationStatus != proxy.InProgress {
		// no reinvoke required
		return hr, false
	}

	req.Context.Invocation = req.Context.Invocation + 1

	cbcx, err := json.Marshal(hr.CallbackContext)

	if err != nil {
		w.logger.Printf("Failed to schedule re-invoke, caused by %s", err.Error())
		hr.Message = err.Error()
		hr.OperationStatus = "FAILED"
		hr.HandlerErrorCode = "InternalFailure"
		return hr, false
	}

	req.Context.CallbackContext = json.RawMessage(cbcx)

	uID, err := scheduler.NewUUID()
	if err != nil {
		w.logger.Printf("Failed to schedule re-invoke, caused by %s", err.Error())
		hr.Message = err.Error()
		hr.OperationStatus = "FAILED"
		hr.HandlerErrorCode = "InternalFailure"
		return hr, false
	}

	rn := fmt.Sprintf("reinvoke-handler-%s", uID)
	tID := fmt.Sprintf("reinvoke-target-%s", uID)

	//Record the CloudWatchEvents objects for cleanup on the callback.
	req.Context.CloudWatchEventsRuleName = rn
	req.Context.CloudWatchEventsTargetID = tID

	rcx, err := json.Marshal(req.Context)

	if err != nil {
		w.logger.Printf("Failed to schedule re-invoke, caused by %s", err.Error())
		hr.Message = err.Error()
		hr.OperationStatus = "FAILED"
		hr.HandlerErrorCode = "InternalFailure"
		return hr, false
	}

	//When a handler requests a sub-minute callback delay, and if the lambda
	//invocation has enough runtime (with 20% buffer), we can reschedule from a thread wait
	//otherwise we re-invoke through CloudWatchEvents which have a granularity of
	//minutes.
	deadline, _ := c.Deadline()
	secondsUnitDeadline := time.Until(deadline).Seconds()

	if hr.CallbackDelaySeconds < 60 && secondsUnitDeadline > float64(hr.CallbackDelaySeconds)*1.2 {

		w.logger.Printf("Scheduling re-invoke locally after %v seconds, with Context %s", hr.CallbackDelaySeconds, string(rcx))

		time.Sleep(time.Duration(hr.CallbackDelaySeconds) * time.Second)

		return hr, true
	}

	w.logger.Printf("Scheduling re-invoke with Context %s", string(rcx))

	rj, err := json.Marshal(req)

	if err := w.sch.RescheduleAfterMinutes(l.InvokedFunctionArn, hr.CallbackDelaySeconds, string(rj), time.Now(), uID, rn, tID); err != nil {
		w.logger.Printf("Failed to schedule re-invoke, caused by %s", err.Error())
		hr.Message = err.Error()
		hr.OperationStatus = "FAILED"
		hr.HandlerErrorCode = "InternalFailure"

	}
	return hr, false
}

//transform transformsthe the request into a resource handler.
//Using reflection, finds the type of th custom resource,
//Unmarshalls DesiredResource and PreviousResourceState and the CallBackContext, sets the field in the
//CustomHandler and returns a ResourceHandlerRequest.
func (w *Wrapper) transform(r HandlerRequest) *ResourceHandlerRequest {

	// Custom resource struct.
	v := reflect.ValueOf(w.customResource)

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

//log is a null-safe logger redirect
func (w *Wrapper) log(message string) {
	if w.logger != nil {
		w.logger.Println(message)
	}
}
