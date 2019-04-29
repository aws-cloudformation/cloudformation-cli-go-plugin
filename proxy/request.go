package proxy

import (
	"encoding/json"
)

type ResourceHandlerRequest struct {
	AwsAccountID        string
	NextToken           string
	Region              string
	ResourceType        string
	ResourceTypeVersion string
	Cred                Credentials
}

type Credentials struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
}

type RequestData struct {
	Creds                      Credentials            `json:"credentials"`
	LogicalResourceID          string                 `json:"logicalResourceId"`
	ResourceProperties         json.RawMessage        `json:"ResourceProperties"`
	PreviousResourceProperties json.RawMessage        `json:"PreviousResourceProperties"`
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

type RequestContext struct {
	Invocation               int         `json:"invocation"`
	CallbackContext          interface{} `json:"callbackContext"`
	CloudWatchEventsRuleName string      `json:"cloudWatchEventsRuleName"`
	CloudWatchEventsTargetID string      `json:"cloudWatchEventsTargetId"`
}
