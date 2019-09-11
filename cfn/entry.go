package cfn

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/action"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/handler"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

const (
	InvalidRequestError  string        = "InvalidRequest"
	ServiceInternalError string        = "ServiceInternal"
	ValidationError      string        = "Validation"
	Timeout              time.Duration = 60 * time.Second
)

// BuilderFn is a convenience type for the builder callback.
// It enables the creation of resource types without being tied
// to a specific resource struct.
type BuilderFn func() interface{}

// Builder enables the creation of resource structs whenever
// they are required.
type Builder interface {
	// BuilderCallback stores the creation function
	//
	// Example:
	// 	type (r *Resource) BuilderCallback(func() interface{} {
	//		return new(Resource)
	//	})
	BuilderCallback(BuilderFn)

	// Build will execute the stored callback function
	Build() interface{}
}

// Handlers represents the actions from the AWS CloudFormation service
//
// Each action maps directly to a CloudFormation action. Every action is
// expected to return a response and/or an error.
//
// A valid error condition would be met if the resource operation failed or
// an API is no longer available.
type Handlers interface {
	// Implement the `Builder` interface to allow the RPDK to create structs that match the resource.
	//
	// This interface is called during the hydration of the event lifecycle.
	// Builder

	// Create action
	Create(request Request) (Response, error)

	// Read action
	Read(request Request) (Response, error)

	// Update action
	Update(request Request) (Response, error)

	// Delete action
	Delete(request Request) (Response, error)

	// List action
	List(request Request) (Response, error)
}

// Event base structure, it will be internal to the RPDK.
//
// @todo Consider moving to an internal pkg
type Event struct {
	Action              action.Action
	AWSAccountID        string `validate:"min=12"`
	BearerToken         string `validate:"nonzero"`
	Context             *RequestContext
	NextToken           string
	Region              string `validate:"nonzero"`
	RequestData         *RequestData
	ResourceType        string `validate:"nonzero"`
	ResourceTypeVersion float32
	ResponseEndpoint    string `validate:"nonzero"`
	StackID             string `validate:"nonzero"`
} // may need to manually unmarshal?

// RequestData is internal to the RPDK. It contains a number of fields that are for
// internal use only.
//
// @todo Consider moving to an internal pkg
type RequestData struct {
	CallerCredentials          *credentials.Credentials
	LogicalResourceID          string
	PlatformCredentials        *credentials.Credentials
	PreviousResourceProperties json.RawMessage
	PreviousStackTags          Tags
	ProviderCredentials        *credentials.Credentials
	ProviderLogGroupName       string
	ResourceProperties         json.RawMessage
	StackTags                  Tags
	SystemTags                 Tags
}

// RequestContext handles elements such as reties and long running creations.
//
// @todo Consider moving to an internal pkg
type RequestContext struct {
	CallbackContext          map[string]string
	CloudWatchEventsRuleName string
	CloudWatchEventsTargetID string
	Invocation               int32
}

// Tags are store as key/value pairs.
type Tags map[string]string

// EventFunc ...
type EventFunc func(ctx context.Context, event Event) (Response, error)

// HandlerFunc ...
type HandlerFunc func(request Request) (Response, error)

// Request will be passed to actions with customer related data, such as resource states
type Request interface {
	Action() action.Action
	PreviousResourceProperties(v interface{}) error
	ResourceProperties(v interface{}) error
	LogicalResourceID() string
	BearerToken() string
	ResponseEndpoint() string
}

// Response ...
type Response interface {
	json.Marshaler
}

// Router decides which handler should be invoked based on the action
//
// The Router will return a route or an error depending on the action passed in
func Router(a action.Action, h Handlers) (HandlerFunc, error) {
	// Figure out which action was called and have a "catch-all"
	switch a {
	case action.Create:
		return h.Create, nil
	case action.Read:
		return h.Read, nil
	case action.Update:
		return h.Update, nil
	case action.Delete:
		return h.Update, nil
	case action.List:
		return h.List, nil
	default:
		// No action matched, we should fail and return an invalidRequestErrorCode
		return nil, cfnerr.New(InvalidRequestError, "No action/invalid action specified", nil)
	}
}

func ValidateEvent(event *Event) error {
	errs := []error{}

	if len(event.BearerToken) == 0 {
		errs = append(errs, errors.New("Bearer Token is empty"))
	}

	if len(event.ResponseEndpoint) == 0 {
		errs = append(errs, errors.New("Response Endpoint is empty"))
	}

	if len(event.ResourceType) == 0 {
		errs = append(errs, errors.New("Resource Type not specified"))
	}

	if len(errs) > 0 {
		return cfnerr.NewBatchError(ValidationError, "Failed Vailidation", errs)
	}

	return nil
}

// Handler is the entry point to all invocations of a custom resource
func Handler(h Handlers) EventFunc {
	return func(ctx context.Context, event Event) (Response, error) {
		handlerFn, err := Router(event.Action, h)
		if err != nil {
			cfnErr := cfnerr.New(ServiceInternalError, "Unable to complete request", err)
			return handler.NewFailedResponse(cfnErr), cfnErr
		}

		// @todo validate input - based on spec?

		request := handler.NewRequest(
			event.Action,
			event.RequestData.PreviousResourceProperties,
			event.RequestData.ResourceProperties,
			event.RequestData.LogicalResourceID,
			event.BearerToken,
			event.ResponseEndpoint,
		)

		resp, err := Invoke(handlerFn, request)
		if err != nil {
			cfnErr := cfnerr.New(ServiceInternalError, "Unable to complete request", err)
			return handler.NewFailedResponse(cfnErr), err
		}

		return resp, nil
	}
}

//Invoke handles the invocation of the handerFn.
func Invoke(handlerFn HandlerFunc, request handler.NewRequest) (Response, error) {
	for {
		// Create a context that is both manually cancellable and will signal
		// a cancel at the specified duration.
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		//We always defer a cancel.
		defer cancel()

		// Create a channel to received a signal that work is done.
		ch := make(chan Response, 1)

		// Create a channel to received error.
		cherror := make(chan error, 1)

		// Ask the goroutine to do some work for us.
		go func() {
			// Report the work is done.
			resp, err := handlerFn(request)

			if err != nil {
				cherror <- err
			}

			ch <- resp
		}()

		// Wait for the work to finish. If it takes too long move on. If the function returns an error, signal the error channel.
		select {
		case e := <-cherror:
			//The handler returned an error.
			return nil, e

		case d := <-ch:
			//Return the response from the handler.
			return d, nil

		case <-ctx.Done():
			//handler failed to respond.
			return nil, errors.New("Handler failed to respond")
		}

	}

}

// Start ...
func Start(h EventFunc) {
	lambda.Start(h)
}
