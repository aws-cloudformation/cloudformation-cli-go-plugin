package cfn

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/action"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/credentials"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/handler"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/metrics"

	"github.com/aws/aws-lambda-go/lambda"
	sdkCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"gopkg.in/validator.v2"
)

const (
	InvalidRequestError  string = "InvalidRequest"
	ServiceInternalError string = "ServiceInternal"
	UnmarshalingError    string = "UnmarshalingError"
	ValidationError      string = "Validation"
	TimeoutError         string = "Timeout"
)

const (
	Timeout time.Duration = 60 * time.Second
)

const (
	//MaxRetries is the number of times to try to call the Handler after it fails to respond.
	MaxRetries int = 3
)

// Handlers represents the actions from the AWS CloudFormation service
//
// Each action maps directly to a CloudFormation action. Every action is
// expected to return a response and/or an error.
//
// A valid error condition would be met if the resource operation failed or
// an API is no longer available.
type Handlers interface {
	// Create action
	Create(request Request, context *RequestContext) (Response, error)

	// Read action
	Read(request Request, context *RequestContext) (Response, error)

	// Update action
	Update(request Request, context *RequestContext) (Response, error)

	// Delete action
	Delete(request Request, context *RequestContext) (Response, error)

	// List action
	List(request Request, context *RequestContext) (Response, error)
}

// Event base structure, it will be internal to the RPDK.
type Event struct {
	Action              action.Action
	AWSAccountID        string `validate:"min=12"`
	BearerToken         string `validate:"nonzero"`
	Context             *RequestContext
	NextToken           string
	Region              string `validate:"nonzero"`
	RequestData         *RequestData
	ResourceType        string `validate:"nonzero"`
	ResourceTypeVersion float64
	ResponseEndpoint    string `validate:"nonzero"`
	StackID             string `validate:"nonzero"`
}

// UnmarshalJSON formats the event into a struct
func (e *Event) UnmarshalJSON(b []byte) error {
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
		return cfnerr.New(UnmarshalingError, "Unable to unmarshal the event", err)
	}

	resourceTypeVersion, err := strconv.ParseFloat(d.ResourceTypeVersion, 64)
	if err != nil {
		return cfnerr.New(UnmarshalingError, "Unable to format float64", err)
	}

	requestData := &RequestData{}
	if err := json.Unmarshal(d.RequestData, requestData); err != nil {
		return cfnerr.New(UnmarshalingError, "Unable to unmarshal the request data", err)
	}

	reqContext := &RequestContext{}
	if len(d.Context) > 0 {
		if err := json.Unmarshal(d.Context, reqContext); err != nil {
			return cfnerr.New(UnmarshalingError, "Unable to unmarshal the request context", err)
		}
	}

	reqContext.Session(credentials.SessionFromCredentialsProvider(requestData.CallerCredentials))

	e.Action = action.Convert(d.Action)
	e.AWSAccountID = d.AWSAccountID
	e.BearerToken = d.BearerToken
	e.Context = reqContext
	e.NextToken = d.NextToken
	e.Region = d.Region
	e.RequestData = requestData
	e.ResourceType = d.ResourceType
	e.ResourceTypeVersion = resourceTypeVersion
	e.ResponseEndpoint = d.ResponseEndpoint
	e.StackID = d.StackID

	return nil
}

// RequestData is internal to the RPDK. It contains a number of fields that are for
// internal use only.
type RequestData struct {
	CallerCredentials          sdkCredentials.Provider
	LogicalResourceID          string
	PlatformCredentials        sdkCredentials.Provider
	PreviousResourceProperties json.RawMessage
	PreviousStackTags          Tags
	ProviderLogGroupName       string
	ResourceProperties         json.RawMessage
	StackTags                  Tags
	SystemTags                 Tags
}

