package handler

import (
	"context"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	// MarshalingError occurs when we can't marshal data from one format into another.
	MarshalingError string = "Marshaling"
	// BodyEmptyError happens when the resource body is empty
	BodyEmptyError string = "BodyEmpty"
	// SessionNotFoundError occurs when the AWS SDK session isn't available in the context
	SessionNotFoundError string = "SessionNotFound"
)

// ContextKey is used to prevent collisions within the context package
// It's used is for the CallbackContext key in the Request Context
//
// 	ctx.Value(handler.ContextKey("foo")).(Thing)
type ContextKey string

// CallbackContextValues ...
type CallbackContextValues map[string]interface{}

// ContextValues creates a context to pass to handlers
func ContextValues(ctx context.Context, values CallbackContextValues) context.Context {
	return context.WithValue(ctx, ContextKey("user_callback_context"), values)
}

// ContextCallback ...
func ContextCallback(ctx context.Context) (CallbackContextValues, error) {
	values, ok := ctx.Value(ContextKey("user_callback_context")).(CallbackContextValues)
	if !ok {
		cfnErr := cfnerr.New(SessionNotFoundError, "Session not found", nil)
		return nil, cfnErr
	}

	return values, nil
}

// ContextInjectSession ..
func ContextInjectSession(ctx context.Context, sess *session.Session) context.Context {
	ctx = context.WithValue(ctx, ContextKey("aws_session"), sess)

	return ctx
}

// ContextSession ..
func ContextSession(ctx context.Context) (*session.Session, error) {
	val, ok := ctx.Value(ContextKey("aws_session")).(*session.Session)
	if !ok {
		cfnErr := cfnerr.New(SessionNotFoundError, "Session not found", nil)
		return nil, cfnErr
	}

	return val, nil
}
