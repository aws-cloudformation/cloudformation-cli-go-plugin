package cfn

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/callback"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/credentials"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/logging"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/metrics"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/scheduler"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

const (
	invalidRequestError  = "InvalidRequest"
	serviceInternalError = "ServiceInternal"
	unmarshalingError    = "UnmarshalingError"
	marshalingError      = "MarshalingError"
	validationError      = "Validation"
	timeoutError         = "Timeout"
	sessionNotFoundError = "SessionNotFound"
)

const (
	unknownAction = "UNKNOWN"
	createAction  = "CREATE"
	readAction    = "READ"
	updateAction  = "UPDATE"
	deleteAction  = "DELETE"
	listAction    = "LIST"
)

// MaxRetries is the number of times to try to call the Handler after it fails to respond.
var MaxRetries int = 3

// Timeout is the length of time to wait before giving up on a request.
var Timeout time.Duration = 60 * time.Second

// Handler is the interface that all resource providers must implement
//
// Each method of Handler maps directly to a CloudFormation action.
// Every action must return a progress event containing details of
// any actions that were undertaken by the resource provider
// or of any error that occurred during operation.
type Handler interface {
	Create(request handler.Request) handler.ProgressEvent
	Read(request handler.Request) handler.ProgressEvent
	Update(request handler.Request) handler.ProgressEvent
	Delete(request handler.Request) handler.ProgressEvent
	List(request handler.Request) handler.ProgressEvent
}

// InvokeScheduler is the interface that all reinvocation schedulers must implement
type InvokeScheduler interface {
	Reschedule(lambdaCtx context.Context, secsFromNow int64, callbackRequest string, invocationIDS *scheduler.ScheduleIDS) (*scheduler.Result, error)
}

// Start is the entry point called from a resource's main function
//
// We define two lambda entry points; MakeEventFunc is the entry point to all
// invocations of a custom resource and MakeTestEventFunc is the entry point that
// allows the CLI's contract testing framework to invoke the resource's CRUDL handlers.
func Start(h Handler) {

	// MODE is an environment variable that is set ONLY
	// when contract test are performed.
	if mode, ok := os.LookupEnv("MODE"); ok == true {
		if mode == "Test" {
			lambda.Start(makeTestEventFunc(h))

		} else {
			lambda.Start(makeEventFunc(h))
		}
	}
}

// Tags are stored as key/value paired strings
type tags map[string]string

// eventFunc is the function signature required to execute an event from the Lambda SDK
type eventFunc func(ctx context.Context, event *event) (response, error)

// testEventFunc is the function signature required to execute an event from the Lambda SDK
// and is only used in contract testing
type testEventFunc func(ctx context.Context, event *testEvent) (handler.ProgressEvent, error)

// handlerFunc is the signature required for all actions
type handlerFunc func(request handler.Request) handler.ProgressEvent

// router decides which handler should be invoked based on the action
// It will return a route or an error depending on the action passed in
func router(a string, h Handler) (handlerFunc, error) {
	// Figure out which action was called and have a "catch-all"
	switch a {
	case createAction:
		return h.Create, nil
	case readAction:
		return h.Read, nil
	case updateAction:
		return h.Update, nil
	case deleteAction:
		return h.Delete, nil
	case listAction:
		return h.List, nil
	default:
		// No action matched, we should fail and return an InvalidRequestErrorCode
		return nil, cfnerr.New(invalidRequestError, "No action/invalid action specified", nil)
	}
}

