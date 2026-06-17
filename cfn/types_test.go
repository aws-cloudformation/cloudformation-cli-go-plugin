package cfn

import (
	"context"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

// EmptyHandler is an implementation of Handler for testing.
type EmptyHandler struct{}

func (h *EmptyHandler) Create(request handler.Request) handler.ProgressEvent {
	return handler.ProgressEvent{}
}
func (h *EmptyHandler) Read(request handler.Request) handler.ProgressEvent {
	return handler.ProgressEvent{}
}
func (h *EmptyHandler) Update(request handler.Request) handler.ProgressEvent {
	return handler.ProgressEvent{}
}
func (h *EmptyHandler) Delete(request handler.Request) handler.ProgressEvent {
	return handler.ProgressEvent{}
}
func (h *EmptyHandler) List(request handler.Request) handler.ProgressEvent {
	return handler.ProgressEvent{}
}

// MockHandler is an implementation of Handler for testing.
type MockHandler struct {
	fn func(callback map[string]interface{}, cfg *aws.Config) handler.ProgressEvent
}

func (m *MockHandler) Create(request handler.Request) handler.ProgressEvent {
	return m.fn(request.CallbackContext, request.Config)
}
func (m *MockHandler) Read(request handler.Request) handler.ProgressEvent {
	return m.fn(request.CallbackContext, request.Config)
}
func (m *MockHandler) Update(request handler.Request) handler.ProgressEvent {
	return m.fn(request.CallbackContext, request.Config)
}
func (m *MockHandler) Delete(request handler.Request) handler.ProgressEvent {
	return m.fn(request.CallbackContext, request.Config)
}
func (m *MockHandler) List(request handler.Request) handler.ProgressEvent {
	return m.fn(request.CallbackContext, request.Config)
}

// MockedMetrics mocks CloudWatch for testing.
type MockedMetrics struct {
	ResourceTypeName               string
	HandlerExceptionCount          int
	HandlerInvocationDurationCount int
	HandlerInvocationCount         int
}

func NewMockedMetrics() *MockedMetrics {
	return &MockedMetrics{}
}

func (m *MockedMetrics) PutMetricData(ctx context.Context, in *cloudwatch.PutMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error) {
	m.ResourceTypeName = *in.Namespace
	d := in.MetricData[0].MetricName
	switch *d {
	case "HandlerException":
		m.HandlerExceptionCount++
	case "HandlerInvocationDuration":
		m.HandlerInvocationDurationCount++
	case "HandlerInvocationCount":
		m.HandlerInvocationCount++
	}
	return nil, nil
}

// MockModel mocks a resource model for testing.
type MockModel struct {
	Property1 *string `json:"property1,omitempty"`
	Property2 *string `json:"property2,omitempty"`
}

// MockModelHandler is an implementation of Handler for testing.
type MockModelHandler struct {
	fn func(r handler.Request) handler.ProgressEvent
}

func (m *MockModelHandler) Create(request handler.Request) handler.ProgressEvent {
	return m.fn(request)
}
func (m *MockModelHandler) Read(request handler.Request) handler.ProgressEvent {
	return m.fn(request)
}
func (m *MockModelHandler) Update(request handler.Request) handler.ProgressEvent {
	return m.fn(request)
}
func (m *MockModelHandler) Delete(request handler.Request) handler.ProgressEvent {
	return m.fn(request)
}
func (m *MockModelHandler) List(request handler.Request) handler.ProgressEvent {
	return m.fn(request)
}
