package cfn

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/credentials"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"

	"gopkg.in/validator.v2"
)

// Event base structure, it will be internal to the RPDK.
type event struct {
	Action              handler.Action
	AWSAccountID        string `validate:"min=12"`
	BearerToken         string `validate:"nonzero"`
	Context             *requestContext
	NextToken           string
	Region              string `validate:"nonzero"`
	RequestData         *requestData
	ResourceType        string `validate:"nonzero"`
	ResourceTypeVersion float64
	ResponseEndpoint    string `validate:"nonzero"`
	StackID             string `validate:"nonzero"`
}

// UnmarshalJSON formats the event into a struct
func (e *event) UnmarshalJSON(b []byte) error {
	var d struct {
		Action              string
		AWSAccountID        string
		BearerToken         string
		Context             json.RawMessage
		NextToken           string
		Region              string
		RequestData         json.RawMessage
		ResourceType        string
		ResourceTypeVersion string
		ResponseEndpoint    string
		StackID             string
	}

	if err := json.Unmarshal(b, &d); err != nil {
		return cfnerr.New(unmarshalingError, "Unable to unmarshal the event", err)
	}

	resourceTypeVersion, err := strconv.ParseFloat(d.ResourceTypeVersion, 64)
	if err != nil {
		return cfnerr.New(unmarshalingError, "Unable to format float64", err)
	}

	reqData := &requestData{}
	if err := json.Unmarshal(d.RequestData, reqData); err != nil {
		return cfnerr.New(unmarshalingError, "Unable to unmarshal the request data", err)
	}

	reqContext := &requestContext{}
	if len(d.Context) > 0 {
		if err := json.Unmarshal(d.Context, reqContext); err != nil {
			return cfnerr.New(unmarshalingError, "Unable to unmarshal the request context", err)
		}
	}

	reqContext.Session(credentials.SessionFromCredentialsProvider(reqData.CallerCredentials))

	e.Action = handler.Action(d.Action)
	e.AWSAccountID = d.AWSAccountID
	e.BearerToken = d.BearerToken
	e.Context = reqContext
	e.NextToken = d.NextToken
	e.Region = d.Region
	e.RequestData = reqData
	e.ResourceType = d.ResourceType
	e.ResourceTypeVersion = resourceTypeVersion
	e.ResponseEndpoint = d.ResponseEndpoint
	e.StackID = d.StackID

	return nil
}

// MarshalJSON ...
func (e *event) MarshalJSON() ([]byte, error) {
	var d struct {
		Action              string
		AWSAccountID        string
		BearerToken         string
		Context             interface{}
		NextToken           string
		Region              string
		RequestData         interface{}
		ResourceType        string
		ResourceTypeVersion string
		ResponseEndpoint    string
		StackID             string
	}

	d.Action = string(e.Action)
	d.AWSAccountID = e.AWSAccountID
	d.BearerToken = e.BearerToken
	d.NextToken = e.NextToken
	d.Region = e.Region
	d.ResourceType = e.ResourceType
	d.ResourceTypeVersion = fmt.Sprintf("%.2f", e.ResourceTypeVersion)
	d.ResponseEndpoint = e.ResponseEndpoint
	d.StackID = e.StackID
	d.RequestData = e.RequestData
	d.Context = e.Context

	b, err := json.Marshal(d)
	if err != nil {
		cfnErr := cfnerr.New(marshalingError, "Unable to marshal event", err)
		return nil, cfnErr
	}

	return b, nil
}

// validateEvent ensures the event struct generated from the Lambda SDK is correct
// A number of the RPDK values are required to be a certain type/length
func validateEvent(event *event) error {
	if err := validator.Validate(event); err != nil {
		return cfnerr.New(validationError, "Failed Validation", err)
	}

	return nil
}
