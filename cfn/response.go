package cfn

import (
	"encoding/json"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
)

// response represents a response to the
// cloudformation service from a resource handler.
// The zero value is ready to use.
type response struct {
	message         string         `json:"message,omitempty"`
	operationStatus handler.Status `json:"operationStatus,omitempty"`
	resourceModel   interface{}    `json:"resourceModel,omitempty"`
	errorCode       error          `json:"errorCode,omitempty"`
	bearerToken     string         `json:"bearerToken,omitempty"`
}

// newFailedResponse returns a response pre-filled with the supplied error
func newFailedResponse(err error) response {
	return response{
		operationStatus: handler.Failed,
		errorCode:       err,
		message:         err.Error(),
	}
}

// marshalResponse converts a progress event into a useable reponse
// for the CloudFormation Resource Provider service to understand.
func marshalResponse(pevt *handler.ProgressEvent, bearerToken string) (response, error) {
	resp := response{
		operationStatus: pevt.OperationStatus,
		message:         pevt.Message,
		bearerToken:     bearerToken,
	}

	if len(pevt.HandlerErrorCode) == 0 {
		resp.errorCode = cfnerr.New(pevt.HandlerErrorCode, pevt.Message, nil)
	}

	if pevt.ResourceModel != nil {
		resp.resourceModel = pevt.ResourceModel
	}

	return resp, nil
}

// MarshalJSON returns the response object as a JSON string
func (r *response) MarshalJSON() ([]byte, error) {
	var resp struct {
		Message         string      `json:"message,omitempty"`
		OperationStatus string      `json:"operationStatus,omitempty"`
		ResourceModel   interface{} `json:"resourceModel,omitempty"`
		ErrorCode       string      `json:"errorCode,omitempty"`
		BearerToken     string      `json:"bearerToken,omitempty"`
	}

	cfnErr, ok := r.Error().(cfnerr.Error)
	if cfnErr != nil && !ok {
		return nil, cfnerr.New(marshalingError, "Unable to marshal response, zomg", r.Error())
	}

	resp.Message = r.Message()
	resp.OperationStatus = string(r.operationStatus)
	resp.ResourceModel = r.ResourceModel()

	if cfnErr != nil {
		resp.ErrorCode = cfnErr.Code()
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return nil, cfnerr.New(marshalingError, "Unable to marshal response", err)
	}

	return b, nil
}

// Message returns the response's message
func (r *response) Message() string {
	return r.message
}

// OperationStatus returns the response's operation status
func (r *response) OperationStatus() handler.Status {
	return r.operationStatus
}

// Error returns the response's error code
func (r *response) Error() error {
	return r.errorCode
}

// ResourceModel returns the response's resource model
func (r *response) ResourceModel() interface{} {
	return r.resourceModel
}
