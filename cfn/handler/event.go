package handler

import (
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
)

// ProgressEvent represent the progress of CRUD handlers.
type ProgressEvent struct {
	// The status indicates whether the handler has reached a terminal state or is
	// still computing and requires more time to complete.
	OperationStatus Status

	// If OperationStatus is FAILED or IN_PROGRESS, an error code should be provided.
	HandlerErrorCode string

	// The handler can (and should) specify a contextual information message which
	// can be shown to callers to indicate the nature of a progress transition or
	// callback delay; for example a message indicating "propagating to edge."
	Message string

	// The callback context is an arbitrary datum which the handler can return in an
	// IN_PROGRESS event to allow the passing through of additional state or
	// metadata between subsequent retries; for example to pass through a Resource
	// identifier which can be used to continue polling for stabilization
	CallbackContext map[string]interface{}

	// A callback will be scheduled with an initial delay of no less than the number
	// of seconds specified in the progress event. Set this value to <= 0 to
	// indicate no callback should be made.
	CallbackDelaySeconds int64

	// The output resource instance populated by a READ/LIST for synchronous results
	// and by CREATE/UPDATE/DELETE for final response validation/confirmation
	ResourceModel interface{}

	// The BearerToken is used to report progress back to CloudFormation and is
	// passed back to CloudFormation
	BearerToken string
}

// NewEvent creates a new event
// with a default OperationStatus of Unkown
func NewProgressEvent(req Request) ProgressEvent {
	return ProgressEvent{
		OperationStatus: UnknownStatus,
		BearerToken:     req.bearerToken,
	}
}

// NewFailedEvent creates a generic failure progress event based on
// the error passed in.
func NewFailedEvent(req Request, err cfnerr.Error) ProgressEvent {
	return ProgressEvent{
		BearerToken:      req.bearerToken,
		OperationStatus:  Failed,
		Message:          err.Message(),
		HandlerErrorCode: err.Code(),
	}
}
