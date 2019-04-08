package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/aws-cloudformation-rpdk-go-plugin/internal/metric"
	"github.com/aws-cloudformation-rpdk-go-plugin/internal/platform/proxy"
	"github.com/aws-cloudformation-rpdk-go-plugin/internal/scheduler"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
)

const succeed = "\u2713"
const failed = "\u2717"

var con context.Context
var hands map[string]proxy.InvokeHandler
var sch *scheduler.CloudWatchScheduler

//MockNilHandler is a mocked resource handler.
type mockNilHandler struct {
	model proxy.Model
}

func (m mockNilHandler) HandleRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {
	return nil, nil

}

//MockFailedHandler is a mocked resource handler.
type mockFailedHandler struct {
	model proxy.Model
}

func (m mockFailedHandler) HandleRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {
	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Failed,
		HandlerErrorCode:     "Custom Fault",
		Message:              "Custom Fault",
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
		ResourceModel:        request.DesiredResourceState,
	}, nil

}

// mockInProgressHandler is a mocked resource handler.
type mockInProgressHandler struct {
	model proxy.Model
}

func (m mockInProgressHandler) HandleRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {
	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.InProgress,
		CallbackDelayMinutes: 5,
		ResourceModel:        request.DesiredResourceState,
	}, nil

}

//MockFailedHandler is a mocked resource handler.
type mockCompleteSynchronouslyHandler struct {
	model proxy.Model
}

func (m mockCompleteSynchronouslyHandler) HandleRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {
	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Complete,
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
		ResourceModel:        request.DesiredResourceState,
	}, nil

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

//MockedMetrics mocks the call to AWS CloudWatch
type MockedMetrics struct {
	cloudwatchiface.CloudWatchAPI
	ResourceTypeName               string
	HandlerExceptionCount          int
	HandlerInvocationDurationCount int
	HandlerInvocationCount         int
}

