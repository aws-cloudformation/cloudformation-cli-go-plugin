package handler

import (
	"encoding/json"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
)

// Request is passed to actions with customer related data
// such as resource states
type Request struct {
	previousResourcePropertiesBody json.RawMessage
	resourcePropertiesBody         json.RawMessage
	logicalResourceID              string
	bearerToken                    string
}

// NewRequest returns a new Request based on the provided parameters
func NewRequest(previousBody json.RawMessage, body json.RawMessage, logicalResourceID string, bearerToken string) Request {
	return Request{
		previousResourcePropertiesBody: previousBody,
		resourcePropertiesBody:         body,
		logicalResourceID:              logicalResourceID,
		bearerToken:                    bearerToken,
	}
}

// UnmarshalPrevious populates the provided interface
// with the previous properties of the resource
func (r *Request) UnmarshalPrevious(v interface{}) error {
	if len(r.resourcePropertiesBody) == 0 {
		return cfnerr.New(BodyEmptyError, "Body is empty", nil)
	}

	if err := json.Unmarshal(r.previousResourcePropertiesBody, v); err != nil {
		return cfnerr.New(MarshalingError, "Unable to convert type", err)
	}

	return nil
}

// Unmarshal populates the provided interface
// with the current properties of the resource
func (r *Request) Unmarshal(v interface{}) error {
	if len(r.resourcePropertiesBody) == 0 {
		return cfnerr.New(BodyEmptyError, "Body is empty", nil)
	}

	if err := json.Unmarshal(r.resourcePropertiesBody, v); err != nil {
		return cfnerr.New(MarshalingError, "Unable to convert type", err)
	}

	return nil
}

// LogicalResourceID returns the logical ID of the related resource
func (r *Request) LogicalResourceID() string {
	return r.logicalResourceID
}

// BearerToken returns the bearer token related to the request
func (r *Request) BearerToken() string {
	return r.bearerToken
}
