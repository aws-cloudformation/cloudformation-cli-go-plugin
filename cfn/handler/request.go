package handler

import (
	"encoding/json"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
)

const (
	// marshalingError occurs when we can't marshal data from one format into another.
	marshalingError = "Marshaling"

	// bodyEmptyError happens when the resource body is empty
	bodyEmptyError = "BodyEmpty"

	// sessionNotFoundError occurs when the AWS SDK session isn't available in the context
	sessionNotFoundError = "SessionNotFound"
)

// Request is passed to actions with customer related data
// such as resource states
type Request struct {
	LogicalResourceID              string
	bearerToken                    string
	previousResourcePropertiesBody json.RawMessage
	resourcePropertiesBody         json.RawMessage
}

// NewRequest returns a new Request based on the provided parameters
func NewRequest(previousBody json.RawMessage, body json.RawMessage, logicalResourceID string, bearerToken string) Request {
	return Request{
		LogicalResourceID:              logicalResourceID,
		bearerToken:                    bearerToken,
		previousResourcePropertiesBody: previousBody,
		resourcePropertiesBody:         body,
	}
}

// UnmarshalPrevious populates the provided interface
// with the previous properties of the resource
func (r *Request) UnmarshalPrevious(v interface{}) error {
	if len(r.resourcePropertiesBody) == 0 {
		return cfnerr.New(bodyEmptyError, "Body is empty", nil)
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
