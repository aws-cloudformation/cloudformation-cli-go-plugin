package cfn

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
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

var once sync.Once

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

	log.Printf("Handler starting")
	lambda.Start(makeEventFunc(h))

	log.Printf("Handler finished")
}

// Tags are stored as key/value paired strings
type tags map[string]string

// eventFunc is the function signature required to execute an event from the Lambda SDK
type eventFunc func(ctx context.Context, event *event) (response, error)

// handlerFunc is the signature required for all actions
type handlerFunc func(request handler.Request) handler.ProgressEvent

// MakeEventFunc is the entry point to all invocations of a custom resource
func makeEventFunc(h Handler) eventFunc {
	return func(ctx context.Context, event *event) (response, error) {
		ps := credentials.SessionFromCredentialsProvider(&event.RequestData.ProviderCredentials)
		m := metrics.New(cloudwatch.New(ps), event.ResourceType)
		once.Do(func() {
			l, err := logging.NewCloudWatchLogsProvider(
				cloudwatchlogs.New(ps),
				m,
				event.RequestData.ProviderLogGroupName,
			)
			if err != nil {
				log.Printf("Error: %v, Logging to Stdout", err)
				m.PublishExceptionMetric(time.Now(), event.Action, err)
				l = os.Stdout
			}
			// Set default logger to output to CWL in the provider account
			logging.SetProviderLogOutput(l)
		})
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
		rctx := handler.RequestContext{
			StackID:    event.StackID,
			Region:     event.Region,
			AccountID:  event.AWSAccountID,
			StackTags:  event.RequestData.StackTags,
			SystemTags: event.RequestData.SystemTags,
			NextToken:  event.NextToken,
		}
		request := handler.NewRequest(
			event.RequestData.LogicalResourceID,
			event.CallbackContext,
			rctx,
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
