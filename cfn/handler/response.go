package handler

import (
	"encoding/json"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
)

// Response represents a response to the
// cloudformation service from a resource handler.
// The zero value is ready to use.
type Response struct {
	message         string      `json:"message,omitempty"`
	operationStatus Status      `json:"operationStatus,omitempty"`
	resourceModel   interface{} `json:"resourceModel,omitempty"`
	errorCode       error       `json:"errorCode,omitempty"`
	bearerToken     string      `json:"bearerToken,omitempty"`
}

// NewFailedResponse returns a Response pre-filled with the supplied error
func NewFailedResponse(err error) Response {
	return Response{
		operationStatus: Failed,
		errorCode:       err,
		message:         err.Error(),
	}
}

// MarshalJSON returns the response object as a JSON string
func (r *Response) MarshalJSON() ([]byte, error) {
	var resp struct {
		Message         string      `json:"message,omitempty"`
		OperationStatus string      `json:"operationStatus,omitempty"`
		ResourceModel   interface{} `json:"resourceModel,omitempty"`
		ErrorCode       string      `json:"errorCode,omitempty"`
		BearerToken     string      `json:"bearerToken,omitempty"`
	}

	cfnErr, ok := r.Error().(cfnerr.Error)
	if cfnErr != nil && !ok {
		return nil, cfnerr.New(MarshalingError, "Unable to marshal response, zomg", r.Error())
	}

	resp.Message = r.Message()
	resp.OperationStatus = string(r.operationStatus)
	resp.ResourceModel = r.ResourceModel()

	if cfnErr != nil {
		resp.ErrorCode = cfnErr.Code()
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return nil, cfnerr.New(MarshalingError, "Unable to marshal response", err)
	}

	return b, nil
}

// Message returns the response's message
func (r *Response) Message() string {
	return r.message
}

// OperationStatus returns the response's operation status
func (r *Response) OperationStatus() Status {
	return r.operationStatus
}

// Error returns the response's error code
func (r *Response) Error() error {
	return r.errorCode
}

// ResourceModel returns the response's resource model
func (r *Response) ResourceModel() interface{} {
	return r.resourceModel
}
