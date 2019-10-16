package handler

import (
	"context"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/operationstatus"
)

// ProgressEvent represent the progress of CRUD handlers.
type ProgressEvent struct {
	// The status indicates whether the handler has reached a terminal state or is
	// still computing and requires more time to complete.
	OperationStatus operationstatus.Status

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
	CallbackContext context.Context

	// A callback will be scheduled with an initial delay of no less than the number
	// of seconds specified in the progress event. Set this value to <= 0 to
	// indicate no callback should be made.
	CallbackDelaySeconds int64

	// The output resource instance populated by a READ/LIST for synchronous results
	// and by CREATE/UPDATE/DELETE for final response validation/confirmation
	ResourceModel interface{}
}

// MarshalResponse converts a progress event into a useable reponse
// for the CloudFormation Resource Provider service to understand.
func (pevt *ProgressEvent) MarshalResponse() (*Response, error) {
	resp := NewResponse()

	resp.operationStatus = pevt.OperationStatus
	resp.message = pevt.Message

	if len(pevt.HandlerErrorCode) == 0 {
		resp.errorCode = cfnerr.New(pevt.HandlerErrorCode, pevt.Message, nil)
	}

	if pevt.ResourceModel != nil {
		resp.ResourceModel = pevt.ResourceModel
	}

	return resp, nil
}

func (pevt *ProgressEvent) MarshalCallback() (context.Context, int64) {
	return p.CallbackContext, p.CallbackDelaySeconds
}

// NewFailedEvent creates a generic failure progress event based on
// an error passed in.
func NewFailedEvent(err cfnerr.Error) *ProgressEvent {
	return &ProgressEvent{
		OperationStatus:  operationstatus.Failed,
		Message:          err.Message(),
		HandlerErrorCode: err.Code(),
	}
}

// NewEvent creates a new progress event
// By using this we can abstract certain aspects away from the user when needed.
func NewEvent() *ProgressEvent {
	return &ProgressEvent{
		CallbackContext: context.Background(),
		OperationStatus: operationstatus.Unknown,
	}
}
