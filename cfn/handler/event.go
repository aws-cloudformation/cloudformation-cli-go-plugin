package handler

import (
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/cfnerr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// ProgressEvent represent the progress of CRUD handlers.
type ProgressEvent struct {
	// OperationStatus indicates whether the handler has reached a terminal state or is
	// still computing and requires more time to complete.
	OperationStatus Status `json:"status,omitempty"`

	// HandlerErrorCode should be provided when OperationStatus is FAILED or IN_PROGRESS.
	HandlerErrorCode string `json:"errorCode,omitempty"`

	// Message which can be shown to callers to indicate the
	//nature of a progress transition or callback delay; for example a message
	//indicating "propagating to edge."
	Message string `json:"message,omitempty"`

	// CallbackContext is an arbitrary datum which the handler can return in an
	// IN_PROGRESS event to allow the passing through of additional state or
	// metadata between subsequent retries; for example to pass through a Resource
	// identifier which can be used to continue polling for stabilization
	CallbackContext map[string]interface{} `json:"callbackContext,omitempty"`

	// CallbackDelaySeconds will be scheduled with an initial delay of no less than the number
	// of seconds specified in the progress event. Set this value to <= 0 to
	// indicate no callback should be made.
	CallbackDelaySeconds int64 `json:"callbackDelaySeconds,omitempty"`

	// ResourceModel is the output resource instance populated by a READ/LIST for synchronous results
	// and by CREATE/UPDATE/DELETE for final response validation/confirmation
	ResourceModel interface{} `json:"resourceModel,omitempty"`

	// ResourceModels is the output resource instances populated by a LIST for synchronous results
	ResourceModels []interface{} `json:"resourceModels,omitempty"`

	// NextToken is the token used to request additional pages of resources for a LIST operation
	NextToken string `json:"nextToken,omitempty"`
}

// NewProgressEvent creates a new event with
// a default OperationStatus of Unkown
func NewProgressEvent() ProgressEvent {
	return ProgressEvent{
		OperationStatus: UnknownStatus,
	}
}

// NewFailedEvent creates a generic failure progress event
// based on the error passed in.
func NewFailedEvent(err error) ProgressEvent {
	cerr := cfnerr.New(
		cloudformation.HandlerErrorCodeGeneralServiceException,
		"Unable to complete request: "+err.Error(),
		err,
	)

	return ProgressEvent{
		OperationStatus:  Failed,
		Message:          cerr.Message(),
		HandlerErrorCode: cloudformation.HandlerErrorCodeGeneralServiceException,
	}
}
