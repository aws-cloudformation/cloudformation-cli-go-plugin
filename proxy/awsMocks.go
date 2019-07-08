package proxy

import (
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
)

//MockedCloudFormation is used to mock the Cloudformation.
type MockedCloudFormation struct {
	Client cloudformationiface.CloudFormationAPI
}

func NewMockedCloudFormation() *MockedCloudFormation {
	return &MockedCloudFormation{}
}

//Todo: enable later
//func (m *MockedCloudFormation) RecordHandlerProgress(*cloudformation.RecordHandlerProgressInput) (*cloudformation.RecordHandlerProgressOutput, error) {
//	return nil, nil
//}

//MockedMetrics mocks the call to AWS CloudWatch
type MockedMetrics struct {
	cloudwatchiface.CloudWatchAPI
	ResourceTypeName               string
	HandlerExceptionCount          int
	HandlerInvocationDurationCount int
	HandlerInvocationCount         int
}

func NewMockedMetrics() *MockedMetrics {
	return &MockedMetrics{}
}

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
}

func NewmockedEvents() *MockedEvents {
	return &MockedEvents{}
}

func (m *MockedEvents) PutRule(in *cloudwatchevents.PutRuleInput) (*cloudwatchevents.PutRuleOutput, error) {
	m.RescheduleAfterMinutesCount++
	return nil, nil
}

func (m *MockedEvents) PutTargets(in *cloudwatchevents.PutTargetsInput) (*cloudwatchevents.PutTargetsOutput, error) {

	return nil, nil

}

func (m *MockedEvents) DeleteRule(*cloudwatchevents.DeleteRuleInput) (*cloudwatchevents.DeleteRuleOutput, error) {
	m.CleanupCloudWatchEventsCount++
	return nil, nil
}

func (m *MockedEvents) RemoveTargets(*cloudwatchevents.RemoveTargetsInput) (*cloudwatchevents.RemoveTargetsOutput, error) {

	return nil, nil
}