// Invoke handles the invocation of the handerFn.
func invoke(handlerFn handlerFunc, request handler.Request, metricsPublisher *metrics.Publisher, action string) (handler.ProgressEvent, error) {
	attempts := 0

	for {
		attempts++
		// Create a context that is both manually cancellable and will signal
		// a cancel at the specified duration.
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		//We always defer a cancel.
		defer cancel()

		// Create a channel to received a signal that work is done.
		ch := make(chan handler.ProgressEvent, 1)

		// Create a channel to received error.
		cherror := make(chan error, 1)

		// Ask the goroutine to do some work for us.
		go func() {
			//start the timer
			start := time.Now()
			if err := metricsPublisher.PublishInvocationMetric(time.Now(), string(action)); err != nil {
				cherror <- err
			}

			// Report the work is done.
			progEvt := handlerFn(request)

			elapsed := time.Since(start)

			if err := metricsPublisher.PublishDurationMetric(time.Now(), string(action), elapsed.Seconds()*1e3); err != nil {
				cherror <- err
			}

			ch <- progEvt
		}()

		// Wait for the work to finish. If it takes too long move on. If the function returns an error, signal the error channel.
		select {
		case e := <-cherror:
			cfnErr := cfnerr.New(timeoutError, "Handler error", e)
			metricsPublisher.PublishExceptionMetric(time.Now(), string(action), cfnErr)
			//The handler returned an error.
			return handler.ProgressEvent{}, e

		case d := <-ch:
			//Return the response from the handler.
			return d, nil

		case <-ctx.Done():
			if attempts == MaxRetries {
				log.Printf("Handler failed to respond, retrying... attempt: %v action: %s \n", attempts, action)
				//handler failed to respond.
				cfnErr := cfnerr.New(timeoutError, "Handler failed to respond in time", nil)
				metricsPublisher.PublishExceptionMetric(time.Now(), string(action), cfnErr)
				return handler.ProgressEvent{}, cfnErr
			}
			log.Printf("Handler failed to respond, retrying... attempt: %v action: %s \n", attempts, action)

		}
	}
}

func isMutatingAction(action string) bool {
	switch action {
	case createAction:
		return true
	case updateAction:
		return true
	case deleteAction:
		return true
	}
	return false
}

func translateStatus(operationStatus handler.Status) callback.Status {

	switch operationStatus {
	case handler.Success:
		return callback.Success
	case handler.Failed:
		return callback.Failed
	case handler.InProgress:
		return callback.InProgress
	default:
		return callback.UnknownStatus
	}

}

func processinvoke(handlerFn handlerFunc, event *event, request handler.Request, metricsPublisher *metrics.Publisher) handler.ProgressEvent {

	progEvt, err := invoke(handlerFn, request, metricsPublisher, event.Action)

	if err != nil {
		return handler.NewFailedEvent(err)
	}
	return progEvt

}

func reschedule(ctx context.Context, invokeScheduler InvokeScheduler, progEvt handler.ProgressEvent, event *event) (bool, error) {
	cusCtx, delay := marshalCallback(&progEvt)
	ids, err := scheduler.GenerateCloudWatchIDS()

	if err != nil {
		return false, err
	}
	// Add IDs to recall the function with Cloudwatch events
	event.RequestContext.CloudWatchEventsRuleName = ids.Handler
	event.RequestContext.CloudWatchEventsTargetID = ids.Target

	// Rebuild the context
	event.RequestContext.CallbackContext = cusCtx

	callbackRequest, err := json.Marshal(event)

	if err != nil {
		return false, err
	}
	scheResult, err := invokeScheduler.Reschedule(ctx, delay, string(callbackRequest), ids)

	if err != nil {
		return false, err
	}

	return scheResult.ComputeLocal, nil
}

