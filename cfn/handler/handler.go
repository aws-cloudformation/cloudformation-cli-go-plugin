package handler

import (
	"encoding/json"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/action"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/operationstatus"
)

const (
	MarshalingError string = "Marshaling"
	BodyEmptyError  string = "BodyEmpty"
)

// NewRequest ...
func NewRequest(a action.Action, previousBody json.RawMessage, body json.RawMessage, logicalResourceID string, bearerToken string, endpoint string) *Request {
	req := &Request{
		action:                         a,
		previousResourcePropertiesBody: previousBody,
		resourcePropertiesBody:         body,
		logicalResourceID:              logicalResourceID,
		bearerToken:                    bearerToken,
		responseEndpoint:               endpoint,
	}

	return req
}

// Request ...
type Request struct {
	action                         action.Action
	previousResourcePropertiesBody json.RawMessage
	resourcePropertiesBody         json.RawMessage
	logicalResourceID              string
	bearerToken                    string
	responseEndpoint               string
}

// Action returns the action to be performed
func (r *Request) Action() action.Action {
	return r.action
}

// PreviousResourceProperties ...
func (r *Request) PreviousResourceProperties(v interface{}) error {
	if len(r.resourcePropertiesBody) == 0 {
		return cfnerr.New(BodyEmptyError, "Body is empty", nil)
	}

	if err := json.Unmarshal(r.previousResourcePropertiesBody, v); err != nil {
		return cfnerr.New(MarshalingError, "Unable to convert type", err)
	}

	return nil
}

// ResourceProperties ...
func (r *Request) ResourceProperties(v interface{}) error {
	if len(r.resourcePropertiesBody) == 0 {
		return cfnerr.New(BodyEmptyError, "Body is empty", nil)
	}

	if err := json.Unmarshal(r.resourcePropertiesBody, v); err != nil {
		return cfnerr.New(MarshalingError, "Unable to convert type", err)
	}

	return nil
}

// LogicalResourceID ...
func (r *Request) LogicalResourceID() string {
	return r.logicalResourceID
}

// BearerToken ...
func (r *Request) BearerToken() string {
	return r.bearerToken
}

// ResponseEndpoint ...
func (r *Request) ResponseEndpoint() string {
	return r.responseEndpoint
}

// NewResponse ...
func NewResponse() *Response {
	return &Response{}
}

// NewFailedResponse ...
func NewFailedResponse(err error) *Response {
	return &Response{
		OperationStatus: operationstatus.Failed,
		ErrorCode:       err,
		Message:         err.Error(),
	}
}

// Response ...
type Response struct {
	Message         string
	OperationStatus operationstatus.Status
	ResourceModel   string
	BearerToken     string
	ErrorCode       error
}

// MarshalJSON ...
func (r *Response) MarshalJSON() ([]byte, error) {
	return nil, nil
}
