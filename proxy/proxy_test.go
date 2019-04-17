package proxy_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/scheduler"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
)

const succeed = "\u2713"
const failed = "\u2717"

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

type mockNilHandler struct{}

func NewMockNil() *mockNilHandler {

	h := mockNilHandler{}
	return &h

}

func (m *mockNilHandler) CreateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {
	return nil, nil
}

func (m *mockNilHandler) DeleteRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return nil, nil
}

func (m *mockNilHandler) ListRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return nil, nil
}

func (m *mockNilHandler) ReadRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return nil, nil
}

func (m *mockNilHandler) UpdateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return nil, nil
}

type mockFailedHandler struct{}

func NewmockFailedHandler() *mockFailedHandler {

	h := mockFailedHandler{}
	return &h

}

func (m *mockFailedHandler) CreateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {
	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Failed,
		HandlerErrorCode:     "Custom Fault",
		Message:              "Custom Fault",
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

func (m *mockFailedHandler) DeleteRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Failed,
		HandlerErrorCode:     "Custom Fault",
		Message:              "Custom Fault",
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

func (m *mockFailedHandler) ListRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Failed,
		HandlerErrorCode:     "Custom Fault",
		Message:              "Custom Fault",
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

func (m *mockFailedHandler) ReadRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Failed,
		HandlerErrorCode:     "Custom Fault",
		Message:              "Custom Fault",
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

func (m *mockFailedHandler) UpdateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Failed,
		HandlerErrorCode:     "Custom Fault",
		Message:              "Custom Fault",
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

type mockCompleteSynchronouslyHandler struct{}

func NewmockCompleteSynchronouslyHandler() *mockCompleteSynchronouslyHandler {

	h := mockCompleteSynchronouslyHandler{}
	return &h

}

func (m *mockCompleteSynchronouslyHandler) CreateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {
	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Complete,
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

func (m *mockCompleteSynchronouslyHandler) DeleteRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Complete,
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

func (m *mockCompleteSynchronouslyHandler) ListRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Complete,
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

func (m *mockCompleteSynchronouslyHandler) ReadRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Complete,
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

func (m *mockCompleteSynchronouslyHandler) UpdateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Complete,
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

type mockInProgressHandler struct{}

func NewmockInProgressHandler() *mockInProgressHandler {

	h := mockInProgressHandler{}
	return &h

}

func (m *mockInProgressHandler) CreateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {
	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.InProgress,
		CallbackDelayMinutes: 5,
	}, nil
}

func (m *mockInProgressHandler) DeleteRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.InProgress,
		CallbackDelayMinutes: 5,
	}, nil
}

func (m *mockInProgressHandler) ListRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.InProgress,
		CallbackDelayMinutes: 5,
	}, nil
}

func (m *mockInProgressHandler) ReadRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.InProgress,
		CallbackDelayMinutes: 5,
	}, nil
}

func (m *mockInProgressHandler) UpdateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.InProgress,
		CallbackDelayMinutes: 5,
	}, nil
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

func Test_processInvocationNullResponse(t *testing.T) {

	proxy.StartWithOutLambda(NewMockNil())

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	readRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/read.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/delete.request.json")
	listRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/list.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type fields struct {
		in proxy.ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		fields                             fields
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"nil CREATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *createRequest, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:   proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},
		{"nil DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:   proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},
		{"nil LIST response", fields{proxy.ProcessInvocationInput{mockContext{}, *listRequest, metric.New(New(), listRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:   proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},
		{"nil READ response", fields{proxy.ProcessInvocationInput{mockContext{}, *readRequest, metric.New(New(), readRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:   proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},
		{"nil UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:   proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(tt.fields.in)

			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation()

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
		})
	}
}

func Test_processInvocationFailedResponse(t *testing.T) {

	proxy.StartWithOutLambda(NewmockFailedHandler())

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	readRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/read.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/delete.request.json")
	listRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/list.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type fields struct {
		in proxy.ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		fields                             fields
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"failed CREATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *createRequest, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
		{"failed DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
		{"failed LIST response", fields{proxy.ProcessInvocationInput{mockContext{}, *listRequest, metric.New(New(), listRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
		{"failed READ response", fields{proxy.ProcessInvocationInput{mockContext{}, *readRequest, metric.New(New(), readRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
		{"failed UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(tt.fields.in)

			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation()
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
		})
	}
}

func Test_processInvocationCompleteSynchronouslyResponse(t *testing.T) {

	proxy.StartWithOutLambda(NewmockCompleteSynchronouslyHandler())

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	readRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/read.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/delete.request.json")
	listRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/list.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type fields struct {
		in proxy.ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		fields                             fields
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"complete synchronously CREATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *createRequest, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously READ response", fields{proxy.ProcessInvocationInput{mockContext{}, *readRequest, metric.New(New(), readRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously LIST response", fields{proxy.ProcessInvocationInput{mockContext{}, *listRequest, metric.New(New(), listRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(tt.fields.in)

			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation()
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
		})
	}
}

func Test_processMalformedSynchronouslyResponse(t *testing.T) {

	proxy.StartWithOutLambda(NewmockCompleteSynchronouslyHandler())

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/malformed.request.json")
	readRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/malformed.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/malformed.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/malformed.request.json")
	listRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/malformed.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type fields struct {
		in proxy.ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		fields                             fields
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"complete synchronously CREATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *createRequest, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously READ response", fields{proxy.ProcessInvocationInput{mockContext{}, *readRequest, metric.New(New(), readRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously LIST response", fields{proxy.ProcessInvocationInput{mockContext{}, *listRequest, metric.New(New(), listRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(tt.fields.in)

			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation()
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
		})
	}
}

func Test_processInvocationInProgressWithContextResponse(t *testing.T) {

	proxy.StartWithOutLambda(NewmockInProgressHandler())

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.with-request-context.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.with-request-context.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/delete.with-request-context.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type fields struct {
		in proxy.ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		fields                             fields
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"in progress CREATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *createRequest, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.InProgress,
			CallbackDelayMinutes: 5,
		}, false, 0, 1, 1, 1, 1},

		{"in progress UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.InProgress,
			CallbackDelayMinutes: 5,
		}, false, 0, 1, 1, 1, 1},

		{"in progress DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			ProgressStatus:       proxy.InProgress,
			CallbackDelayMinutes: 5,
			ResourceModel:        deleteRequest.Data.ResourceProperties,
		}, false, 0, 1, 1, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(tt.fields.in)

			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation()
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
		})
	}
}