// MakeEventFunc is the entry point to all invocations of a custom resource
func makeEventFunc(h Handler) eventFunc {
	return func(ctx context.Context, event *event) (response, error) {
		platformSession := credentials.SessionFromCredentialsProvider(&event.RequestData.PlatformCredentials)
		providerSession := credentials.SessionFromCredentialsProvider(&event.RequestData.ProviderCredentials)
		logsProvider, err := logging.NewCloudWatchLogsProvider(
			cloudwatchlogs.New(providerSession),
			event.RequestData.ProviderLogGroupName,
		)

		// Set default logger to output to CWL in the provider account
		logging.SetProviderLogOutput(logsProvider)

		metricsPublisher := metrics.New(cloudwatch.New(platformSession), event.AWSAccountID)
		metricsPublisher.SetResourceTypeName(event.ResourceType)
		callbackAdapter := callback.New(cloudformation.New(platformSession), event.BearerToken)
		invokeScheduler := scheduler.New(cloudwatchevents.New(platformSession))
		re := newReportErr(callbackAdapter, metricsPublisher)

		handlerFn, err := router(event.Action, h)

		if err != nil {
			return re.report(event, "router error", err, serviceInternalError)
		}

		if err := validateEvent(event); err != nil {
			return re.report(event, "validation error", err, invalidRequestError)
		}

		// If this invocation was triggered by a 're-invoke' CloudWatch Event, clean it up.
		if event.RequestContext.CallbackContext != nil {
			err := invokeScheduler.CleanupEvents(event.RequestContext.CloudWatchEventsRuleName, event.RequestContext.CloudWatchEventsTargetID)

			if err != nil {
				// We will log the error in the metric, but carry on.
				cfnErr := cfnerr.New(serviceInternalError, "Cloudwatch Event clean up error", err)
				metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
			}
		}

		if len(event.RequestContext.CallbackContext) == 0 || event.RequestContext.Invocation == 0 {
			// Acknowledge the task for first time invocation.
			if err := callbackAdapter.ReportInitialStatus(); err != nil {
				return re.report(event, "callback initial report error", err, serviceInternalError)
			}
		}

		re.setPublishSatus(true)
		for {
			request := handler.NewRequest(
				event.RequestData.LogicalResourceID,
				event.RequestContext.CallbackContext,
				credentials.SessionFromCredentialsProvider(&event.RequestData.CallerCredentials),
				event.RequestData.PreviousResourceProperties,
				event.RequestData.ResourceProperties,
			)
			event.RequestContext.Invocation = event.RequestContext.Invocation + 1

			progEvt := processinvoke(handlerFn, event, request, metricsPublisher)

			cusCtx, delay := marshalCallback(&progEvt)

			r, err := newResponse(&progEvt, event.BearerToken)

			if err != nil {
				return re.report(event, "Response error", err, unmarshalingError)
			}

			log.Printf("Handler returned  OperationStatus: %v Message: %v CallbackContext: %v Delay: %v, ErrorCode: %v  ",
				r.OperationStatus, progEvt.Message,
				cusCtx, delay, progEvt.HandlerErrorCode)

			if !isMutatingAction(event.Action) && r.OperationStatus == handler.InProgress {
				return re.report(event, "Response error", errors.New("READ and LIST handlers must return synchronous"), invalidRequestError)
			}

			if isMutatingAction(event.Action) {
				callbackAdapter.ReportStatus(translateStatus(progEvt.OperationStatus), event.RequestData.ResourceProperties, progEvt.Message, string(r.ErrorCode))
			}

			switch r.OperationStatus {
			case handler.InProgress:
				local, err := reschedule(ctx, invokeScheduler, progEvt, event)

				if err != nil {
					return re.report(event, "Reschedule error", err, serviceInternalError)
				}

				// If not computing local, exit and return response.
				if !local {
					return r, nil
				}
			default:
				return r, nil
			}

		}
	}
}

// MakeTestEventFunc is the entry point that allows the CLI's
// contract testing framework to invoke the resource's CRUDL handlers.
func makeTestEventFunc(h Handler) testEventFunc {
	return func(ctx context.Context, event *testEvent) (handler.ProgressEvent, error) {

		handlerFn, err := router(event.Action, h)

		if err != nil {
			return handler.NewFailedEvent(err), err
		}

		request := handler.NewRequest(
			event.Request.LogicalResourceIdentifier,
			event.CallbackContext,
			credentials.SessionFromCredentialsProvider(&event.Credentials),
			event.Request.PreviousResourceState,
			event.Request.DesiredResourceState,
		)

		progEvt := handlerFn(request)

		return progEvt, nil
	}
}
