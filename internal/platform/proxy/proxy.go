package proxy

import "encoding/json"

const (
	InProgress = "InProgress"
	Complete   = "Complete"
	Failed     = "Failed"
)

const (
	InvalidRequest       = "InvalidRequest"
	AccessDenied         = "AccessDenied"
	InvalidCredentials   = "InvalidCredentials"
	NoOperationToPerform = "NoOperationToPerform"
	NotUpdatable         = "NotUpdatable"
	NotFound             = "NotFound"
	NotReady             = "NotRead"
	Throttling           = "Throttling"
	ServiceLimitExceeded = "ServiceLimitExceeded"
	ServiceTimeout       = "ServiceTimeout"
	ServiceException     = "ServiceException"
	NetworkFailure       = "NetworkFailure"
	InternalFailure      = "InternalFailure"
)

type ProgressEvent struct {
	ProgressStatus       string          `json:"progressStatus"`
	HandlerErrorCode     string          `json:"handlerErrorCode"`
	Message              string          `json:"message"`
	CallbackContext      json.RawMessage `json:"callbackContext"`
	CallbackDelayMinutes int             `json:"callbackDelayMinutes"`
	ResourceModel        Model           `json:"resourceModel"`
}

type Credentials struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
}

type RequestContext struct {
	Invocation               int             `json:"invocation"`
	CallbackContext          json.RawMessage `json:"callbackContext"`
	CloudWatchEventsRuleName string          `json:"cloudWatchEventsRuleName"`
	CloudWatchEventsTargetID string          `json:"cloudWatchEventsTargetId"`
}

type RequestData struct {
	Creds                      Credentials            `json:"credentials"`
	LogicalResourceID          string                 `json:"logicalResourceId"`
	ResourceProperties         Model                  `json:"ResourceProperties"`
	PreviousResourceProperties Model                  `json:"PreviousResourceProperties"`
	SystemTags                 map[string]interface{} `json:"systemTags"`
	StackTags                  map[string]interface{} `json:"stackTags"`
	PreviousStackTags          map[string]interface{} `json:"previousStackTags"`
}

type HandlerRequest struct {
	AwsAccountID        string         `json:"awsAccountId"`
	BearerToken         string         `json:"bearerToken"`
	NextToken           string         `json:"nextToken "`
	Region              string         `json:"region"`
	Action              string         `json:"action"`
	ResourceType        string         `json:"resourceType"`
	ResourceTypeVersion string         `json:"resourceTypeVersion"`
	Context             RequestContext `json:"requestContext"`
	Data                RequestData    `json:"requestData"`
	StackID             string         `json:"stackId"`
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

type ResourceHandlerRequest struct {
	AwsAccountID          string      `json:"awsAccountId"`
	NextToken             string      `json:"nextToken"`
	Region                string      `json:"region"`
	ResourceType          string      `json:"resourceType "`
	ResourceTypeVersion   string      `json:"resourceTypeVersion"`
	Cred                  Credentials `json:"credentials"`
	DesiredResourceState  Model       `json:"desiredResourceState"`
	PreviousResourceState Model       `json:"previousResourceState"`
}

type InvokeHandler interface {
	HandleRequest(request *ResourceHandlerRequest, callbackContext RequestContext) (*ProgressEvent, error)
}

func Invoke(i InvokeHandler, request *ResourceHandlerRequest, callbackContext RequestContext) (*ProgressEvent, error) {

	r, error := i.HandleRequest(request, callbackContext)

	return r, error
}

//Transform the the request into a resource handler
func Transform(r HandlerRequest) *ResourceHandlerRequest {

	return &ResourceHandlerRequest{
		AwsAccountID:          r.AwsAccountID,
		NextToken:             r.NextToken,
		Region:                r.Region,
		ResourceType:          r.ResourceType,
		ResourceTypeVersion:   r.ResourceTypeVersion,
		Cred:                  r.Data.Creds,
		DesiredResourceState:  r.Data.ResourceProperties,
		PreviousResourceState: r.Data.PreviousResourceProperties,
	}
}
