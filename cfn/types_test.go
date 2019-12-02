package cfn

import (
	"context"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/scheduler"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

// EmptyHandler is a implementation of Handler
//
// This implementation of the handlers is only used for testing.
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

// MockHandler is a implementation of Handler
//
// This implementation of the handlers is only used for testing.
type MockHandler struct {
	fn func() handler.ProgressEvent
}

func (m *MockHandler) Create(request handler.Request) handler.ProgressEvent {
	return m.fn()
}

func (m *MockHandler) Read(request handler.Request) handler.ProgressEvent {
	return m.fn()
}

func (m *MockHandler) Update(request handler.Request) handler.ProgressEvent {
	return m.fn()
}

func (m *MockHandler) Delete(request handler.Request) handler.ProgressEvent {
	return m.fn()
}

func (m *MockHandler) List(request handler.Request) handler.ProgressEvent {
	return m.fn()
}

//MockedMetrics mocks the call to AWS CloudWatch Metrics
//
// This implementation of the handlers is only used for testing.
type MockedMetrics struct {
	cloudwatchiface.CloudWatchAPI
	ResourceTypeName               string
	HandlerExceptionCount          int
	HandlerInvocationDurationCount int
	HandlerInvocationCount         int
}

//NewMockedMetrics is a factory function that returns a new MockedMetrics.
//
// This implementation of the handlers is only used for testing.
func NewMockedMetrics() *MockedMetrics {
	return &MockedMetrics{}
}

//PutMetricData mocks the PutMetricData method.
//
// This implementation of the handlers is only used for testing.
func (m *MockedMetrics) PutMetricData(in *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {
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

//MockScheduler mocks the reinvocation scheduler.
//
// This implementation of the handlers is only used for testing.
type MockScheduler struct {
	Err    error
	Result *scheduler.Result
}

func (m MockScheduler) Reschedule(lambdaCtx context.Context, secsFromNow int64, callbackRequest string, invocationIDS *scheduler.ScheduleIDS) (*scheduler.Result, error) {
	return m.Result, m.Err
}
