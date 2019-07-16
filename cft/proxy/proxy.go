package proxy

//The OperationStatus of a progress event
const (
	InProgress = "IN_PROGRESS"
	Complete   = "SUCCESS"
	FAILED     = "FAILED"
)

//The list of Vaild handler error codes.
const (
	//The customer tried perform an update to a property that is CreateOnly.
	//Only applicable to Update Handler. (Terminal)
	NotUpdatable = "NotUpdatable"
	//A generic exception caused by invalid input from the customer. (Terminal)
	InvalidRequest = "InvalidRequest"

	//The customer has insufficient permissions to perform this action. (Terminal)
	AccessDenied = "AccessDenied"

	//The customer's provided credentials were invalid. (Terminal)
	InvalidCredentials = "InvalidCredentials"

	//the specified resource already existed prior to the execution of the handler.
	//Only applicable to Create Handler (Terminal) Handlers MUST return this error when duplicate creation requests are received.
	AlreadyExists = "AlreadyExists"

	//The specified resource does not exist, or is in a terminal, inoperable, and
	//irrecoverable state. (Terminal)
	NotFound = "NotFound"

	//The resource is temporarily unable to be acted upon; for example,
	//if the resource is currently undergoing an operation and cannot
	//be acted upon until that operation is finished. (Retriable)
	ResourceConflict = "ResourceConflict"

	//The request was throttled by the downstream service. (Retriable)
	Throttling = "Throttling"

	//a non-transient resource limit was reached on the service side. (Terminal)
	ServiceLimitExceeded = "ServiceLimitExceeded"

	//The downstream resource failed to complete all of its ready state checks. (Retriable)
	NotStabilized = "NotStabilized"

	//An exception from the downstream service that does not map to any other error codes. (Terminal)
	GeneralServiceException = "GeneralServiceException"

	//the downstream service returned an internal error, typically with a 5XX HTTP
	//Status code (Retriable)

	ServiceInternalError = "ServiceInternalError"

	//The request was unable to be completed due to networking issues, such as
	//failure to receive a response from the server. (Retriable)
	NetworkFailure = "NetworkFailure"

	//An unexpected error occurred within the handler, such as an NPE, etc. (Terminal)

	InternalFailure = "InternalFailure"
)

//ProgressEvent represent the progress of CRUD handlers.
type ProgressEvent struct {
	//The status indicates whether the handler has reached a terminal state or is
	//still computing and requires more time to complete.
	OperationStatus string

	//If OperationStatus is FAILED or IN_PROGRESS, an error code should be provided.
	HandlerErrorCode string

	//The handler can (and should) specify a contextual information message which
	//can be shown to callers to indicate the nature of a progress transition or
	//callback delay; for example a message indicating "propagating to edge."
	Message string

	//The callback context is an arbitrary datum which the handler can return in an
	//IN_PROGRESS event to allow the passing through of additional state or
	//metadata between subsequent retries; for example to pass through a Resource
	//identifier which can be used to continue polling for stabilization
	CallbackContext interface{}

	//A callback will be scheduled with an initial delay of no less than the number
	//of seconds specified in the progress event. Set this value to <= 0 to
	//indicate no callback should be made.
	CallbackDelaySeconds int

	//The output resource instance populated by a READ/LIST for synchronous results
	//and by CREATE/UPDATE/DELETE for final response validation/confirmation
	ResourceModel interface{}

	//The output resource instances populated by a LIST for synchronous results
	ResourceModels []interface{}
}

//DefaultFailureHandler is a convenience method for constructing a FAILED response.
func DefaultFailureHandler(e error, handlerErrorCode string) *ProgressEvent {

	return &ProgressEvent{
		OperationStatus:  FAILED,
		HandlerErrorCode: handlerErrorCode,
		Message:          e.Error(),
	}
}

//Failed is a convenience method for constructing a FAILED response.
func Failed(model interface{}, cxt interface{}, code string, message string) *ProgressEvent {

	return &ProgressEvent{
		OperationStatus:  FAILED,
		HandlerErrorCode: code,
		Message:          message,
		ResourceModel:    nil,
		CallbackContext:  nil,
	}
}
