package proxy

import "encoding/json"

type ProgressEvent struct {
	ProgressStatus       string          `json:"progressStatus"`
	HandlerErrorCode     string          `json:"handlerErrorCode"`
	Message              string          `json:"message"`
	CallbackContext      json.RawMessage `json:"callbackContext"`
	CallbackDelayMinutes int             `json:"callbackDelayMinutes"`
	ResourceModel        interface{}     `json:"resourceModel"`
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
