package cfn

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

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/errs"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/internal/callback"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/internal/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/internal/platform/injection/provider"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/internal/scheduler"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/proxy"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

//Vaild Cloudformation actions.
const (
	create = "CREATE"
	delete = "DELETE"
	list   = "LIST"
	read   = "READ"
	update = "UPDATE"
)

// InvokeHandler is an interface that the custom resource must implement.
type InvokeHandler interface {
	CreateRequest(request *proxy.ResourceHandlerRequest, proxyClient *proxy.AWSClientProxy) (*proxy.ProgressEvent, error)
	DeleteRequest(request *proxy.ResourceHandlerRequest, proxyClient *proxy.AWSClientProxy) (*proxy.ProgressEvent, error)
	ListRequest(request *proxy.ResourceHandlerRequest, proxyClient *proxy.AWSClientProxy) (*proxy.ProgressEvent, error)
	ReadRequest(request *proxy.ResourceHandlerRequest, proxyClient *proxy.AWSClientProxy) (*proxy.ProgressEvent, error)
	UpdateRequest(request *proxy.ResourceHandlerRequest, proxyClient *proxy.AWSClientProxy) (*proxy.ProgressEvent, error)
}

//Wrapper contains the dependencies off the Lambda function.
type Wrapper struct {
	metpub         *metric.Publisher
	sch            *scheduler.CloudWatchScheduler
	cbak           *callback.CloudFormationCallbackAdapter
	customResource InvokeHandler
	creds          *credentials.Provider
	logger         *log.Logger
	ContextType    reflect.Type
}

//initialiseRuntime initialises the dependencies.
func (w *Wrapper) initialiseRuntime(creds *Credentials, u *url.URL) {

	// initialisation skipped if these dependencies were set during injection (in
	// test)
	cp := provider.NewPlatformCredentialsProvider(creds.AccessKeyID, creds.SecretAccessKey, creds.AccessKeyID)
	cfp := provider.NewCloudFormationProvider(cp)
	cfp.SetCallbackEndpoint(u)

	// If null, we are not running a test.
	if w.cbak == nil {
		w.cbak = callback.New(cfp)
		w.cbak.RefreshClient()
	}
	cwp := provider.NewCloudWatchProvider(cp)
	if w.metpub == nil {
		w.metpub = metric.New(cwp)
		w.metpub.RefreshClient()
	}
	cwe := provider.NewCloudWatchEventsProvider(cp)
	if w.sch == nil {
		w.sch = scheduler.New(cwe)
		w.sch.RefreshClient()
	}
}

