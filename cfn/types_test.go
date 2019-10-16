package cfn

import (
	"context"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/handler"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

type EmptyHandlers struct{}

func (h *EmptyHandlers) Create(ctx context.Context, request handler.Request) (handler.ProgressEvent, error) {
	return nil, nil
}

func (h *EmptyHandlers) Read(ctx context.Context, request handler.Request) (handler.ProgressEvent, error) {
	return nil, nil
}

func (h *EmptyHandlers) Update(ctx context.Context, request handler.Request) (handler.ProgressEvent, error) {
	return nil, nil
}

func (h *EmptyHandlers) Delete(ctx context.Context, request handler.Request) (handler.ProgressEvent, error) {
	return nil, nil
}

func (h *EmptyHandlers) List(ctx context.Context, request handler.Request) (handler.ProgressEvent, error) {
	return nil, nil
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
