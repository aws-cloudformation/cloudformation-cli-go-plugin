package cfn

import (
	"encoding/json"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/credentials"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/encoding"

	"gopkg.in/validator.v2"
)

// Event base structure, it will be internal to the RPDK.
type event struct {
	AWSAccountID        string                 `json:"awsAccountId"`
	BearerToken         string                 `json:"bearerToken" validate:"nonzero"`
	Region              string                 `json:"region" validate:"nonzero"`
	Action              string                 `json:"action"`
	ResourceType        string                 `json:"resourceType" validate:"nonzero"`
	ResourceTypeVersion encoding.Float         `json:"resourceTypeVersion"`
	CallbackContext     map[string]interface{} `json:"callbackContext,omitempty"`
	RequestData         requestData            `json:"requestData"`
	StackID             string                 `json:"stackId"`

	NextToken string
}

// RequestData is internal to the RPDK. It contains a number of fields that are for
// internal use only.
type requestData struct {
	CallerCredentials          credentials.CloudFormationCredentialsProvider `json:"callerCredentials"`
	LogicalResourceID          string                                        `json:"logicalResourceId"`
	ResourceProperties         json.RawMessage                               `json:"resourceProperties"`
	PreviousResourceProperties json.RawMessage                               `json:"previousResourceProperties"`
	ProviderCredentials        credentials.CloudFormationCredentialsProvider `json:"providerCredentials"`
	ProviderLogGroupName       string                                        `json:"providerLogGroupName"`
	StackTags                  tags                                          `json:"stackTags"`
	SystemTags                 tags                                          `json:"systemTags"`
}

// validateEvent ensures the event struct generated from the Lambda SDK is correct
// A number of the RPDK values are required to be a certain type/length
func validateEvent(event *event) error {
	if err := validator.Validate(event); err != nil {
		return cfnerr.New(validationError, "Failed Validation", err)
	}

	return nil
}

// testEvent base structure, it will be internal to the RPDK.
type testEvent struct {
	Action          string                                        `json:"action"`
	Credentials     credentials.CloudFormationCredentialsProvider `json:"credentials"`
	CallbackContext map[string]interface{}                        `json:"callbackContext"`

	Request resourceHandlerRequest
}

// resourceHandlerRequest is internal to the RPDK. It contains a number of fields that are for
// internal contract testing use only.
type resourceHandlerRequest struct {
	ClientRequestToken        string          `json:"clientRequestToken"`
	DesiredResourceState      json.RawMessage `json:"desiredResourceState"`
	PreviousResourceState     json.RawMessage `json:"previousResourceState"`
	DesiredResourceTags       tags            `json:"desiredResourceTags"`
	SystemTags                tags            `json:"systemTags"`
	AWSAccountID              string          `json:"awsAccountId"`
	AwsPartition              string          `json:"awsPartition"`
	LogicalResourceIdentifier string          `json:"logicalResourceIdentifier"`
	NextToken                 string          `json:"nextToken"`
	Region                    string          `json:"region"`
}
