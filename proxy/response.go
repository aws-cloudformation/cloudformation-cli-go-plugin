package proxy

//ProgressEvent repersents the progress of CRUD handlers.
type ProgressEvent struct {

	/**
	 * The status indicates whether the handler has reached a terminal state or is
	 * still computing and requires more time to complete
	 */
	OperationStatus string

	/**
	 * If OperationStatus is FAILED or IN_PROGRESS, an error code should be provided
	 */
	HandlerErrorCode string

	/**
	 * The handler can (and should) specify a contextual information message which
	 * can be shown to callers to indicate the nature of a progress transition or
	 * callback delay; for example a message indicating "propagating to edge"
	 */
	Message string

	/**
	 * The callback context is an arbitrary datum which the handler can return in an
	 * IN_PROGRESS event to allow the passing through of additional state or
	 * metadata between subsequent retries; for example to pass through a Resource
	 * identifier which can be used to continue polling for stabilization
	 */
	CallbackContext interface{}

	/**
	 * A callback will be scheduled with an initial delay of no less than the number
	 * of seconds specified in the progress event. Set this value to <= 0 to
	 * indicate no callback should be made.
	 */
	CallbackDelaySeconds int

	// The output resource instance populated by a READ/LIST for synchronous results
	// and by CREATE/UPDATE/DELETE for final response validation/confirmation
	ResourceModel interface{}

	/**
	 * The output resource instances populated by a LIST for synchronous results
	 */

	ResourceModels []interface{}
}

type HandlerResponse struct {
	BearerToken       string            `json:"bearerToken"`
	ErrorCode         string            `json:"errorCode"`
	Message           string            `json:"message"`
	NextToken         string            `json:"nextToken"`
	OperationStatus   string            `json:"operationStatus"`
	ResourceModel     interface{}       `json:"resourceModel"`
	StabilizationData stabilizationData `json:"stabilizationData"`
}

type stabilizationData struct {
	DelayBase         int32  `json:"delayBase "`
	InitialDelay      int    `json:"initialDelay"`
	MaxDelay          int    `json:"maxDelay"`
	StabilizationMode string `json:"stabilizationMode"`
	StabilizationTime int    `json:"dtabilizationTime"`
}
