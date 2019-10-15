package cfn

import (
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

type EmptyHandlers struct{}

func (h *EmptyHandlers) Create(request Request, rc *RequestContext) (Response, error) {
	return nil, nil
}

func (h *EmptyHandlers) Read(request Request, rc *RequestContext) (Response, error) {
	return nil, nil
}

func (h *EmptyHandlers) Update(request Request, rc *RequestContext) (Response, error) {
	return nil, nil
}

func (h *EmptyHandlers) Delete(request Request, rc *RequestContext) (Response, error) {
	return nil, nil
}

func (h *EmptyHandlers) List(request Request, rc *RequestContext) (Response, error) {
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
