package handler

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
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
	LogicalResourceID              string
	CallbackContext                map[string]interface{}
	Session                        *session.Session
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

	if err := json.Unmarshal(r.previousResourcePropertiesBody, v); err != nil {
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

	if err := json.Unmarshal(r.resourcePropertiesBody, v); err != nil {
		return cfnerr.New(marshalingError, "Unable to convert type", err)
	}

	return nil
}
