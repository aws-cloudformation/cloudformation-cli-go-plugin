package cfn

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/credentials"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/logging"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/metrics"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
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

// Start is the entry point called from a resource's main function
//
// We define two lambda entry points; MakeEventFunc is the entry point to all
// invocations of a custom resource and MakeTestEventFunc is the entry point that
// allows the CLI's contract testing framework to invoke the resource's CRUDL handlers.
func Start(h Handler) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Handler panicked: %s", r)
			panic(r) // Continue the panic
		}
	}()

	// MODE is an environment variable that is set ONLY
	// when contract test are performed.
	mode, _ := os.LookupEnv("MODE")

	if mode == "Test" {
		log.Printf("Handler starting in test mode")
		lambda.Start(makeTestEventFunc(h))
	} else {
		log.Printf("Handler starting")
		lambda.Start(makeEventFunc(h))
	}

	log.Printf("Handler finished")
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

// MakeEventFunc is the entry point to all invocations of a custom resource
func makeEventFunc(h Handler) eventFunc {
	return func(ctx context.Context, event *event) (response, error) {
		//pls := credentials.SessionFromCredentialsProvider(&event.RequestData.PlatformCredentials)
		ps := credentials.SessionFromCredentialsProvider(&event.RequestData.ProviderCredentials)
		l, err := logging.NewCloudWatchLogsProvider(
			cloudwatchlogs.New(ps),
			event.RequestData.ProviderLogGroupName,
		)
		// Set default logger to output to CWL in the provider account
		logging.SetProviderLogOutput(l)
		m := metrics.New(cloudwatch.New(ps), event.AWSAccountID, event.ResourceType)
		re := newReportErr(m)
		if err := scrubFiles("/tmp"); err != nil {
			log.Printf("Error: %v", err)
			m.PublishExceptionMetric(time.Now(), event.Action, err)
		}
		handlerFn, err := router(event.Action, h)
		log.Printf("Handler received the %s action", event.Action)
		if err != nil {
			return re.report(event, "router error", err, serviceInternalError)
		}
		if err := validateEvent(event); err != nil {
			return re.report(event, "validation error", err, invalidRequestError)
		}
		request := handler.NewRequest(
			event.RequestData.LogicalResourceID,
			event.NextToken,
			event.StackID,
			event.RequestData.StackTags,
			event.Region,
			event.AWSAccountID,
			event.RequestData.SystemTags,
			event.CallbackContext,
			credentials.SessionFromCredentialsProvider(&event.RequestData.CallerCredentials),
			event.RequestData.PreviousResourceProperties,
			event.RequestData.ResourceProperties,
		)
		p := invoke(handlerFn, request, m, event.Action)
		r, err := newResponse(&p, event.BearerToken)
		if err != nil {
			log.Printf("Error creating response: %v", err)
			return re.report(event, "Response error", err, unmarshalingError)
		}
		if !isMutatingAction(event.Action) && r.OperationStatus == handler.InProgress {
			return re.report(event, "Response error", errors.New("READ and LIST handlers must return synchronous"), invalidRequestError)
		}
		return r, nil
	}
}

func scrubFiles(dir string) error {
	names, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entery := range names {
		os.RemoveAll(path.Join([]string{dir, entery.Name()}...))
	}
	return nil
}

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
			event.Request.NextToken,
			event.Request.StackId,  //TODO: understand how to fix this to use a proper value
			event.Request.StackTags, //TODO: understand how to fix this to use a proper value
			event.Request.Region,
			event.Request.AWSAccountID,
			event.Request.SystemTags,
			event.CallbackContext,
			credentials.SessionFromCredentialsProvider(&event.Credentials),
			event.Request.PreviousResourceState,
			event.Request.DesiredResourceState,
		)
		progEvt := handlerFn(request)
		return progEvt, nil
	}
}

// Invoke handles the invocation of the handerFn.
func invoke(handlerFn handlerFunc, request handler.Request, metricsPublisher *metrics.Publisher, action string) handler.ProgressEvent {

	// Create a channel to received a signal that work is done.
	ch := make(chan handler.ProgressEvent, 1)

	// Ask the goroutine to do some work for us.
	go func() {
		//start the timer
		s := time.Now()
		metricsPublisher.PublishInvocationMetric(time.Now(), string(action))

		// Report the work is done.
		pe := handlerFn(request)
		log.Printf("Received event: %s\nMessage: %s\n",
			pe.OperationStatus,
			pe.Message,
		)
		e := time.Since(s)
		metricsPublisher.PublishDurationMetric(time.Now(), string(action), e.Seconds()*1e3)
		ch <- pe
	}()
	return <-ch
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
