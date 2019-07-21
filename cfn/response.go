package cfn

// A HandlerResponse describes the response of the handler.
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
