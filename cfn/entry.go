package cfn

import (
	"context"
	"encoding/json"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/action"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/cfnerr"

	"github.com/aws/aws-lambda-go/lambda"
)

const (
	invalidRequestErrorCode string = "InvalidRequest"
)

// Builder ...
type Builder interface {
	Build() interface{}
}

// BuilderCallbackFn ...
type BuilderCallbackFn func() interface{}

// Handlers ...
type Handlers interface {
	Builder

	Create(ctx context.Context, request Request)
	Read(ctx context.Context, request Request)
	Update(ctx context.Context, request Request)
	Delete(ctx context.Context, request Request)
	List(ctx context.Context, request Request)
}

type Body struct {
	AWSAccountID        string
	ResponseEndpoint    string
	BearerToken         string
	NextToken           string
	Region              string
	ResourceType        string
	ResourceTypeVersion string
	Context             context.Context
	Data                string
	StackID             string
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
		return nil, cfnerr.New(invalidRequestErrorCode, "No action/invalid action specified", nil)
	}
}

// HandlerFunc type
type HandlerFunc func(ctx context.Context, request Request)

// Request ...
type Request interface {
	json.Unmarshaler

	Action() action.Action
}

// Response ...
type Response interface {
	json.Marshaler
}

// Handler is the entry point to all invocations of a custom resource
func Handler(h Handlers) HandlerFunc {
	return func(ctx context.Context, request Request) {
		handlerFn, err := Router(request.Action(), h)
		if err != nil {
			// return a failure output
		}

		handlerFn(ctx, request)
	}
}

// Start ...
func Start(h HandlerFunc) {
	lambda.Start(h)
}
