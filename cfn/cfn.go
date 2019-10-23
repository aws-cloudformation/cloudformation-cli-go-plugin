// Package cfn defines the common interfaces and values used by the RPDK
package cfn

import (
	"context"
	"log"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/credentials"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/metrics"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/scheduler"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

const (
	InvalidRequestError  string = "InvalidRequest"
	ServiceInternalError string = "ServiceInternal"
	UnmarshalingError    string = "UnmarshalingError"
	MarshalingError      string = "MarshalingError"
	ValidationError      string = "Validation"
	TimeoutError         string = "Timeout"

	//MaxRetries is the number of times to try to call the Handler after it fails to respond.
	MaxRetries int = 3
)

// Timeout is the length of time to wait before giving up on a request.
var Timeout time.Duration = 60 * time.Second

// Handler represents the actions from the AWS CloudFormation service
//
// Each action maps directly to a CloudFormation action. Every action is
// expected to return a response and/or an error.
//
// A valid error condition would be met if the resource operation failed or
// an API is no longer available.
type Handler interface {
	Create(ctx context.Context, request handler.Request) (handler.ProgressEvent, error)
	Read(ctx context.Context, request handler.Request) (handler.ProgressEvent, error)
	Update(ctx context.Context, request handler.Request) (handler.ProgressEvent, error)
	Delete(ctx context.Context, request handler.Request) (handler.ProgressEvent, error)
	List(ctx context.Context, request handler.Request) (handler.ProgressEvent, error)
}

// Start is the entry point called from a resource's lambda function
func Start(h Handler) {
	lambda.Start(makeEventFunc(h))
}

// Tags are stored as key/value paired strings
type tags map[string]string

// eventFunc is the function signature required to execute an event from the Lambda SDK
type eventFunc func(ctx context.Context, event *event) (handler.Response, error)

// handlerFunc is the signature required for all actions
type handlerFunc func(ctx context.Context, request handler.Request) (handler.ProgressEvent, error)

// router decides which handler should be invoked based on the action
// It will return a route or an error depending on the action passed in
func router(a handler.Action, h Handler) (handlerFunc, error) {
	// Figure out which action was called and have a "catch-all"
	switch a {
	case handler.Create:
		return h.Create, nil
	case handler.Read:
		return h.Read, nil
	case handler.Update:
		return h.Update, nil
	case handler.Delete:
		return h.Update, nil
	case handler.List:
		return h.List, nil
	default:
		// No action matched, we should fail and return an InvalidRequestErrorCode
		return nil, cfnerr.New(InvalidRequestError, "No action/invalid action specified", nil)
	}
}

//Invoke handles the invocation of the handerFn.
func invoke(handlerFn handlerFunc, request handler.Request, reqContext *requestContext, metricsPublisher *metrics.Publisher, action handler.Action) (handler.ProgressEvent, error) {
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

			customerCtx := handler.ContextValues(context.Background(), reqContext.CallbackContext)
			customerCtx = handler.ContextInjectSession(customerCtx, reqContext.GetSession())

			// Report the work is done.
			progEvt, err := handlerFn(customerCtx, request)
			if err != nil {
				cherror <- err
			}

			elapsed := time.Since(start)

			if err := metricsPublisher.PublishDurationMetric(time.Now(), string(action), elapsed.Seconds()*1e3); err != nil {
				cherror <- err
			}

			ch <- progEvt
		}()

		// Wait for the work to finish. If it takes too long move on. If the function returns an error, signal the error channel.
		select {
		case e := <-cherror:
			cfnErr := cfnerr.New(TimeoutError, "Handler error", e)
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
				cfnErr := cfnerr.New(TimeoutError, "Handler failed to respond in time", nil)
				metricsPublisher.PublishExceptionMetric(time.Now(), string(action), cfnErr)
				return handler.ProgressEvent{}, cfnErr
			}
			log.Printf("Handler failed to respond, retrying... attempt: %v action: %s \n", attempts, action)

		}
	}
}

// makeEventFunc is the entry point to all invocations of a custom resource
func makeEventFunc(h Handler) eventFunc {
	return func(ctx context.Context, event *event) (handler.Response, error) {
		platformSession := credentials.SessionFromCredentialsProvider(event.RequestData.PlatformCredentials)
		metricsPublisher := metrics.New(cloudwatch.New(platformSession))
		metricsPublisher.SetResourceTypeName(event.ResourceType)
		invokeScheduler := scheduler.New(cloudwatchevents.New(platformSession))
		var resp handler.Response

		handlerFn, err := router(event.Action, h)
		if err != nil {
			cfnErr := cfnerr.New(ServiceInternalError, "Unable to complete request", err)
			metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
			return handler.NewFailedResponse(cfnErr), cfnErr
		}

		if err := validateEvent(event); err != nil {
			cfnErr := cfnerr.New(InvalidRequestError, "Failed to validate input", err)
			metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
			return handler.NewFailedResponse(cfnErr), cfnErr
		}

		request := handler.NewRequest(
			event.RequestData.PreviousResourceProperties,
			event.RequestData.ResourceProperties,
			event.RequestData.LogicalResourceID,
			event.BearerToken,
		)
		for {
			progEvt, err := invoke(handlerFn, request, event.Context, metricsPublisher, event.Action)

			if err != nil {
				cfnErr := cfnerr.New(ServiceInternalError, "Unable to complete request", err)
				metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
				return handler.NewFailedResponse(cfnErr), err
			}

			r, err := progEvt.MarshalResponse()
			if err != nil {
				cfnErr := cfnerr.New(ServiceInternalError, "Unable to complete request", err)
				metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
				return handler.NewFailedResponse(cfnErr), err
			}

			switch r.OperationStatus() {
			case handler.Success:
				return r, nil
			case handler.Failed:
				return r, nil
			case handler.InProgress:

				customerCtx, delay := progEvt.MarshalCallback()

				invocationIDS, err := scheduler.GenerateCloudWatchIDS()
				if err != nil {
					cfnErr := cfnerr.New(ServiceInternalError, "Unable to complete request", err)
					metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
					return handler.NewFailedResponse(cfnErr), err
				}

				//Add IDs to recall the function with Cloudwatch events
				event.Context.CloudWatchEventsRuleName = invocationIDS.Handler
				event.Context.CloudWatchEventsTargetID = invocationIDS.Target

				callbackRequest, err := event.MarshalJSON()

				if err != nil {
					cfnErr := cfnerr.New(ServiceInternalError, "Unable to complete request", err)
					metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
					return handler.NewFailedResponse(cfnErr), err
				}

				scheResult, err := invokeScheduler.Reschedule(ctx, delay, string(callbackRequest), invocationIDS)

				if err != nil {
					cfnErr := cfnerr.New(ServiceInternalError, "Unable to complete request", err)
					metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), cfnErr)
					return handler.NewFailedResponse(cfnErr), err
				}

				//If not computing local, exit and return response
				if !scheResult.ComputeLocal {
					return r, nil
				}

				//Rebuild the context
				event.Context.CallbackContext = customerCtx

			}

		}

		return resp, nil
	}
}
