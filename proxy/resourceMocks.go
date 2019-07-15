package proxy

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

//MockContext is a mocked Context struct
type mockContext struct{}

func (mc mockContext) Deadline() (deadline time.Time, ok bool) {
	return time.Now().Add(time.Minute * 15), true

}
func (mc mockContext) Done() <-chan struct{} {
	return nil
}
func (mc mockContext) Err() error {
	return nil

}
func (mc mockContext) Value(key interface{}) interface{} {
	return &lambdacontext.LambdaContext{
		AwsRequestID:       "12345676787",
		InvokedFunctionArn: "arn:aws:lambda:us-east-2:123456789:function:myproject",
	}
}

type tFunc func(cb json.RawMessage) (*ProgressEvent, error)

type MockCustomResource struct {
	Property1 string `json:"property1"`
	Property2 int    `json:"property2"`
}

type MockCallBackContext struct {
	Count int
}

type MockResourceHandler struct {
	DesiredResourceState  MockCustomResource
	PreviousResourceState MockCustomResource
	TestFunction          tFunc
}

func NewMockResourceHandler(tr tFunc) *MockResourceHandler {

	h := MockResourceHandler{
		TestFunction: tr,
	}

	return &h

}

func (m *MockResourceHandler) CreateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*ProgressEvent, error) {

	return m.TestFunction(callbackContext)
}

func (m *MockResourceHandler) DeleteRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*ProgressEvent, error) {

	return m.TestFunction(callbackContext)
}

func (m *MockResourceHandler) ListRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*ProgressEvent, error) {

	return m.TestFunction(callbackContext)
}

func (m *MockResourceHandler) ReadRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*ProgressEvent, error) {

	return m.TestFunction(callbackContext)
}

func (m *MockResourceHandler) UpdateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*ProgressEvent, error) {

	return m.TestFunction(callbackContext)
}

type MockHandlerResourceNoDesired struct {
	Property1 string `json:"property1"`
	Property2 int    `json:"property2"`
}

type MockHandlerNoDesired struct {
	PreviousResourceState MockHandlerResourceNoDesired
	ReturnState           *ProgressEvent
}

func NewMockNoDesired(state *ProgressEvent, stateError error) *MockHandlerNoDesired {

	h := MockHandlerNoDesired{
		ReturnState: state,
	}
	return &h

}

func (m *MockHandlerNoDesired) CreateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoDesired) DeleteRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoDesired) ListRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoDesired) ReadRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoDesired) UpdateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent {

	return m.ReturnState
}

type MockHandlerResourceNoPre struct {
	Property1 string `json:"property1"`
	Property2 int    `json:"property2"`
}

type MockHandlerNoPre struct {
	DesiredResourceState MockHandlerResourceNoPre
	ReturnState          *ProgressEvent
}

func NewMockNoPre(state *ProgressEvent, stateError error) *MockHandlerNoPre {

	h := MockHandlerNoPre{
		ReturnState: state,
	}
	return &h

}

func (m *MockHandlerNoPre) CreateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoPre) DeleteRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoPre) ListRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoPre) ReadRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoPre) UpdateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *ProgressEvent {

	return m.ReturnState
}
