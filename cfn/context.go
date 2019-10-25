package cfn

import (
	"context"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/aws/session"
)

// contextKey is used to prevent collisions within the context package
// and to guarantee returning the correct values from a context
type contextKey string

// callbackContextValues is used to guarantee the type of
// values stored in the context
type callbackContextValues map[string]interface{}

// setContextValues creates a context to pass to handlers
func setContextValues(ctx context.Context, values callbackContextValues) context.Context {
	return context.WithValue(ctx, contextKey("user_callback_context"), values)
}

// getContextValues unwraps callbackContextValues from a given context
func getContextValues(ctx context.Context) (callbackContextValues, error) {
	values, ok := ctx.Value(contextKey("user_callback_context")).(callbackContextValues)
	if !ok {
		cfnErr := cfnerr.New(sessionNotFoundError, "Session not found", nil)
		return nil, cfnErr
	}

	return values, nil
}

// setContextSession adds the supplied session to the given context
func setContextSession(ctx context.Context, sess *session.Session) context.Context {
	return context.WithValue(ctx, contextKey("aws_session"), sess)
}

// getContextSession unwraps a session from a given context
func contextSession(ctx context.Context) (*session.Session, error) {
	val, ok := ctx.Value(contextKey("aws_session")).(*session.Session)
	if !ok {
		cfnErr := cfnerr.New(sessionNotFoundError, "Session not found", nil)
		return nil, cfnErr
	}

	return val, nil
}

// marshalCallback allows for a handler.ProgressEvent to be parsed into something
// the RPDK can use to reinvoke the resource provider with the same context.
func marshalCallback(pevt *handler.ProgressEvent) (callbackContextValues, int64) {
	return pevt.CallbackContext, pevt.CallbackDelaySeconds
}