//HandleLambdaEvent is the main entry point for the lambda function.
// A response will be output on all paths, though CloudFormation will
// not block on invoking the handlers, but rather listen for callbacks.
func (w *Wrapper) HandleLambdaEvent(ctx context.Context, request HandlerRequest) (lr HandlerResponse, e error) {

	//All panics from the resource handler are logged and a fail progressEvent is returned.
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
		hr = proxy.DefaultFailureHandler(err, proxy.InternalFailure)
		w.logger.Println(fmt.Sprintf("%s in a %s action on a %s: n%s", "HandlerRequest panic", request.Action, request.ResourceType, debug.Stack()))

		if (!reflect.DeepEqual(request, HandlerRequest{})) {
			if (!reflect.DeepEqual(request.Data, RequestData{})) {
				hr.ResourceModel = request.Data.ResourceProperties
			}
			if perr := w.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				w.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}

		} else {
			if perr := w.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				w.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
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
			pr = proxy.DefaultFailureHandler(err, proxy.InternalFailure)
			if perr := w.metpub.PublishExceptionMetric(time.Now(), event.Action, err); perr != nil {
				w.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			e = nil
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
	cred := req.Data.PlatformCredentials
	u := url.URL{
		Scheme: "https",
		Host:   req.ResponseEndpoint,
	}
	w.metpub.SetResourceTypeName(req.ResourceType)
	w.initialiseRuntime(&cred, &u)
	hr := &proxy.ProgressEvent{}

	//Get and set the lambda Context.
	lc, _ := lambdacontext.FromContext(cx)

	// Transform the request object to pass to caller.
	resHanReq := w.transform(req)
	w.checkReinvoke(req.Context)

	// Set a duration.
	duration := 60 * time.Second
	computeLocally := true
	for {
		// Create a context that is both manually cancellable and will signal
		// a cancel at the specified duration.
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		//We always defer a cancel.
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
		//w.cbak.ReportProgress(req.BearerToken, hr.HandlerErrorCode, hr.OperationStatus, hr.ResourceModel, hr.Message)
		if !computeLocally {
			break
		}

		//Rebuild the CallContext.
		v := reflect.ValueOf(w.customResource)

		// Custom resource DesiredResourceState struct.
		cv := v.Elem().FieldByName("CallBackContext")

		//Check if the field is found and that it's a strut value.
		if !cv.IsValid() || cv.Kind() != reflect.Struct {
			panic("Unable to find CallbackContext in customResource")
		}

		cv.Set(reflect.ValueOf(hr.CallbackContext))

	}

	return hr, nil
}

// checkReinvoke : if this invocation was triggered by a 're-invoke' CloudWatch Event, clean it up.
func (w *Wrapper) checkReinvoke(context RequestContext) {
	if context.CloudWatchEventsRuleName != "" && context.CloudWatchEventsTargetID != "" {
		if err := w.sch.CleanupCloudWatchEvents(context.CloudWatchEventsRuleName, context.CloudWatchEventsTargetID); err != nil {
			panic(err)
		}
	}
}

//validateResourceProperties validates if the request has resource properties.
func validateResourceProperties(in json.RawMessage) error {
	dst := new(bytes.Buffer)
	err := json.Compact(dst, []byte(in))
	if err != nil {
		return &errs.TerminalError{CustomerFacingErrorMessage: "Invalid resource properties object received"}
	}
	if dst.String() == "{}" {
		return &errs.TerminalError{CustomerFacingErrorMessage: "Invalid resource properties object received"}
	}
	return nil
}

func (w *Wrapper) logUnhandledError(errorDescription string, request *HandlerRequest, e error) {
	w.logger.Printf("%s in a %s action on a %s: %s\n%s", errorDescription, request.Action, request.ResourceType, e.Error(), debug.Stack())
}

//WrapInvocationAndHandleErrors invokes the handler implementation for the request, and handles certain classes of errors and correctly map those to
//the appropriate HandlerErrorCode Also wraps the invocation in last-mile
//timing metrics.
func (w *Wrapper) wrapInvocationAndHandleErrors(input *proxy.ResourceHandlerRequest, request *HandlerRequest) (progressEvent *proxy.ProgressEvent) {

	//Handle all panics from the resource handler, log the error and return a failed Wrapper event.
	defer func(event *HandlerRequest) {
		if r := recover(); r != nil {
			err := fmt.Errorf("%s in a %s action on a %s: %s\n%s", "Handler panic", event.Action, event.ResourceType, r, debug.Stack())

			// Log the Go stack trace for this panic'd goroutine.
			w.logger.Println(err.Error())
			if perr := w.metpub.PublishExceptionMetric(time.Now(), request.Action, err); perr != nil {
				w.logger.Printf("%s : %s", "Publish error metric failed ", perr.Error())
			}
			progressEvent = proxy.DefaultFailureHandler(err, proxy.InternalFailure)
		}
	}(request)
	var e *proxy.ProgressEvent
	var err error
	switch request.Action {
	case create:
		e, err = w.customResource.CreateRequest(input, nil)

	case delete:
		e, err = w.customResource.DeleteRequest(input, nil)

	case list:
		e, err = w.customResource.ListRequest(input, nil)

	case read:
		e, err = w.customResource.ReadRequest(input, nil)

	case update:
		e, err = w.customResource.UpdateRequest(input, nil)
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

//transform transformst he the request into a resource handler.
//Using reflection, finds the type of th custom resource,
//Unmarshalls DesiredResource and PreviousResourceState and the CallBackContext, sets the field in the
//CustomHandler and returns a ResourceHandlerRequest.
func (w *Wrapper) transform(r HandlerRequest) *proxy.ResourceHandlerRequest {

	// Custom resource struct.
	v := reflect.ValueOf(w.customResource)

	// Custom resource DesiredResourceState struct.
	dv := v.Elem().FieldByName("DesiredResourceState")

	//Check if the field is found and that it's a strut value.
	if !dv.IsValid() || dv.Kind() != reflect.Struct {
		panic("Unable to find DesiredResource in customResource")
	}

	// Custom resource PreviousResourceState struct.
	pv := v.Elem().FieldByName("PreviousResourceState")

	//Check if the field is found and that it's a strut value.
	if !pv.IsValid() || pv.Kind() != reflect.Struct {
		panic("Unable to find PreviousResource in customResource")
	}

	// Custom resource DesiredResourceState struct.
	cv := v.Elem().FieldByName("CallBackContext")

	//Check if the field is found and that it's a strut value.
	if !cv.IsValid() || cv.Kind() != reflect.Struct {
		panic("Unable to find CallbackContext in customResource")
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

	//Set the customResource.
	pv.Set(pr.Elem())

	//Create new callBackContext.
	cr := reflect.New(cv.Type())

	//Set the CallBackContext type in wrapper so that it can be used to rebuild the context
	//when the CallbackDelaySeconds is < 60 seconds.

	w.ContextType = cv.Type()

	//Try to unmarshhal the into the strut field.
	if r.Data.PreviousResourceProperties != nil {
		if err := json.Unmarshal([]byte(r.Data.PreviousResourceProperties), cr.Interface()); err != nil {
			panic(err)
		}
	}

	//Set the callBackContext.
	cv.Set(cr.Elem())

	return &proxy.ResourceHandlerRequest{
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
