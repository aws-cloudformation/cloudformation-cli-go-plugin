package proxy

//ProgressEvent repersents the progress of CRUD handlers.
type ProgressEvent struct {
	//The status indicates whether the handler has reached a terminal state or
	//is still computing and requires more time to complete

	OperationStatus string `json:"operationStatus"`

	//If OperationStatus is FAILED, an error code should be provided
	HandlerErrorCode string `json:"handlerErrorCode"`

	// The handler can (and should) specify a contextual information message which
	// can be shown to callers to indicate the nature of a progress transition
	// or callback delay; for example a message indicating "propagating to edge"
	Message string `json:"message"`

	//The handler can (and should) specify a contextual information message which
	// can be shown to callers to indicate the nature of a progress transition
	// or callback delay; for example a message indicating "propagating to edge"
	//Message          string `json:"message"`
	//message: The handler can (and should) specify a contextual information message which can be shown to callers to indicate the nature of a progress transition or callback delay; for example a message indicating "propagating to edge"
	CallbackContext interface{} `json:"callbackContext"`

	// A callback will be scheduled with an initial delay of no less than
	// the number of minutes specified in the progress event. Set this
	// value to <= 0 to indicate no callback should be made.
	CallbackDelayMinutes int `json:"callbackDelayMinutes"`

	// The output resource instance populated by a READ/LIST for synchronous results
	// and by CREATE/UPDATE/DELETE for final response validation/confirmation
	ResourceModel interface{} `json:"resourceModel"`
}

type HandlerResponse struct {
	BearerToken       string                 `json:"bearerToken"`
	ErrorCode         string                 `json:"errorCode"`
	Message           string                 `json:"message"`
	NextToken         string                 `json:"nextToken "`
	ProgressStatus    string                 `json:"progressStatus"`
	ResponseData      map[string]interface{} `json:"responseData"`
	StabilizationData map[string]interface{} `json:"stabilizationData"`
}

//Reponse is used to return the response of invoking the Lambda function to the caller.
//status: indicates whether the handler has reached a terminal state or is still computing and requires more time to complete
//message: The handler can (and should) specify a contextual information message which can be shown to callers to indicate the nature of a progress transition or callback delay; for example a message indicating "propagating to edge"
//resourceModel: The output resource instance populated by a READ/LIST for synchronous results and by CREATE/UPDATE/DELETE for final response validation/confirmation.
type Reponse struct {
	Status        string
	Message       string
	ResourceModel interface{}
}
