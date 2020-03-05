package handler

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/encoding"
)

const (
	// marshalingError occurs when we can't marshal data from one format into another.
	marshalingError = "Marshaling"

	// bodyEmptyError happens when the resource body is empty
	bodyEmptyError = "BodyEmpty"
)

// Request is passed to actions with customer related data
// such as resource states
type Request struct {
	// The logical ID of the resource in the CloudFormation stack
	LogicalResourceID string

	// The callback context is an arbitrary datum which the handler can return in an
	// IN_PROGRESS event to allow the passing through of additional state or
	// metadata between subsequent retries; for example to pass through a Resource
	// identifier which can be used to continue polling for stabilization
	CallbackContext map[string]interface{}

	// An authenticated AWS session that can be used with the AWS Go SDK
	Session *session.Session

	previousResourcePropertiesBody []byte
	resourcePropertiesBody         []byte
}

// NewRequest returns a new Request based on the provided parameters
func NewRequest(id string, ctx map[string]interface{}, sess *session.Session, previousBody, body []byte) Request {
	return Request{
		LogicalResourceID:              id,
		CallbackContext:                ctx,
		Session:                        sess,
		previousResourcePropertiesBody: previousBody,
		resourcePropertiesBody:         body,
	}
}

// UnmarshalPrevious populates the provided interface
// with the previous properties of the resource
func (r *Request) UnmarshalPrevious(v interface{}) error {
	if len(r.previousResourcePropertiesBody) == 0 {
		return nil
	}

	if err := encoding.Unmarshal(r.previousResourcePropertiesBody, v); err != nil {
		return cfnerr.New(marshalingError, "Unable to convert type", err)
	}

	return nil
}

// Unmarshal populates the provided interface
// with the current properties of the resource
func (r *Request) Unmarshal(v interface{}) error {
	if len(r.resourcePropertiesBody) == 0 {
		return cfnerr.New(bodyEmptyError, "Body is empty", nil)
	}

	if err := encoding.Unmarshal(r.resourcePropertiesBody, v); err != nil {
		return cfnerr.New(marshalingError, "Unable to convert type", err)
	}

	return nil
}
