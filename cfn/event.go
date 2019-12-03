package cfn

import (
	"encoding/json"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/credentials"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/encoding"
	"github.com/aws/aws-sdk-go/aws/session"

	"gopkg.in/validator.v2"
)

// Event base structure, it will be internal to the RPDK.
type event struct {
	AWSAccountID        string         `json:"awsAccountId"`
	BearerToken         string         `json:"bearerToken" validate:"nonzero"`
	Region              string         `json:"region" validate:"nonzero"`
	Action              string         `json:"action"`
	ResponseEndpoint    string         `json:"responseEndpoint" validate:"nonzero"`
	ResourceType        string         `json:"resourceType" validate:"nonzero"`
	ResourceTypeVersion encoding.Float `json:"resourceTypeVersion"`
	RequestContext      requestContext `json:"requestContext"`
	RequestData         requestData    `json:"requestData"`
	StackID             string         `json:"stackId"`

	NextToken string
}

// RequestData is internal to the RPDK. It contains a number of fields that are for
// internal use only.
type requestData struct {
	CallerCredentials          credentials.CloudFormationCredentialsProvider `json:"callerCredentials"`
	PlatformCredentials        credentials.CloudFormationCredentialsProvider `json:"platformCredentials"`
	LogicalResourceID          string                                        `json:"logicalResourceId"`
	ResourceProperties         json.RawMessage                               `json:"resourceProperties"`
	PreviousResourceProperties json.RawMessage                               `json:"previousResourceProperties"`
	PreviousStackTags          tags                                          `json:"previousStackTags"`
	ProviderCredentials        credentials.CloudFormationCredentialsProvider `json:"providerCredentials"`
	ProviderLogGroupName       string                                        `json:"providerLogGroupName"`
	StackTags                  tags                                          `json:"stackTags"`
	SystemTags                 tags                                          `json:"systemTags"`
}

// requestContext handles elements such as retries and long running creations.
//
// Updating the requestContext key will do nothing in subsequent requests or retries,
// instead you should opt to return your context items in the action
type requestContext struct {
	CallbackContext          map[string]interface{} `json:"callbackContext,omitempty"`
	CloudWatchEventsRuleName string                 `json:"cloudWatchEventsRuleName,omitempty"`
	CloudWatchEventsTargetID string                 `json:"cloudWatchEventsTargetId,omitempty"`
	Invocation               encoding.Int           `json:"invocation,omitempty"`
	Session                  *session.Session       `json:"session,omitempty"`
}

// validateEvent ensures the event struct generated from the Lambda SDK is correct
// A number of the RPDK values are required to be a certain type/length
func validateEvent(event *event) error {
	if err := validator.Validate(event); err != nil {
		return cfnerr.New(validationError, "Failed Validation", err)
	}

	return nil
}
