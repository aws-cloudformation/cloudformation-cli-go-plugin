package cfn

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
)

//MockedCloudFormation is used to mock the Cloudformation service.
type MockedCloudFormation struct {
	Client cloudformationiface.CloudFormationAPI
}

//NewMockedCloudFormation is a factory function that returns a new MockedCloudFormation.
func NewMockedCloudFormation() *MockedCloudFormation {
	return &MockedCloudFormation{}
}

//RecordHandlerProgress mocks the RecordHandlerProgress method.
func (m *MockedCloudFormation) RecordHandlerProgress(*cloudformation.RecordHandlerProgressInput) (*cloudformation.RecordHandlerProgressOutput, error) {
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

//MockedEvents mocks the call to AWS CloudWatch Events
type MockedEvents struct {
	cloudwatcheventsiface.CloudWatchEventsAPI
	RescheduleAfterMinutesCount  int
	CleanupCloudWatchEventsCount int
	RuleName                     string
	TargetName                   string
	CallBackRequest              string
}

//NewMockedEvents is a factory function that returns a new MockedEvents.
func NewMockedEvents() *MockedEvents {
	return &MockedEvents{}
}

//PutRule mocks the PutRule method.
func (m *MockedEvents) PutRule(in *cloudwatchevents.PutRuleInput) (*cloudwatchevents.PutRuleOutput, error) {
	m.RescheduleAfterMinutesCount++
	return nil, nil
}

//PutTargets mocks the PutTargets method.
func (m *MockedEvents) PutTargets(in *cloudwatchevents.PutTargetsInput) (*cloudwatchevents.PutTargetsOutput, error) {

	m.CallBackRequest = *in.Targets[0].Input
	return nil, nil

}

//DeleteRule mocks the DeleteRule method.
func (m *MockedEvents) DeleteRule(*cloudwatchevents.DeleteRuleInput) (*cloudwatchevents.DeleteRuleOutput, error) {
	m.CleanupCloudWatchEventsCount++
	return nil, nil
}

//RemoveTargets mocks the RemoveTargets method.
func (m *MockedEvents) RemoveTargets(*cloudwatchevents.RemoveTargetsInput) (*cloudwatchevents.RemoveTargetsOutput, error) {

	return nil, nil
}

// MockCloudFormationProvider mocks a credential provider,
type MockCloudFormationProvider struct {
	Client cloudformationiface.CloudFormationAPI
}

//NewMockCloudFormationProvider is a factory function that returns a new MockCloudFormationProvider.
func NewMockCloudFormationProvider(cf cloudformationiface.CloudFormationAPI) *MockCloudFormationProvider {

	return &MockCloudFormationProvider{
		Client: cf,
	}
}

//Get returns a new mocked CloudFormation service sesson.
func (c *MockCloudFormationProvider) Get() (cloudformationiface.CloudFormationAPI, error) {
	return c.Client, nil
}

//MockCloudWatchEventsProvider mocks a CloudWatchEventsProvider
type MockCloudWatchEventsProvider struct {
	Client *MockedEvents
}

//NewMockCloudWatchEventsProvider is a factory function that returns a new MockCloudWatchEventsProvider.
func NewMockCloudWatchEventsProvider(m *MockedEvents) *MockCloudWatchEventsProvider {

	return &MockCloudWatchEventsProvider{
		Client: m,
	}
}

//Get returns a new Mocked CloudWatchEvents service sesson.
func (c *MockCloudWatchEventsProvider) Get() (cloudwatcheventsiface.CloudWatchEventsAPI, error) {

	return c.Client, nil
}

//MockCloudWatchProvider mocks a CloudWatchProvider
type MockCloudWatchProvider struct {
	Client *MockedMetrics
}

//NewMockCloudWatchProvider is a factory function that returns a new MockCloudWatchEventsProvider.
func NewMockCloudWatchProvider(cw *MockedMetrics) *MockCloudWatchProvider {

	return &MockCloudWatchProvider{
		Client: cw,
	}
}

//Get returns a new Mocked CloudWatchEvents service sesson.
func (c *MockCloudWatchProvider) Get() (cloudwatchiface.CloudWatchAPI, error) {
	return c.Client, nil
}