// UnmarshalJSON formats the request data into a usable struct
func (rd *RequestData) UnmarshalJSON(b []byte) error {
	var d struct {
		CallerCredentials          map[string]string
		LogicalResourceID          string
		PlatformCredentials        map[string]string
		PreviousResourceProperties json.RawMessage
		PreviousStackTags          Tags
		ProviderLogGroupName       string
		ResourceProperties         json.RawMessage
		StackTags                  Tags
		SystemTags                 Tags
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

// RequestContextKey is used to prevent collisions within the context package
// It's used is for the CallbackContext key in the Request Context
//
// 	reqContext.CallbackContext.Value(cfn.RequestContextKey("foo")).(Thing)
type RequestContextKey string

// RequestContext handles elements such as reties and long running creations.
//
// Updating the RequestContext key will do nothing in subsequent requests or retries,
// instead you should opt to return your context items in the action
type RequestContext struct {
	CallbackContext          context.Context
	CloudWatchEventsRuleName string
	CloudWatchEventsTargetID string
	Invocation               int64

	session *session.Session
}

// Session adds a session to the return context
func (rc *RequestContext) Session(s *session.Session) {
	rc.session = s
}

// GetSession returns the customer session for interaction with their AWS account
func (rc *RequestContext) GetSession() *session.Session {
	return rc.session
}

// UnmarshalJSON parses the request context into a usable struct
func (rc *RequestContext) UnmarshalJSON(b []byte) error {
	var d struct {
		CallbackContext          map[string]interface{} `json:"callbackContext,omitempty"`
		CloudWatchEventsRuleName string                 `json:"cloudWatchEventsRuleName,omitempty"`
		CloudWatchEventsTargetID string                 `json:"cloudWatchEventsTargetId,omitempty"`
		Invocation               int64                  `json:"invocation,omitempty"`
	}

	if err := json.Unmarshal(b, &d); err != nil {
		return cfnerr.New(UnmarshalingError, "Unable to unmarshal the request data", err)
	}

	ctx := context.Background()

	for k, v := range d.CallbackContext {
		ctx = context.WithValue(ctx, RequestContextKey(k), v)
	}

	rc.CallbackContext = ctx
	rc.CloudWatchEventsRuleName = d.CloudWatchEventsRuleName
	rc.CloudWatchEventsTargetID = d.CloudWatchEventsTargetID
	rc.Invocation = d.Invocation

	return nil
}

// Tags are stored as key/value paired strings
type Tags map[string]string

// EventFunc is the function signature required to execute an event from the Lambda SDK
type EventFunc func(ctx context.Context, event *Event) (Response, error)

// HandlerFunc is the signature required for all actions
type HandlerFunc func(request Request, context *RequestContext) (Response, error)

// Request will be passed to actions with customer related data, such as resource states
type Request interface {
	PreviousResourceProperties(v interface{}) error
	ResourceProperties(v interface{}) error
	LogicalResourceID() string
	BearerToken() string
}

// Response ...
type Response interface {
	json.Marshaler
}

// Router decides which handler should be invoked based on the action
//
// The Router will return a route or an error depending on the action passed in
func Router(a action.Action, h Handlers) (HandlerFunc, error) {
	// Figure out which action was called and have a "catch-all"
	switch a {
	case action.Create:
		return h.Create, nil
	case action.Read:
		return h.Read, nil
	case action.Update:
		return h.Update, nil
	case action.Delete:
		return h.Update, nil
	case action.List:
		return h.List, nil
	default:
		// No action matched, we should fail and return an InvalidRequestErrorCode
		return nil, cfnerr.New(InvalidRequestError, "No action/invalid action specified", nil)
	}
}

// ValidateEvent ensures the event struct generated from the Lambda SDK is correct
// A number of the RPDK values are required to be a certain type/length
func ValidateEvent(event *Event) error {
	if err := validator.Validate(event); err != nil {
		return cfnerr.New(ValidationError, "Failed Validation", err)
	}

	return nil
}

// Handler is the entry point to all invocations of a custom resource
func Handler(h Handlers) EventFunc {
	return func(ctx context.Context, event *Event) (Response, error) {
		platformSession := credentials.SessionFromCredentialsProvider(event.RequestData.PlatformCredentials)
		metricsPublisher := metrics.New(cloudwatch.New(platformSession))
		metricsPublisher.SetResourceTypeName(event.ResourceType)

		handlerFn, err := Router(event.Action, h)
		if err != nil {
			cfnErr := cfnerr.New(ServiceInternalError, "Unable to complete request", err)
			metricsPublisher.PublishExceptionMetric(time.Now(), event.Action, cfnErr)
			return handler.NewFailedResponse(cfnErr), cfnErr
		}

		if err := ValidateEvent(event); err != nil {
			cfnErr := cfnerr.New(InvalidRequestError, "Failed to validate input", err)
			metricsPublisher.PublishExceptionMetric(time.Now(), event.Action, cfnErr)
			return handler.NewFailedResponse(cfnErr), cfnErr
		}

		request := handler.NewRequest(
			event.RequestData.PreviousResourceProperties,
			event.RequestData.ResourceProperties,
			event.RequestData.LogicalResourceID,
			event.BearerToken,
		)

		resp, err := Invoke(handlerFn, request, event.Context, metricsPublisher, event.Action)
		if err != nil {
			cfnErr := cfnerr.New(ServiceInternalError, "Unable to complete request", err)
			metricsPublisher.PublishExceptionMetric(time.Now(), event.Action, cfnErr)
			return handler.NewFailedResponse(cfnErr), err
		}

		metricsPublisher.PublishInvocationMetric(time.Now(), event.Action)

		return resp, nil
	}
}

//Invoke handles the invocation of the handerFn.
func Invoke(handlerFn HandlerFunc, request Request, reqContext *RequestContext, metricsPublisher *metrics.Publisher, action action.Action) (Response, error) {
	attempts := 0

	for {
		attempts++
		// Create a context that is both manually cancellable and will signal
		// a cancel at the specified duration.
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		//We always defer a cancel.
		defer cancel()

		// Create a channel to received a signal that work is done.
		ch := make(chan Response, 1)

		// Create a channel to received error.
		cherror := make(chan error, 1)

		// Ask the goroutine to do some work for us.
		go func() {
			//start the timer
			start := time.Now()
			if err := metricsPublisher.PublishInvocationMetric(time.Now(), action); err != nil {
				cherror <- err
			}
			// Report the work is done.
			resp, err := handlerFn(request, reqContext)

			elapsed := time.Since(start)
			if err := metricsPublisher.PublishDurationMetric(time.Now(), action, elapsed.Seconds()*1e3); err != nil {
				cherror <- err
			}

			if err != nil {
				cherror <- err
			}

			ch <- resp
		}()

		// Wait for the work to finish. If it takes too long move on. If the function returns an error, signal the error channel.
		select {
		case e := <-cherror:
			cfnErr := cfnerr.New(TimeoutError, "Handler error", e)
			metricsPublisher.PublishExceptionMetric(time.Now(), action, cfnErr)
			//The handler returned an error.
			return nil, e

		case d := <-ch:
			//Return the response from the handler.
			return d, nil

		case <-ctx.Done():
			if attempts == MaxRetries {
				log.Printf("Handler failed to respond, retrying... attempt: %v action: %s \n", attempts, action)
				//handler failed to respond.
				cfnErr := cfnerr.New(TimeoutError, "Handler failed to respond in time", nil)
				metricsPublisher.PublishExceptionMetric(time.Now(), action, cfnErr)
				return nil, cfnErr
			}
			log.Printf("Handler failed to respond, retrying... attempt: %v action: %s \n", attempts, action)

		}
	}
}

// Start ...
func Start(h EventFunc) {
	lambda.Start(h)
}
