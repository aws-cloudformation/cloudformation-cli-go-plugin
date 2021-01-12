package cfn

import (
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// response represents a response to the
// cloudformation service from a resource handler.
// The zero value is ready to use.
type response struct {
	// Message which can be shown to callers to indicate the nature of a
	//progress transition or callback delay; for example a message
	//indicating "propagating to edge"
	Message string `json:"message,omitempty"`

	//The operationStatus indicates whether the handler has reached a terminal
	//state or is still computing and requires more time to complete
	OperationStatus handler.Status `json:"status,omitempty"`

	//ResourceModel it The output resource instance populated by a READ/LIST for
	//synchronous results and by CREATE/UPDATE/DELETE for final response
	//validation/confirmation
	ResourceModel interface{} `json:"resourceModel,omitempty"`

	// ErrorCode is used to report granular failures back to CloudFormation
	ErrorCode string `json:"errorCode,omitempty"`

	// BearerToken is used to report progress back to CloudFormation and is
	//passed back to CloudFormation
	BearerToken string `json:"bearerToken,omitempty"`

	// ResourceModels is the output resource instances populated by a LIST for
	// synchronous results. ResourceModels must be returned by LIST so it's
	// always included in the response. When ResourceModels is not set, null is
	// returned.
	ResourceModels []interface{} `json:"resourceModels"`

	// NextToken the token used to request additional pages of resources for a LIST operation
	NextToken string `json:"nextToken,omitempty"`

	// CallbackContext is an arbitrary datum which the handler can return in an
	// IN_PROGRESS event to allow the passing through of additional state or
	// metadata between subsequent retries; for example to pass through a Resource
	// identifier which can be used to continue polling for stabilization
	CallbackContext map[string]interface{} `json:"callbackContext,omitempty"`

	// CallbackDelaySeconds will be scheduled with an initial delay of no less than the number
	// of seconds specified in the progress event. Set this value to <= 0 to
	// indicate no callback should be made.
	CallbackDelaySeconds int64 `json:"callbackDelaySeconds,omitempty"`
}

// newFailedResponse returns a response pre-filled with the supplied error
func newFailedResponse(err error, bearerToken string) response {
	return response{
		OperationStatus: handler.Failed,
		ErrorCode:       cloudformation.HandlerErrorCodeInternalFailure,
		Message:         err.Error(),
		BearerToken:     bearerToken,
	}
}

// newResponse converts a progress event into a useable reponse
// for the CloudFormation Resource Provider service to understand.
func newResponse(pevt *handler.ProgressEvent, bearerToken string) (response, error) {

	resp := response{
		BearerToken:          bearerToken,
		Message:              pevt.Message,
		OperationStatus:      pevt.OperationStatus,
		ResourceModel:        pevt.ResourceModel,
		ResourceModels:       pevt.ResourceModels,
		NextToken:            pevt.NextToken,
		CallbackContext:      pevt.CallbackContext,
		CallbackDelaySeconds: pevt.CallbackDelaySeconds,
	}

	if pevt.HandlerErrorCode != "" {
		resp.ErrorCode = pevt.HandlerErrorCode
	}

	return resp, nil
}