func New() *MockedMetrics {
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

//MockContext is a mocked Context struct
type mockContext struct{}

func (mc mockContext) Deadline() (deadline time.Time, ok bool) {
	return time.Now(), true

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

//Load the request data from file.
func loadData(theRequest *proxy.HandlerRequest, path string) (*proxy.HandlerRequest, error) {

	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(dat, &theRequest); err != nil {

		return nil, err
	}

	return theRequest, nil

}

//Set the context.
func setContext() {
	con = mockContext{}

}

func loadNilHanders() map[string]proxy.InvokeHandler {
	hands = make(map[string]proxy.InvokeHandler, 5)
	hands["CREATE"] = mockNilHandler{}
	hands["DELETE"] = mockNilHandler{}
	hands["LIST"] = mockNilHandler{}
	hands["READ"] = mockNilHandler{}
	hands["UPDATE"] = mockNilHandler{}

	return hands

}

func loadFailedHanders() map[string]proxy.InvokeHandler {
	hands = make(map[string]proxy.InvokeHandler, 5)
	hands["CREATE"] = mockFailedHandler{}
	hands["DELETE"] = mockFailedHandler{}
	hands["LIST"] = mockFailedHandler{}
	hands["READ"] = mockFailedHandler{}
	hands["UPDATE"] = mockFailedHandler{}

	return hands

}

func loadCompleteSynchronouslyHanders() map[string]proxy.InvokeHandler {
	hands = make(map[string]proxy.InvokeHandler, 5)
	hands["CREATE"] = mockCompleteSynchronouslyHandler{}
	hands["DELETE"] = mockCompleteSynchronouslyHandler{}
	hands["LIST"] = mockCompleteSynchronouslyHandler{}
	hands["READ"] = mockCompleteSynchronouslyHandler{}
	hands["UPDATE"] = mockCompleteSynchronouslyHandler{}

	return hands

}

func loadCompleteInProgressHanders() map[string]proxy.InvokeHandler {
	hands = make(map[string]proxy.InvokeHandler, 5)
	hands["CREATE"] = mockInProgressHandler{}
	hands["DELETE"] = mockInProgressHandler{}
	hands["LIST"] = mockInProgressHandler{}
	hands["READ"] = mockInProgressHandler{}
	hands["UPDATE"] = mockInProgressHandler{}

	return hands

}

func Test_processInvocationNullResponse(t *testing.T) {
	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	readRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/read.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/delete.request.json")
	listRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/list.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	setContext()
	h := loadNilHanders()

	type args struct {
		in ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		args                               args
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"nil CREATE response", args{ProcessInvocationInput{mockContext{}, *createRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:   proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
			ResourceModel:    proxy.Model{Property1: "abc", Property2: 123},
		}, true, 1, 1, 1, 0, 0},

		{"nil READ response", args{ProcessInvocationInput{mockContext{}, *readRequest, h, metric.New(New(), readRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:   proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
			ResourceModel:    proxy.Model{Property1: "abc", Property2: 123},
		}, true, 1, 1, 1, 0, 0},

		{"nil UPDATE response", args{ProcessInvocationInput{mockContext{}, *updateRequest, h, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:   proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
			ResourceModel:    proxy.Model{Property1: "abc", Property2: 123},
		}, true, 1, 1, 1, 0, 0},

		{"nil DELETE response", args{ProcessInvocationInput{mockContext{}, *deleteRequest, h, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:   proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
			ResourceModel:    proxy.Model{Property1: "abc", Property2: 123},
		}, true, 1, 1, 1, 0, 0},

		{"nil LIST response", args{ProcessInvocationInput{mockContext{}, *listRequest, h, metric.New(New(), listRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:   proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
			ResourceModel:    proxy.Model{Property1: "abc", Property2: 123},
		}, true, 1, 1, 1, 0, 0},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\tTest: %d\tWhen checking %q for progress status %v", i, tt.name, tt.want.ProgressStatus)
			{
				m := tt.args.in.Metric.Client.(*MockedMetrics)
				e := tt.args.in.Sched.Client.(*MockedEvents)
				got, err := processInvocation(tt.args.in)
				if (err != nil) != tt.wantErr {
					t.Errorf("\t%s\tShould be able to make the processInvocation call : %v", failed, err)
					return
				}
				t.Logf("\t%s\tShould be able to make the processInvocation call.", succeed)

				if reflect.DeepEqual(got, tt.want) {
					t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.ProgressStatus)
				} else {
					t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.ProgressStatus, got.ProgressStatus)
				}

				if m.HandlerExceptionCount == tt.wantHandlerExceptionCount {
					t.Logf("\t%s\tHandlerException metric should be invoked (%v) times.", succeed, tt.wantHandlerExceptionCount)
				} else {
					t.Errorf("\t%s\tHandlerException metric should be invoked (%v) times : %v", failed, tt.want.ProgressStatus, m.HandlerExceptionCount)
				}

				if m.HandlerInvocationCount == tt.wantHandlerInvocationCount {
					t.Logf("\t%s\tHandlerInvocation metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationCount)
				} else {
					t.Errorf("\t%s\tHandlerInvocation metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationCount, m.HandlerInvocationCount)
				}

				if m.HandlerInvocationDurationCount == tt.wantHandlerInvocationDurationCount {
					t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationDurationCount)
				} else {
					t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationDurationCount, m.HandlerInvocationDurationCount)
				}
				if e.RescheduleAfterMinutesCount == tt.wantrescheduleAfterMinutesCount {
					t.Logf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times.", succeed, tt.wantrescheduleAfterMinutesCount)
				} else {
					t.Errorf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times : %v", failed, tt.wantrescheduleAfterMinutesCount, e.RescheduleAfterMinutesCount)
				}

				if e.CleanupCloudWatchEventsCount == tt.wantcleanupCloudWatchEvents {
					t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
				} else {
					t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
				}

			}
			t.Log()
		})
	}
}

func Test_processInvocationFailedResponse(t *testing.T) {
	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	readRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/read.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/delete.request.json")
	listRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/list.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	setContext()
	h := loadFailedHanders()

	type args struct {
		in ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		args                               args
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"failed CREATE response", args{ProcessInvocationInput{mockContext{}, *createRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"failed READ response", args{ProcessInvocationInput{mockContext{}, *readRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"failed UPDATE response", args{ProcessInvocationInput{mockContext{}, *updateRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"failed DELETE response", args{ProcessInvocationInput{mockContext{}, *deleteRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"failed LIST response", args{ProcessInvocationInput{mockContext{}, *listRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\tTest: %d\tWhen checking %q for progress status %v", i, tt.name, tt.want.ProgressStatus)
			{
				m := tt.args.in.Metric.Client.(*MockedMetrics)
				e := tt.args.in.Sched.Client.(*MockedEvents)
				got, err := processInvocation(tt.args.in)
				if (err != nil) != tt.wantErr {
					t.Errorf("\t%s\tShould be able to make the processInvocation call : %v", failed, err)
					return
				}
				t.Logf("\t%s\tShould be able to make the processInvocation call.", succeed)

				if reflect.DeepEqual(got, tt.want) {
					t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.ProgressStatus)
				} else {
					t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.ProgressStatus, got.ProgressStatus)
				}

				if m.HandlerExceptionCount == tt.wantHandlerExceptionCount {
					t.Logf("\t%s\tHandlerException metric should be invoked (%v) times.", succeed, tt.wantHandlerExceptionCount)
				} else {
					t.Errorf("\t%s\tHandlerException metric should be invoked (%v) times : %v", failed, tt.wantHandlerExceptionCount, m.HandlerExceptionCount)
				}

				if m.HandlerInvocationCount == tt.wantHandlerInvocationCount {
					t.Logf("\t%s\tHandlerInvocation metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationCount)
				} else {
					t.Errorf("\t%s\tHandlerInvocation metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationCount, m.HandlerInvocationCount)
				}

				if m.HandlerInvocationDurationCount == tt.wantHandlerInvocationDurationCount {
					t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationDurationCount)
				} else {
					t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationDurationCount, m.HandlerInvocationDurationCount)
				}

				if e.RescheduleAfterMinutesCount == tt.wantrescheduleAfterMinutesCount {
					t.Logf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times.", succeed, tt.wantrescheduleAfterMinutesCount)
				} else {
					t.Errorf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times : %v", failed, tt.wantrescheduleAfterMinutesCount, e.RescheduleAfterMinutesCount)
				}

				if e.CleanupCloudWatchEventsCount == tt.wantcleanupCloudWatchEvents {
					t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
				} else {
					t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
				}

			}
		})
	}
}

func Test_processInvocationCompleteSynchronouslyResponse(t *testing.T) {
	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	readRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/read.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/delete.request.json")
	listRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/list.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	setContext()
	h := loadCompleteSynchronouslyHanders()

	type args struct {
		in ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		args                               args
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"complete synchronously CREATE response", args{ProcessInvocationInput{mockContext{}, *createRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously READ response", args{ProcessInvocationInput{mockContext{}, *readRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously UPDATE response", args{ProcessInvocationInput{mockContext{}, *updateRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously DELETE response", args{ProcessInvocationInput{mockContext{}, *deleteRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously LIST response", args{ProcessInvocationInput{mockContext{}, *listRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\tTest: %d\tWhen checking %q for progress status %v", i, tt.name, tt.want.ProgressStatus)
			{
				m := tt.args.in.Metric.Client.(*MockedMetrics)
				e := tt.args.in.Sched.Client.(*MockedEvents)
				got, err := processInvocation(tt.args.in)
				if (err != nil) != tt.wantErr {
					t.Errorf("\t%s\tShould be able to make the processInvocation call : %v", failed, err)
					return
				}
				t.Logf("\t%s\tShould be able to make the processInvocation call.", succeed)

				if reflect.DeepEqual(got, tt.want) {
					t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.ProgressStatus)
				} else {
					t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.ProgressStatus, got.ProgressStatus)
				}

				if m.HandlerExceptionCount == tt.wantHandlerExceptionCount {
					t.Logf("\t%s\tHandlerException metric should be invoked (%v) times.", succeed, tt.wantHandlerExceptionCount)
				} else {
					t.Errorf("\t%s\tHandlerException metric should be invoked (%v) times : %v", failed, tt.wantHandlerExceptionCount, m.HandlerExceptionCount)
				}

				if m.HandlerInvocationCount == tt.wantHandlerInvocationCount {
					t.Logf("\t%s\tHandlerInvocation metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationCount)
				} else {
					t.Errorf("\t%s\tHandlerInvocation metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationCount, m.HandlerInvocationCount)
				}

				if m.HandlerInvocationDurationCount == tt.wantHandlerInvocationDurationCount {
					t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationDurationCount)
				} else {
					t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationDurationCount, m.HandlerInvocationDurationCount)
				}

				if e.RescheduleAfterMinutesCount == tt.wantrescheduleAfterMinutesCount {
					t.Logf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times.", succeed, tt.wantrescheduleAfterMinutesCount)
				} else {
					t.Errorf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times : %v", failed, tt.wantrescheduleAfterMinutesCount, e.RescheduleAfterMinutesCount)
				}

				if e.CleanupCloudWatchEventsCount == tt.wantcleanupCloudWatchEvents {
					t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
				} else {
					t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
				}

			}
		})
	}
}

func Test_processInvocationInProgressWithContextResponse(t *testing.T) {
	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.with-request-context.request.json")
	//readRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/read.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.with-request-context.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/delete.with-request-context.request.json")
	//listRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/list.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	setContext()
	h := loadCompleteInProgressHanders()

	type args struct {
		in ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		args                               args
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"in progress CREATE response", args{ProcessInvocationInput{mockContext{}, *createRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.InProgress,
			CallbackDelayMinutes: 5,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 1, 1},
		/*
			{"in progress READ response", args{ProcessInvocationInput{mockContext{}, *readRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
				ProgressStatus:       proxy.InProgress,
				CallbackDelayMinutes: 5,
				ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
			}, false, 0, 1, 1, 1, 0},
		*/
		{"in progress UPDATE response", args{ProcessInvocationInput{mockContext{}, *updateRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.InProgress,
			CallbackDelayMinutes: 5,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 1, 1},

		{"in progress DELETE response", args{ProcessInvocationInput{mockContext{}, *deleteRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.InProgress,
			CallbackDelayMinutes: 5,
			ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 1, 1},
		/*
			{"in progress LIST response", args{ProcessInvocationInput{mockContext{}, *listRequest, h, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
				ProgressStatus:       proxy.InProgress,
				CallbackDelayMinutes: 5,
				ResourceModel:        proxy.Model{Property1: "abc", Property2: 123},
			}, false, 0, 1, 1, 1, 0},
		*/
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\tTest: %d\tWhen checking %q for progress status %v", i, tt.name, tt.want.ProgressStatus)
			{
				m := tt.args.in.Metric.Client.(*MockedMetrics)
				e := tt.args.in.Sched.Client.(*MockedEvents)
				got, err := processInvocation(tt.args.in)
				if (err != nil) != tt.wantErr {
					t.Errorf("\t%s\tShould be able to make the processInvocation call : %v", failed, err)
					return
				}
				t.Logf("\t%s\tShould be able to make the processInvocation call.", succeed)

				if got.ProgressStatus == tt.want.ProgressStatus {
					t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.ProgressStatus)
				} else {
					t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.ProgressStatus, got.ProgressStatus)
				}

				if m.HandlerExceptionCount == tt.wantHandlerExceptionCount {
					t.Logf("\t%s\tHandlerException metric should be invoked (%v) times.", succeed, tt.wantHandlerExceptionCount)
				} else {
					t.Errorf("\t%s\tHandlerException metric should be invoked (%v) times : %v", failed, tt.wantHandlerExceptionCount, m.HandlerExceptionCount)
				}

				if m.HandlerInvocationCount == tt.wantHandlerInvocationCount {
					t.Logf("\t%s\tHandlerInvocation metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationCount)
				} else {
					t.Errorf("\t%s\tHandlerInvocation metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationCount, m.HandlerInvocationCount)
				}

				if m.HandlerInvocationDurationCount == tt.wantHandlerInvocationDurationCount {
					t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationDurationCount)
				} else {
					t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationDurationCount, m.HandlerInvocationDurationCount)
				}

				if e.RescheduleAfterMinutesCount == tt.wantrescheduleAfterMinutesCount {
					t.Logf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times.", succeed, tt.wantrescheduleAfterMinutesCount)
				} else {
					t.Errorf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times : %v", failed, tt.wantrescheduleAfterMinutesCount, e.RescheduleAfterMinutesCount)
				}

				if e.CleanupCloudWatchEventsCount == tt.wantcleanupCloudWatchEvents {
					t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
				} else {
					t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
				}

			}
		})
	}
}
