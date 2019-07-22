package cfn

import (
	"encoding/json"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/proxy"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

//MockContext describes a mocked version of the Context object with in a lambda request.
//MockContext implements the context.Context interface.
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

//tFunc type is a testing function that is passed into the mockResource for testing.
//The funcion simulates the response of a Invoker.
type tFunc func(resource mockCustomResource) (*proxy.ProgressEvent, error)

//mockCustomResource is a Mocked Custom Resource.
type mockCustomResource struct {
	Property1 string `json:"property1"`
	Property2 int    `json:"property2"`
}

//MockCallBackContext is a mocked CallBackContext
type MockCallBackContext struct {
	Count int
}

//MockResourceHandler describes a mocked version of the ResourceHandler object.
//MockResourceHandler implements the invoker interface.
type MockResourceHandler struct {
	DesiredResourceState  mockCustomResource
	PreviousResourceState mockCustomResource
	CallBackContext       MockCallBackContext
	TestFunction          tFunc
}

//NewMockResourceHandler returns a pointer to a new mockResourceHandler object.
func NewMockResourceHandler(tr tFunc) *MockResourceHandler {

	h := MockResourceHandler{
		TestFunction: tr,
	}

	return &h

}

//CreateRequest is the handler function for the CloudFormation create event.
func (m *MockResourceHandler) CreateRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error) {

	return m.TestFunction(m.DesiredResourceState)
}

//DeleteRequest is the handler function for the CloudFormation delete event.
func (m *MockResourceHandler) DeleteRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error) {

	return m.TestFunction(m.DesiredResourceState)
}

//ListRequest is the handler function for the CloudFormation list event.
func (m *MockResourceHandler) ListRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error) {

	return m.TestFunction(m.DesiredResourceState)
}

//ReadRequest is the handler function for the CloudFormation read event.
func (m *MockResourceHandler) ReadRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error) {

	return m.TestFunction(m.DesiredResourceState)
}

//UpdateRequest is the handler function for the CloudFormation update event.
func (m *MockResourceHandler) UpdateRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) (*proxy.ProgressEvent, error) {

	return m.TestFunction(m.DesiredResourceState)
}

//MockHandlerResourceNoDesired is a Mocked Custom Resource.
type MockHandlerResourceNoDesired struct {
	Property1 string `json:"property1"`
	Property2 int    `json:"property2"`
}

//MockHandlerNoDesired describes a mocked version of the ResourceHandler object.
//MockHandlerNoDesired  implements the invoker interface.
//Used to mock an error when a DesiredResourceState field is missing.
type MockHandlerNoDesired struct {
	PreviousResourceState MockHandlerResourceNoDesired
	ReturnState           *proxy.ProgressEvent
}

//NewMockNoDesired returns a pointer to a new MockNoDesired object.
func NewMockNoDesired(state *proxy.ProgressEvent, stateError error) *MockHandlerNoDesired {

	h := MockHandlerNoDesired{
		ReturnState: state,
	}
	return &h

}

//CreateRequest is the handler function for the CloudFormation create event.
func (m *MockHandlerNoDesired) CreateRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

//DeleteRequest is the handler function for the CloudFormation delete event.
func (m *MockHandlerNoDesired) DeleteRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

//ListRequest is the handler function for the CloudFormation list event.
func (m *MockHandlerNoDesired) ListRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

//ReadRequest is the handler function for the CloudFormation read event.
func (m *MockHandlerNoDesired) ReadRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

//UpdateRequest is the handler function for the CloudFormation update event.
func (m *MockHandlerNoDesired) UpdateRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

//MockHandlerResourceNoPre  is a Mocked Custom Resource.
type MockHandlerResourceNoPre struct {
	Property1 string `json:"property1"`
	Property2 int    `json:"property2"`
}

//MockHandlerNoPre describes a mocked version of the ResourceHandler object.
//MockHandlerNoPre implements the invoker interface.
//Used to mock an error when a PreviousResourceState field is missing.
type MockHandlerNoPre struct {
	DesiredResourceState MockHandlerResourceNoPre
	ReturnState          *proxy.ProgressEvent
}

//NewMockNoPre returns a pointer to a new MockNoPre object.
func NewMockNoPre(state *proxy.ProgressEvent, stateError error) *MockHandlerNoPre {

	h := MockHandlerNoPre{
		ReturnState: state,
	}
	return &h

}

//CreateRequest is the handler function for the CloudFormation create event.
func (m *MockHandlerNoPre) CreateRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

//DeleteRequest is the handler function for the CloudFormation delete event.
func (m *MockHandlerNoPre) DeleteRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

//ListRequest is the handler function for the CloudFormation list event.
func (m *MockHandlerNoPre) ListRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

//ReadRequest is the handler function for the CloudFormation read event.
func (m *MockHandlerNoPre) ReadRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}

//UpdateRequest is the handler function for the CloudFormation update event.
func (m *MockHandlerNoPre) UpdateRequest(request *proxy.ResourceHandlerRequest, callbackContext json.RawMessage) *proxy.ProgressEvent {

	return m.ReturnState
}
