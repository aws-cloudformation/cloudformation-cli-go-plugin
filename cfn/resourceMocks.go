package cfn

import (
	"encoding/json"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/proxy"
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

type tFunc func(resource mockCustomResource) (*proxy.ProgressEvent, error)

type mockCustomResource struct {
	Property1 string `json:"property1"`
	Property2 int    `json:"property2"`
}

type mockCallBackContext struct {
	Count int
}

type mockResourceHandler struct {
	DesiredResourceState  mockCustomResource
	PreviousResourceState mockCustomResource
	TestFunction          tFunc
}

func NewMockResourceHandler(tr tFunc) *mockResourceHandler {

	h := mockResourceHandler{
		TestFunction: tr,
	}

	return &h

}

func (m *mockResourceHandler) CreateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error) {

	return m.TestFunction(m.DesiredResourceState)
}

func (m *mockResourceHandler) DeleteRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error) {

	return m.TestFunction(m.DesiredResourceState)
}

func (m *mockResourceHandler) ListRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error) {

	return m.TestFunction(m.DesiredResourceState)
}

func (m *mockResourceHandler) ReadRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error) {

	return m.TestFunction(m.DesiredResourceState)
}

func (m *mockResourceHandler) UpdateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error) {

	return m.TestFunction(m.DesiredResourceState)
}

type MockHandlerResourceNoDesired struct {
	Property1 string `json:"property1"`
	Property2 int    `json:"property2"`
}

type MockHandlerNoDesired struct {
	PreviousResourceState MockHandlerResourceNoDesired
	ReturnState           *proxy.ProgressEvent
}

func NewMockNoDesired(state *proxy.ProgressEvent, stateError error) *MockHandlerNoDesired {

	h := MockHandlerNoDesired{
		ReturnState: state,
	}
	return &h

}

func (m *MockHandlerNoDesired) CreateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoDesired) DeleteRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoDesired) ListRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoDesired) ReadRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoDesired) UpdateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

type MockHandlerResourceNoPre struct {
	Property1 string `json:"property1"`
	Property2 int    `json:"property2"`
}

type MockHandlerNoPre struct {
	DesiredResourceState MockHandlerResourceNoPre
	ReturnState          *proxy.ProgressEvent
}

func NewMockNoPre(state *proxy.ProgressEvent, stateError error) *MockHandlerNoPre {

	h := MockHandlerNoPre{
		ReturnState: state,
	}
	return &h

}

func (m *MockHandlerNoPre) CreateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoPre) DeleteRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoPre) ListRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoPre) ReadRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

func (m *MockHandlerNoPre) UpdateRequest(request *ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}
