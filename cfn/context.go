package cfn

import (
	"context"
	"fmt"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go-v2/aws"
)

type contextKey string
type callbackContextValues map[string]interface{}

const (
	valuesKey  = contextKey("user_callback_context")
	configKey  = contextKey("aws_config")
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

// SetContextConfig adds the supplied AWS config to the given context
func SetContextConfig(ctx context.Context, cfg *aws.Config) context.Context {
	return context.WithValue(ctx, configKey, cfg)
}

// GetContextConfig unwraps an aws.Config from a given context
func GetContextConfig(ctx context.Context) (*aws.Config, error) {
	val, ok := ctx.Value(configKey).(*aws.Config)
	if !ok {
		return nil, fmt.Errorf("Config not found")
	}
	return val, nil
}

// marshalCallback allows for a handler.ProgressEvent to be parsed into something
// the RPDK can use to reinvoke the resource provider with the same context.
//
//nolint:all
func marshalCallback(pevt *handler.ProgressEvent) (map[string]interface{}, int64) {
	return map[string]interface{}(pevt.CallbackContext), pevt.CallbackDelaySeconds
}
