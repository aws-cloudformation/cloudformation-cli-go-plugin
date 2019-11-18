package cfn

import (
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
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

//MockedMetrics mocks the call to AWS CloudWatch Metrics
type MockedMetrics struct {
	cloudwatchiface.CloudWatchAPI
	ResourceTypeName               string
	HandlerExceptionCount          int
	HandlerInvocationDurationCount int
	HandlerInvocationCount         int
}

//NewMockedMetrics is a factory function that returns a new MockedMetrics.
func NewMockedMetrics() *MockedMetrics {
	return &MockedMetrics{}
}

//PutMetricData mocks the PutMetricData method.
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
