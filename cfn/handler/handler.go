package handler

import (
	"context"
	"encoding/json"
	"io"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/operationstatus"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	// MarshalingError occurs when we can't marshal data from one format into another.
	MarshalingError string = "Marshaling"
	// BodyEmptyError happens when the resource body is empty
	BodyEmptyError string = "BodyEmpty"
	// SessionNotFoundError occurs when the AWS SDK session isn't available in the context
	SessionNotFoundError string = "SessionNotFound"

	// LogProviderNotFoundError ...
	LogProviderNotFoundError string = "LogProviderNotFound"
)

// Request will be passed to actions with customer related data, such as resource states
type Request interface {
	PreviousResourceProperties(v interface{}) error
	ResourceProperties(v interface{}) error
	LogicalResourceID() string
	BearerToken() string
}

// Response ...
type Response interface {
	Error() error
	Message() string
	OperationStatus() operationstatus.Status
}

// ProgressEvent returns the status of any given action
type ProgressEvent interface {
	MarshalResponse() (Response, error)
	MarshalCallback() (CallbackContextValues, int64)
}

// NewRequest ...
func NewRequest(previousBody json.RawMessage, body json.RawMessage, logicalResourceID string, bearerToken string) Request {
	req := &IRequest{
		previousResourcePropertiesBody: previousBody,
		resourcePropertiesBody:         body,
		logicalResourceID:              logicalResourceID,
		bearerToken:                    bearerToken,
	}

	return req
}

// IRequest ...
type IRequest struct {
	previousResourcePropertiesBody json.RawMessage
	resourcePropertiesBody         json.RawMessage
	logicalResourceID              string
	bearerToken                    string
}

// PreviousResourceProperties ...
func (r *IRequest) PreviousResourceProperties(v interface{}) error {
	if len(r.resourcePropertiesBody) == 0 {
		return cfnerr.New(BodyEmptyError, "Body is empty", nil)
	}

	if err := json.Unmarshal(r.previousResourcePropertiesBody, v); err != nil {
		return cfnerr.New(MarshalingError, "Unable to convert type", err)
	}

	return nil
}

// ResourceProperties ...
func (r *IRequest) ResourceProperties(v interface{}) error {
	if len(r.resourcePropertiesBody) == 0 {
		return cfnerr.New(BodyEmptyError, "Body is empty", nil)
	}

	if err := json.Unmarshal(r.resourcePropertiesBody, v); err != nil {
		return cfnerr.New(MarshalingError, "Unable to convert type", err)
	}

	return nil
}

// LogicalResourceID ...
func (r *IRequest) LogicalResourceID() string {
	return r.logicalResourceID
}

// BearerToken ...
func (r *IRequest) BearerToken() string {
	return r.bearerToken
}

// NewResponse ...
func NewResponse() *IResponse {
	return &IResponse{}
}

// NewFailedResponse ...
func NewFailedResponse(err error) Response {
	return &IResponse{
		operationStatus: operationstatus.Failed,
		errorCode:       err,
		message:         err.Error(),
	}
}

// IResponse ...
type IResponse struct {
	message         string                 `json:"Message,omitempty"`
	operationStatus operationstatus.Status `json:"OperationStatus,omitempty"`
	resourceModel   interface{}            `json:"ResourceModel,omitempty"`
	errorCode       error                  `json:"ErrorCode,omitempty"`
}

func (r *IResponse) MarshalJSON() ([]byte, error) {
	var resp struct {
		Message         string      `json:"Message,omitempty"`
		OperationStatus string      `json:"OperationStatus,omitempty"`
		ResourceModel   interface{} `json:"ResourceModel,omitempty"`
		ErrorCode       string      `json:"ErrorCode,omitempty"`
	}

	cfnErr, ok := r.Error().(cfnerr.Error)
	if cfnErr != nil && !ok {
		return nil, cfnerr.New(MarshalingError, "Unable to marshal response, zomg", r.Error())
	}

	resp.Message = r.Message()
	resp.OperationStatus = r.operationStatus.String()
	resp.ResourceModel = r.ResourceModel()

	if cfnErr != nil {
		resp.ErrorCode = cfnErr.Code()
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return nil, cfnerr.New(MarshalingError, "Unable to marshal response", err)
	}

	return b, nil
}

// Message ...
func (r *IResponse) Message() string {
	return r.message
}

// OperationStatus ...
func (r *IResponse) OperationStatus() operationstatus.Status {
	return r.operationStatus
}

// Error ...
func (r *IResponse) Error() error {
	return r.errorCode
}

func (r *IResponse) ResourceModel() interface{} {
	return r.resourceModel
}

// ContextKey is used to prevent collisions within the context package
// It's used is for the CallbackContext key in the Request Context
//
// 	ctx.Value(handler.ContextKey("foo")).(Thing)
type ContextKey string

// CallbackContextValues ...
type CallbackContextValues map[string]interface{}

// ContextValues creates a context to pass to handlers
func ContextValues(ctx context.Context, values CallbackContextValues) context.Context {
	return context.WithValue(ctx, ContextKey("user_callback_context"), values)
}

// ContextCallback ...
func ContextCallback(ctx context.Context) (CallbackContextValues, error) {
	values, ok := ctx.Value(ContextKey("user_callback_context")).(CallbackContextValues)
	if !ok {
		cfnErr := cfnerr.New(SessionNotFoundError, "Session not found", nil)
		return nil, cfnErr
	}

	return values, nil
}

// ContextInjectSession ..
func ContextInjectSession(ctx context.Context, sess *session.Session) context.Context {
	return context.WithValue(ctx, ContextKey("aws_session"), sess)
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

// ContextInjectLogProvider ..
func ContextInjectLogProvider(ctx context.Context, writer io.Writer) context.Context {
	return context.WithValue(ctx, ContextKey("log_provider"), writer)
}

// ContextLogProvider ..
func ContextLogProvider(ctx context.Context) (io.Writer, error) {
	val, ok := ctx.Value(ContextKey("log_provider")).(io.Writer)
	if !ok {
		cfnErr := cfnerr.New(LogProviderNotFoundError, "LogProvider not found", nil)
		return nil, cfnErr
	}

	return val, nil
}
