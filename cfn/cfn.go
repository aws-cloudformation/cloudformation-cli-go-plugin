package cfn

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/credentials"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/logging"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/metrics"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/scheduler"

	"github.com/aws/aws-lambda-go/lambda"
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

// makeEventFunc is the entry point to all invocations of a custom resource
func makeEventFunc(h Handler) eventFunc {
	return func(ctx context.Context, event *event) (response, error) {
		event.RequestContext.Session = credentials.SessionFromCredentialsProvider(&event.RequestData.CallerCredentials)
		platformSession := credentials.SessionFromCredentialsProvider(&event.RequestData.PlatformCredentials)
		metricsPublisher := metrics.New(cloudwatch.New(platformSession))
		metricsPublisher.SetResourceTypeName(event.ResourceType)
		invokeScheduler := scheduler.New(cloudwatchevents.New(platformSession))

		providerSession := credentials.SessionFromCredentialsProvider(&event.RequestData.ProviderCredentials)
		logsProvider, err := logging.NewCloudWatchLogsProvider(
			cloudwatchlogs.New(providerSession),
			event.RequestData.ProviderLogGroupName,
		)

		if err != nil {
			// we will log the error in the metric, but carry on.
			cfnErr := cfnerr.New(serviceInternalError, "Unable to complete request", err)
			metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
		}

		// set default logger to output to CWL in the provider account
		logging.SetProviderLogOutput(logsProvider)

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

		for {
			progEvt, err := invoke(handlerFn, request, &event.RequestContext, metricsPublisher, event.Action)
			if err != nil {
				cfnErr := cfnerr.New(serviceInternalError, "Unable to complete request; invoke error", err)
				metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
				return newFailedResponse(cfnErr, event.BearerToken), err
			}

			r, err := newResponse(&progEvt, event.BearerToken)
			if err != nil {
				cfnErr := cfnerr.New(serviceInternalError, "Unable to complete request; response error", err)
				metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
				return newFailedResponse(cfnErr, event.BearerToken), err
			}

			switch r.OperationStatus {
			case handler.Success:
				return r, nil

			case handler.Failed:
				return r, nil

			case handler.InProgress:
				customerCtx, delay := marshalCallback(&progEvt)

				invocationIDS, err := scheduler.GenerateCloudWatchIDS()
				if err != nil {
					cfnErr := cfnerr.New(serviceInternalError, "Unable to complete request; IDS error", err)
					metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
					return newFailedResponse(cfnErr, event.BearerToken), err
				}

				//Add IDs to recall the function with Cloudwatch events
				event.RequestContext.CloudWatchEventsRuleName = invocationIDS.Handler
				event.RequestContext.CloudWatchEventsTargetID = invocationIDS.Target

				callbackRequest, err := json.Marshal(event)
				if err != nil {
					cfnErr := cfnerr.New(serviceInternalError, "Unable to complete request; marshaling error", err)
					metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
					return newFailedResponse(cfnErr, event.BearerToken), err
				}

				scheResult, err := invokeScheduler.Reschedule(ctx, delay, string(callbackRequest), invocationIDS)
				if err != nil {
					cfnErr := cfnerr.New(serviceInternalError, "Unable to complete request; scheduler error", err)
					metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
					return newFailedResponse(cfnErr, event.BearerToken), err
				}

				//If not computing local, exit and return response
				if !scheResult.ComputeLocal {
					return r, nil
				}

				//Rebuild the context
				event.RequestContext.CallbackContext = customerCtx
			}
		}
	}
}
