package cfnerr

const (
	// Unknown ...
	UnknownError = "UNKNOWN"

	// NotUpdatable is when the customer tried perform an update to a property that is CreateOnly. Only
	// applicable to Update Handler. (Terminal)
	NotUpdatable = "NOTUPDATABLE"

	// InvalidRequest is a generic exception caused by invalid input from the customer. (Terminal)
	InvalidRequest = "INVALIDREQUEST"

	//AccessDenied is when the customer has insufficient permissions to perform this action. (Terminal)
	AccessDenied = "ACCESSDENIED"

	//InvalidCredentials is when the customer's provided credentials were invalid. (Terminal)
	InvalidCredentials = "INVALIDCREDENTIALS"

	//AlreadyExists is when the specified resource already existed prior to the execution of the handler.
	//Only applicable to Create Handler (Terminal) Handlers MUST return this error
	//when duplicate creation requests are received.
	AlreadyExists = "ALREADYEXISTS"

	//NotFound is when the specified resource does not exist, or is in a terminal, inoperable, and
	//irrecoverable state. (Terminal)
	NotFound = "NOTFOUND"

	//ResourceConflict is when the resource is temporarily unable to be acted upon; for example, if the
	//resource is currently undergoing an operation and cannot be acted upon until
	//that operation is finished (Retriable)
	ResourceConflict = "RESOURCECONFLICT"

	//Throttling is when the request was throttled by the downstream service. (Retriable)
	Throttling = "THROTTLING"

	//ServiceLimitExceeded is when a non-transient resource limit was reached on the service side. (Terminal)
	ServiceLimitExceeded = "SERVICELIMITEXCEEDED"

	//NotStabilized is when the downstream resource failed to complete all of its ready state checks.
	//(Retriable)
	NotStabilized = "NOTSTABILIZED"

	//GeneralServiceException is an exception from the downstream service that does not map to any other error
	//codes. (Terminal)
	GeneralServiceException = "GENERALSERVICEEXCEPTION"

	//ServiceInternalis when the downstream service returned an internal error, typically with a 5XX HTTP
	//code. (Retriable)
	ServiceInternalError = "SERVICEINTERNALERROR"

	//NetworkFailure is when the request was unable to be completed due to networking issues, such as
	//failure to receive a response from the server. (Retriable)
	NetworkFailure = "NETWORKFAILURE"

	// InternalFailure is an unexpected error occurred within the handler, such as an NPE, etc.
	//(Terminal)
	InternalFailure = "INTERNALFAILURE"
)
