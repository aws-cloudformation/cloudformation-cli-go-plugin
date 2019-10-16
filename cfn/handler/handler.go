package handler

import (
	"context"
	"encoding/json"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/operationstatus"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	MarshalingError      string = "Marshaling"
	BodyEmptyError       string = "BodyEmpty"
	SessionNotFoundError string = "SessionNotFound"
)

// NewRequest ...
func NewRequest(previousBody json.RawMessage, body json.RawMessage, logicalResourceID string, bearerToken string) *Request {
	req := &Request{
		previousResourcePropertiesBody: previousBody,
		resourcePropertiesBody:         body,
		logicalResourceID:              logicalResourceID,
		bearerToken:                    bearerToken,
	}

	return req
}

// Request ...
type Request struct {
	previousResourcePropertiesBody json.RawMessage
	resourcePropertiesBody         json.RawMessage
	logicalResourceID              string
	bearerToken                    string
}

// PreviousResourceProperties ...
func (r *Request) PreviousResourceProperties(v interface{}) error {
	if len(r.resourcePropertiesBody) == 0 {
		return cfnerr.New(BodyEmptyError, "Body is empty", nil)
	}

	if err := json.Unmarshal(r.previousResourcePropertiesBody, v); err != nil {
		return cfnerr.New(MarshalingError, "Unable to convert type", err)
	}

	return nil
}

// ResourceProperties ...
func (r *Request) ResourceProperties(v interface{}) error {
	if len(r.resourcePropertiesBody) == 0 {
		return cfnerr.New(BodyEmptyError, "Body is empty", nil)
	}

	if err := json.Unmarshal(r.resourcePropertiesBody, v); err != nil {
		return cfnerr.New(MarshalingError, "Unable to convert type", err)
	}

	return nil
}

// LogicalResourceID ...
func (r *Request) LogicalResourceID() string {
	return r.logicalResourceID
}

// BearerToken ...
func (r *Request) BearerToken() string {
	return r.bearerToken
}

// NewResponse ...
func NewResponse() *Response {
	return &Response{}
}

// NewFailedResponse ...
func NewFailedResponse(err error) *Response {
	return &Response{
		operationStatus: operationstatus.Failed,
		errorCode:       err,
		message:         err.Error(),
	}
}

// Response ...
type Response struct {
	message         string
	operationStatus operationstatus.Status
	ResourceModel   string
	BearerToken     string
	errorCode       error
}

func (r *Response) Message() string {
	return r.message
}

func (r *Response) OperationStatus() operationstatus.Status {
	return r.operationStatus
}

func (r *Response) Error() error {
	return r.errorCode
}

// MarshalJSON ...
func (r *Response) MarshalJSON() ([]byte, error) {
	return nil, nil
}

// ContextKey is used to prevent collisions within the context package
// It's used is for the CallbackContext key in the Request Context
//
// 	ctx.Value(handler.ContextKey("foo")).(Thing)
type ContextKey string

// CreateContext creates a context to pass to handlers
func CreateContext(items map[string]interface{}) context.Context {
	ctx := context.Background()

	for k, v := range items {
		ctx = context.WithValue(ctx, ContextKey(k), v)
	}

	return ctx
}

// ContextInjectSession ..
func ContextInjectSession(ctx context.Context, sess *session.Session) context.Context {
	ctx = context.WithValue(ctx, ContextKey("aws_session"), sess)

	return ctx
}

// ContextSession ..
func ContextSession(ctx context.Context) (*session.Session, error) {
	val, ok := ctx.Value(ContextKey("aws_session")).(*session.Session)
	if !ok {
		cfnErr := cfnerr.New(SessionNotFoundError, "Session not found", nil)
		return nil, cfnErr
	}

	return val, nil
}
