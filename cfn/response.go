package cfn

import (
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
)

// response represents a response to the
// cloudformation service from a resource handler.
// The zero value is ready to use.
type response struct {
	Message         string         `json:"message,omitempty"`
	OperationStatus handler.Status `json:"operationStatus,omitempty"`
	ResourceModel   interface{}    `json:"resourceModel,omitempty"`
	ErrorCode       cfnerr.Error   `json:"errorCode,omitempty"`
	BearerToken     string         `json:"bearerToken,omitempty"`
}

// newFailedResponse returns a response pre-filled with the supplied error
func newFailedResponse(err error, bearerToken string) response {
	return response{
		OperationStatus: handler.Failed,
		ErrorCode:       cfnerr.New(cfnerr.InternalFailure, "Unpexected error", err),
		Message:         err.Error(),
		BearerToken:     bearerToken,
	}
}

// newResponse converts a progress event into a useable reponse
// for the CloudFormation Resource Provider service to understand.
func newResponse(pevt *handler.ProgressEvent, bearerToken string) (response, error) {
	resp := response{
		BearerToken:     bearerToken,
		Message:         pevt.Message,
		OperationStatus: pevt.OperationStatus,
		ResourceModel:   pevt.ResourceModel,
	}

	if pevt.HandlerErrorCode != "" {
		resp.ErrorCode = cfnerr.New(pevt.HandlerErrorCode, pevt.Message, nil)
	}

	return resp, nil
}
