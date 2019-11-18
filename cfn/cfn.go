package cfn

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/callback"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/credentials"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/logging"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/metrics"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/scheduler"

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
	Create(ctx context.Context, request handler.Request) handler.ProgressEvent
	Read(ctx context.Context, request handler.Request) handler.ProgressEvent
	Update(ctx context.Context, request handler.Request) handler.ProgressEvent
	Delete(ctx context.Context, request handler.Request) handler.ProgressEvent
	List(ctx context.Context, request handler.Request) handler.ProgressEvent
}

// Start is the entry point called from a resource's main function
func Start(h Handler) {
	lambda.Start(makeEventFunc(h))
}

// Tags are stored as key/value paired strings
type tags map[string]string

// eventFunc is the function signature required to execute an event from the Lambda SDK
type eventFunc func(ctx context.Context, event *event) (response, error)

// handlerFunc is the signature required for all actions
type handlerFunc func(ctx context.Context, request handler.Request) handler.ProgressEvent

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

//Invoke handles the invocation of the handerFn.
func invoke(handlerFn handlerFunc, request handler.Request, reqContext *requestContext, metricsPublisher *metrics.Publisher, action string) (handler.ProgressEvent, error) {
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

			customerCtx := SetContextValues(context.Background(), reqContext.CallbackContext)
			customerCtx = SetContextSession(customerCtx, reqContext.Session)

			// Report the work is done.
			progEvt := handlerFn(customerCtx, request)

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

func reportInitialStatus(event *event, metricsPublisher *metrics.Publisher, callbackAdapter *callback.CloudFormationCallbackAdapter) error {
	if err := callbackAdapter.ReportProgress(event.BearerToken, "", string(handler.InProgress), string(handler.Pending), "", ""); err != nil {
		cfnErr := cfnerr.New(serviceInternalError, "Unable to complete request; Callback falure", err)
		metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
		return cfnErr
	}
	return nil
}

func reportFailureStatus(event *event, metricsPublisher *metrics.Publisher, callbackAdapter *callback.CloudFormationCallbackAdapter, model string) error {
	if err := callbackAdapter.ReportProgress(event.BearerToken, cfnerr.InvalidRequest, string(handler.Failed), string(handler.InProgress), model, "Unable to complete request"); err != nil {
		cfnErr := cfnerr.New(serviceInternalError, "Unable to complete request; Callback falure", err)
		metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
		return cfnErr
	}
	return nil
}

// makeEventFunc is the entry point to all invocations of a custom resource
func makeEventFunc(h Handler) eventFunc {
	return func(ctx context.Context, event *event) (response, error) {
		platformSession := credentials.SessionFromCredentialsProvider(&event.RequestData.PlatformCredentials)
		providerSession := credentials.SessionFromCredentialsProvider(&event.RequestData.ProviderCredentials)
		logsProvider, err := logging.NewCloudWatchLogsProvider(
			cloudwatchlogs.New(providerSession),
			event.RequestData.ProviderLogGroupName,
		)

		// set default logger to output to CWL in the provider account
		logging.SetProviderLogOutput(logsProvider)

		metricsPublisher := metrics.New(cloudwatch.New(platformSession))
		metricsPublisher.SetResourceTypeName(event.ResourceType)
		invokeScheduler := scheduler.New(cloudwatchevents.New(platformSession))
		callbackAdapter := callback.New(cloudformation.New(platformSession))

		handlerFn, err := router(event.Action, h)
		if err != nil {
			cfnErr := cfnerr.New(serviceInternalError, "Unable to complete request; router error", err)
			metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
			return newFailedResponse(cfnErr, event.BearerToken), cfnErr
		}

		if err := validateEvent(event); err != nil {
			cfnErr := cfnerr.New(invalidRequestError, "Failed to validate input; validation error", err)
			metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
			return newFailedResponse(cfnErr, event.BearerToken), cfnErr
		}

		request := handler.NewRequest(
			event.RequestData.PreviousResourceProperties,
			event.RequestData.ResourceProperties,
			event.RequestData.LogicalResourceID,
		)

		if len(event.RequestContext.CallbackContext) == 0 || event.RequestContext.Invocation == 0 {
			// Acknowledge the task for first time invocation
			if err := reportInitialStatus(event, metricsPublisher, callbackAdapter); err != nil {
				return newFailedResponse(err, event.BearerToken), err
			}
		}

		// If this invocation was triggered by a 're-invoke' CloudWatch Event, clean it up.
		if event.RequestContext.CallbackContext != nil {
			err := invokeScheduler.CleanupEvents(event.RequestContext.CloudWatchEventsRuleName, event.RequestContext.CloudWatchEventsTargetID)

			if err != nil {
				// we will log the error in the metric, but carry on.
				cfnErr := cfnerr.New(serviceInternalError, "Unable to complete request", err)
				metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
			}
		}

		for {
			event.RequestContext.Session = credentials.SessionFromCredentialsProvider(&event.RequestData.CallerCredentials)
			event.RequestContext.Invocation = event.RequestContext.Invocation + 1

			progEvt, err := invoke(handlerFn, request, &event.RequestContext, metricsPublisher, event.Action)

			if err != nil {
				errs := []error{err}
				if reportErr := reportFailureStatus(event, metricsPublisher, callbackAdapter, ""); reportErr != nil {
					errs = append(errs, reportErr)
				}
				cfnErr := cfnerr.NewBatchError(serviceInternalError, "Unable to complete request; invoke error", errs)
				metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
				return newFailedResponse(cfnErr, event.BearerToken), err
			}

			r, err := newResponse(&progEvt, event.BearerToken)

			if err != nil {
				errs := []error{err}
				if reportErr := reportFailureStatus(event, metricsPublisher, callbackAdapter, ""); reportErr != nil {
					errs = append(errs, reportErr)
				}
				cfnErr := cfnerr.NewBatchError(serviceInternalError, "Unable to complete request; invoke error", errs)
				metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
				return newFailedResponse(cfnErr, event.BearerToken), err
			}

			modelString, err := json.Marshal(r.ResourceModel)

			if err != nil {
				errs := []error{err}
				if reportErr := reportFailureStatus(event, metricsPublisher, callbackAdapter, ""); reportErr != nil {
					errs = append(errs, reportErr)
				}
				cfnErr := cfnerr.NewBatchError(serviceInternalError, "Unable to complete request; invoke error", errs)
				metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
				return newFailedResponse(cfnErr, event.BearerToken), err
			}

			if !isMutatingAction(event.Action) && r.OperationStatus == "PENDING" {
				errs := []error{err}
				if reportErr := reportFailureStatus(event, metricsPublisher, callbackAdapter, string(modelString)); reportErr != nil {
					errs = append(errs, reportErr)
				}
				cfnErr := cfnerr.NewBatchError(serviceInternalError, "READ and LIST handlers must return synchronously", errs)
				metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
				return newFailedResponse(cfnErr, event.BearerToken), err
			}

			if isMutatingAction(event.Action) {
				callbackAdapter.ReportProgress(event.BearerToken, progEvt.HandlerErrorCode, string(progEvt.OperationStatus), string(handler.InProgress), string(modelString), progEvt.Message)
			}
			switch r.OperationStatus {

			case handler.InProgress:
				customerCtx, delay := marshalCallback(&progEvt)

				invocationIDS, err := scheduler.GenerateCloudWatchIDS()
				if err != nil {
					errs := []error{err}
					if reportErr := reportFailureStatus(event, metricsPublisher, callbackAdapter, string(modelString)); reportErr != nil {
						errs = append(errs, reportErr)
					}
					cfnErr := cfnerr.NewBatchError(serviceInternalError, "Unable to complete request; IDS error", errs)
					metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
					return newFailedResponse(cfnErr, event.BearerToken), err
				}

				//Add IDs to recall the function with Cloudwatch events
				event.RequestContext.CloudWatchEventsRuleName = invocationIDS.Handler
				event.RequestContext.CloudWatchEventsTargetID = invocationIDS.Target

				//Set the session to nil to prevent marshaling
				event.RequestContext.Session = nil

				//Rebuild the context
				event.RequestContext.CallbackContext = customerCtx

				callbackRequest, err := json.Marshal(event)
				if err != nil {
					errs := []error{err}
					if reportErr := reportFailureStatus(event, metricsPublisher, callbackAdapter, string(modelString)); reportErr != nil {
						errs = append(errs, reportErr)
					}
					cfnErr := cfnerr.NewBatchError(serviceInternalError, "Unable to complete request; marshaling error", errs)
					metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
					return newFailedResponse(cfnErr, event.BearerToken), err
				}

				scheResult, err := invokeScheduler.Reschedule(ctx, delay, string(callbackRequest), invocationIDS)
				if err != nil {
					errs := []error{err}
					if reportErr := reportFailureStatus(event, metricsPublisher, callbackAdapter, string(modelString)); reportErr != nil {
						errs = append(errs, reportErr)
					}
					cfnErr := cfnerr.NewBatchError(serviceInternalError, "Unable to complete request; scheduler error", errs)
					metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
					return newFailedResponse(cfnErr, event.BearerToken), err
				}

				//If not computing local, exit and return response
				if !scheResult.ComputeLocal {
					return r, nil
				}
			default:
				return r, nil
			}
		}
	}
}
