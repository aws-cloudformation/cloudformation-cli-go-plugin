package cfn

import (
	"context"
	"fmt"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/aws/session"
)

// contextKey is used to prevent collisions within the context package
// and to guarantee returning the correct values from a context
type contextKey string

// callbackContextValues is used to guarantee the type of
// values stored in the context
type callbackContextValues map[string]interface{}

const (
	valuesKey  = contextKey("user_callback_context")
	sessionKey = contextKey("aws_session")
)

// SetContextValues creates a context to pass to handlers
func SetContextValues(ctx context.Context, values map[string]interface{}) context.Context {
	return context.WithValue(ctx, valuesKey, callbackContextValues(values))
}

// GetContextValues unwraps callbackContextValues from a given context
func GetContextValues(ctx context.Context) (map[string]interface{}, error) {
	values, ok := ctx.Value(valuesKey).(callbackContextValues)
	if !ok {
		return nil, fmt.Errorf("Values not found")
	}

	return map[string]interface{}(values), nil
}

// SetContextSession adds the supplied session to the given context
func SetContextSession(ctx context.Context, sess *session.Session) context.Context {
	return context.WithValue(ctx, sessionKey, sess)
}

// GetContextSession unwraps a session from a given context
func GetContextSession(ctx context.Context) (*session.Session, error) {
	val, ok := ctx.Value(sessionKey).(*session.Session)
	if !ok {
		return nil, fmt.Errorf("Session not found")
	}

	return val, nil
}

// marshalCallback allows for a handler.ProgressEvent to be parsed into something
// the RPDK can use to reinvoke the resource provider with the same context.
func marshalCallback(pevt *handler.ProgressEvent) (map[string]interface{}, int64) {
	return map[string]interface{}(pevt.CallbackContext), pevt.CallbackDelaySeconds
}
