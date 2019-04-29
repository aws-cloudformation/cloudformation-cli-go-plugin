package proxy

import (
	"encoding/json"
)

//ResourceHandlerRequest repersents a request set to the resource CRUD handlers.
type ResourceHandlerRequest struct {
	AwsAccountID        string
	NextToken           string
	Region              string
	ResourceType        string
	ResourceTypeVersion string
}

//RequestData  repersents the data used to build the resource handler request.
type RequestData struct {
	LogicalResourceID          string                 `json:"logicalResourceId"`
	ResourceProperties         json.RawMessage        `json:"ResourceProperties"`
	PreviousResourceProperties json.RawMessage        `json:"PreviousResourceProperties"`
	SystemTags                 map[string]interface{} `json:"systemTags"`
	StackTags                  map[string]interface{} `json:"stackTags"`
	PreviousStackTags          map[string]interface{} `json:"previousStackTags"`
}

//HandlerRequest repersents the request made from the Cloudformation service.
type HandlerRequest struct {
	//The AWS account ID
	AwsAccountID string `json:"awsAccountId"`
	//The Bearer token
	BearerToken string `json:"bearerToken"`
	//NextToken
	NextToken string `json:"nextToken "`
	//The Region
	Region string `json:"region"`
	//Action: CREATE, UPDATE, DELETE, LIST, READ
	Action string `json:"action"`
	//The resource type
	ResourceType string `json:"resourceType"`
	//The version of the resource
	ResourceTypeVersion string `json:"resourceTypeVersion"`
	//The context of the call
	Context RequestContext `json:"requestContext"`
	//The resource Handler Data
	Data RequestData `json:"requestData"`
	//The StackID
	StackID string `json:"stackId"`
}

// RequestContext represents the context of the current invocation.
type RequestContext struct {

	//The number of times the handler has been invoked (including current)
	Invocation int `json:"invocation"`

	//Custom context object to enable handlers to process re-invocation
	CallbackContext interface{} `json:"callbackContext"`

	//If the request was the result of a CloudWatchEvents re-invoke trigger the
	//CloudWatchEvents Rule name is stored to allow cleanup
	CloudWatchEventsRuleName string `json:"cloudWatchEventsRuleName"`

	//If the request was the result of a CloudWatchEvents re-invoke trigger the
	//CloudWatchEvents Trigger Id is stored to allow cleanup
	CloudWatchEventsTargetID string `json:"cloudWatchEventsTargetId"`
}
