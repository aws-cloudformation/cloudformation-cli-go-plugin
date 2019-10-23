package cfn

import (
	"context"
	"encoding/json"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/credentials"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"

	sdkCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// RequestData is internal to the RPDK. It contains a number of fields that are for
// internal use only.
type requestData struct {
	CallerCredentials          sdkCredentials.Provider
	LogicalResourceID          string
	PlatformCredentials        sdkCredentials.Provider
	PreviousResourceProperties json.RawMessage
	PreviousStackTags          tags
	ProviderLogGroupName       string
	ResourceProperties         json.RawMessage
	StackTags                  tags
	SystemTags                 tags
}

// UnmarshalJSON formats the request data into a usable struct
func (rd *requestData) UnmarshalJSON(b []byte) error {
	var d struct {
		CallerCredentials          map[string]string
		LogicalResourceID          string
		PlatformCredentials        map[string]string
		PreviousResourceProperties json.RawMessage
		PreviousStackTags          tags
		ProviderLogGroupName       string
		ResourceProperties         json.RawMessage
		StackTags                  tags
		SystemTags                 tags
	}

	if err := json.Unmarshal(b, &d); err != nil {
		return cfnerr.New(UnmarshalingError, "Unable to unmarshal the request data", err)
	}

	rd.LogicalResourceID = d.LogicalResourceID
	rd.ProviderLogGroupName = d.ProviderLogGroupName
	rd.PreviousResourceProperties = d.PreviousResourceProperties
	rd.ResourceProperties = d.ResourceProperties
	rd.PreviousStackTags = d.PreviousStackTags
	rd.StackTags = d.StackTags
	rd.SystemTags = d.SystemTags

	rd.CallerCredentials = credentials.NewProvider(
		d.CallerCredentials["accessKeyId"],
		d.CallerCredentials["secretAccessKey"],
		d.CallerCredentials["sessionToken"],
	)

	rd.PlatformCredentials = credentials.NewProvider(
		d.PlatformCredentials["accessKeyId"],
		d.PlatformCredentials["secretAccessKey"],
		d.PlatformCredentials["sessionToken"],
	)

	return nil
}

// MarshalJSON ...
func (rd *requestData) MarshalJSON() ([]byte, error) {
	var d struct {
		CallerCredentials          map[string]string
		LogicalResourceID          string
		PlatformCredentials        map[string]string
		PreviousResourceProperties interface{}
		PreviousStackTags          tags
		ProviderLogGroupName       string
		ResourceProperties         interface{}
		StackTags                  tags
		SystemTags                 tags
	}

	// we can swallow the errors, it's never raised.
	caller, _ := rd.CallerCredentials.Retrieve()
	platform, _ := rd.CallerCredentials.Retrieve()

	d.CallerCredentials = map[string]string{
		"accessKeyId":     caller.AccessKeyID,
		"secretAccessKey": caller.SecretAccessKey,
		"sessionToken":    caller.SessionToken,
	}
	d.LogicalResourceID = rd.LogicalResourceID
	d.PlatformCredentials = map[string]string{
		"accessKeyId":     platform.AccessKeyID,
		"secretAccessKey": platform.SecretAccessKey,
		"sessionToken":    platform.SessionToken,
	}
	d.PreviousResourceProperties = rd.PreviousResourceProperties
	d.PreviousStackTags = rd.PreviousStackTags
	d.ProviderLogGroupName = rd.ProviderLogGroupName
	d.ResourceProperties = rd.ResourceProperties
	d.StackTags = rd.StackTags
	d.SystemTags = rd.StackTags

	b, err := json.Marshal(d)
	if err != nil {
		cfnErr := cfnerr.New(MarshalingError, "Unable to marshal request data", err)
		return nil, cfnErr
	}

	return b, nil
}

// requestContext handles elements such as reties and long running creations.
//
// Updating the requestContext key will do nothing in subsequent requests or retries,
// instead you should opt to return your context items in the action
type requestContext struct {
	CallbackContext          handler.CallbackContextValues
	CloudWatchEventsRuleName string
	CloudWatchEventsTargetID string
	Invocation               int64

	session *session.Session
}

// Session adds a session to the return context
func (rc *requestContext) Session(s *session.Session) {
	rc.session = s
}

// GetSession returns the customer session for interaction with their AWS account
func (rc *requestContext) GetSession() *session.Session {
	return rc.session
}

// UnmarshalJSON parses the request context into a usable struct
func (rc *requestContext) UnmarshalJSON(b []byte) error {
	var d struct {
		CallbackContext          handler.CallbackContextValues `json:"callbackContext,omitempty"`
		CloudWatchEventsRuleName string                        `json:"cloudWatchEventsRuleName,omitempty"`
		CloudWatchEventsTargetID string                        `json:"cloudWatchEventsTargetId,omitempty"`
		Invocation               int64                         `json:"invocation,omitempty"`
	}

	if err := json.Unmarshal(b, &d); err != nil {
		return cfnerr.New(UnmarshalingError, "Unable to unmarshal the request data", err)
	}

	ctx := handler.ContextValues(context.Background(), d.CallbackContext)
	callbackCtx, err := handler.ContextCallback(ctx)

	if err != nil {
		return err
	}

	rc.CallbackContext = callbackCtx
	rc.CloudWatchEventsRuleName = d.CloudWatchEventsRuleName
	rc.CloudWatchEventsTargetID = d.CloudWatchEventsTargetID
	rc.Invocation = d.Invocation

	return nil
}

// MarshalJSON ...
func (rc *requestContext) MarshalJSON() ([]byte, error) {
	var d struct {
		CallbackContext          handler.CallbackContextValues `json:"callbackContext,omitempty"`
		CloudWatchEventsRuleName string                        `json:"cloudWatchEventsRuleName,omitempty"`
		CloudWatchEventsTargetID string                        `json:"cloudWatchEventsTargetId,omitempty"`
		Invocation               int64                         `json:"invocation,omitempty"`
	}

	d.CallbackContext = rc.CallbackContext
	d.CloudWatchEventsRuleName = rc.CloudWatchEventsRuleName
	d.CloudWatchEventsTargetID = rc.CloudWatchEventsTargetID
	d.Invocation = rc.Invocation

	b, err := json.Marshal(d)
	if err != nil {
		cfnErr := cfnerr.New(MarshalingError, "Unable to marshal request context", err)
		return nil, cfnErr
	}

	return b, nil
}
