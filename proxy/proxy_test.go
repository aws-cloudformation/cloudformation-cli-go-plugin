package proxy_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/scheduler"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/request"
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

type MockHandlerResource struct {
	Property1 string `json:"property1"`
	Property2 int    `json:"property2"`
}

type MockHandler struct {
	DesiredResourceState  MockHandlerResource
	PreviousResourceState MockHandlerResource
	ReturnState           *proxy.ProgressEvent
	ReturnError           error
}

func NewMock(state *proxy.ProgressEvent, stateError error) *MockHandler {

	h := MockHandler{
		ReturnState: state,
		ReturnError: stateError,
	}
	return &h

}

func (m *MockHandler) CreateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

func (m *MockHandler) DeleteRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

func (m *MockHandler) ListRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

func (m *MockHandler) ReadRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

func (m *MockHandler) UpdateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

type MockHandlerResourceNoDesired struct {
	Property1 string `json:"property1"`
	Property2 int    `json:"property2"`
}

type MockHandlerNoDesired struct {
	PreviousResourceState MockHandlerResourceNoDesired
	ReturnState           *proxy.ProgressEvent
	ReturnError           error
}

func NewMockNoDesired(state *proxy.ProgressEvent, stateError error) *MockHandlerNoDesired {

	h := MockHandlerNoDesired{
		ReturnState: state,
		ReturnError: stateError,
	}
	return &h

}

func (m *MockHandlerNoDesired) CreateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

func (m *MockHandlerNoDesired) DeleteRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

func (m *MockHandlerNoDesired) ListRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

func (m *MockHandlerNoDesired) ReadRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

func (m *MockHandlerNoDesired) UpdateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

type MockHandlerResourceNoPre struct {
	Property1 string `json:"property1"`
	Property2 int    `json:"property2"`
}

type MockHandlerNoPre struct {
	DesiredResourceState MockHandlerResource
	ReturnState          *proxy.ProgressEvent
	ReturnError          error
}

func NewMockNoPre(state *proxy.ProgressEvent, stateError error) *MockHandlerNoPre {

	h := MockHandlerNoPre{
		ReturnState: state,
		ReturnError: stateError,
	}
	return &h

}

func (m *MockHandlerNoPre) CreateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

func (m *MockHandlerNoPre) DeleteRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

func (m *MockHandlerNoPre) ListRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

func (m *MockHandlerNoPre) ReadRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
}

func (m *MockHandlerNoPre) UpdateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return m.ReturnState, m.ReturnError
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

func Test_processInvocationNoProps(t *testing.T) {
	re := NewMock(nil, nil)
	proxy.StartWithOutLambda(re)
	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.no.props.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.no.props.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/delete.request.no.props.json")

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
			OperationStatus:  proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},
		{"nil DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:  proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},

		{"nil UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:  proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(re)

			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation(&tt.fields.in)
			t.Log(got)

			if got.OperationStatus == tt.want.OperationStatus {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.OperationStatus)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.OperationStatus, got.OperationStatus)
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

func Test_processInvocationNullResponse(t *testing.T) {
	re := NewMock(nil, nil)
	proxy.StartWithOutLambda(re)
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
			OperationStatus:  proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},
		{"nil DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:  proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},
		{"nil LIST response", fields{proxy.ProcessInvocationInput{mockContext{}, *listRequest, metric.New(New(), listRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:  proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},
		{"nil READ response", fields{proxy.ProcessInvocationInput{mockContext{}, *readRequest, metric.New(New(), readRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:  proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},
		{"nil UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:  proxy.Failed,
			HandlerErrorCode: proxy.InvalidRequest,
			Message:          "Handler failed to provide a response",
		},
			true, 1, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(re)

			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation(&tt.fields.in)
			t.Log(got)

			if got.OperationStatus == tt.want.OperationStatus {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.OperationStatus)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.OperationStatus, got.OperationStatus)
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
	re := NewMock(&proxy.ProgressEvent{
		OperationStatus:      proxy.Failed,
		HandlerErrorCode:     "Custom Fault",
		Message:              "Custom Fault",
		CallbackDelayMinutes: 0,
	}, nil)
	proxy.StartWithOutLambda(re)

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
			OperationStatus:      proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
		{"failed DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
		{"failed LIST response", fields{proxy.ProcessInvocationInput{mockContext{}, *listRequest, metric.New(New(), listRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
		{"failed READ response", fields{proxy.ProcessInvocationInput{mockContext{}, *readRequest, metric.New(New(), readRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
		{"failed UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(re)
			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation(&tt.fields.in)
			if got.OperationStatus == tt.want.OperationStatus {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.OperationStatus)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.OperationStatus, got.OperationStatus)
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

	re := NewMock(&proxy.ProgressEvent{
		OperationStatus:      proxy.Complete,
		CallbackDelayMinutes: 0,
	}, nil)
	proxy.StartWithOutLambda(re)

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
			OperationStatus:      proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously READ response", fields{proxy.ProcessInvocationInput{mockContext{}, *readRequest, metric.New(New(), readRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously LIST response", fields{proxy.ProcessInvocationInput{mockContext{}, *listRequest, metric.New(New(), listRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(re)
			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation(&tt.fields.in)
			if got.OperationStatus == tt.want.OperationStatus {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.OperationStatus)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.OperationStatus, got.OperationStatus)
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

	re := NewMock(&proxy.ProgressEvent{
		OperationStatus:      proxy.Complete,
		CallbackDelayMinutes: 0,
	}, nil)
	proxy.StartWithOutLambda(re)

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
			OperationStatus:      proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously READ response", fields{proxy.ProcessInvocationInput{mockContext{}, *readRequest, metric.New(New(), readRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously LIST response", fields{proxy.ProcessInvocationInput{mockContext{}, *listRequest, metric.New(New(), listRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(re)
			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation(&tt.fields.in)
			if got.OperationStatus == tt.want.OperationStatus {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.OperationStatus)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.OperationStatus, got.OperationStatus)
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

	re := NewMock(&proxy.ProgressEvent{
		OperationStatus:      proxy.InProgress,
		CallbackDelayMinutes: 5,
	}, nil)
	proxy.StartWithOutLambda(re)

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
			OperationStatus:      proxy.InProgress,
			CallbackDelayMinutes: 5,
		}, false, 0, 1, 1, 1, 1},

		{"in progress UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.InProgress,
			CallbackDelayMinutes: 5,
		}, false, 0, 1, 1, 1, 1},

		{"in progress DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents())}}, &proxy.ProgressEvent{
			OperationStatus:      proxy.InProgress,
			CallbackDelayMinutes: 5,
			ResourceModel:        deleteRequest.Data.ResourceProperties,
		}, false, 0, 1, 1, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(re)
			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation(&tt.fields.in)
			if got.OperationStatus == tt.want.OperationStatus {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.OperationStatus)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.OperationStatus, got.OperationStatus)
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

func TestTransform(t *testing.T) {

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type args struct {
		r       proxy.HandlerRequest
		handler *proxy.CustomHandler
	}
	tests := []struct {
		name         string
		args         args
		want         *proxy.ResourceHandlerRequest
		wantResource *MockHandler
		wantErr      bool
	}{

		{"Transform CREATE response", args{*createRequest, proxy.New(NewMock(nil, nil))}, &proxy.ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandler{
				MockHandlerResource{"abc", 123},
				MockHandlerResource{},
				nil,
				nil,
			},

			false},

		{"Transform UPDATE response", args{*updateRequest, proxy.New(NewMock(nil, nil))}, &proxy.ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandler{
				MockHandlerResource{"abc", 123},
				MockHandlerResource{"cba", 321},
				nil,
				nil,
			},

			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := proxy.Transform(tt.args.r, tt.args.handler)
			r := tt.args.handler.CustomResource.(*MockHandler)

			if (err != nil) != tt.wantErr {
				t.Errorf("Transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if reflect.DeepEqual(got, tt.want) {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want, got)
			}

			if reflect.DeepEqual(got, tt.want) {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want, got)
			}

			if reflect.DeepEqual(r, tt.wantResource) {
				t.Logf("\t%s\tShould update resource.", succeed)
			} else {
				t.Errorf("\t%s\tShould update resource %v was : %v", failed, tt.wantResource, r)
			}

		})
	}
}

func TestTransformNoDesired(t *testing.T) {

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type args struct {
		r       proxy.HandlerRequest
		handler *proxy.CustomHandler
	}
	tests := []struct {
		name         string
		args         args
		want         *proxy.ResourceHandlerRequest
		wantResource *MockHandlerNoDesired
		wantErr      bool
	}{

		{"Transform CREATE response", args{*createRequest, proxy.New(NewMockNoDesired(nil, nil))}, &proxy.ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandlerNoDesired{},

			false},

		{"Transform UPDATE response", args{*updateRequest, proxy.New(NewMockNoDesired(nil, nil))}, &proxy.ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandlerNoDesired{
				MockHandlerResourceNoDesired{},
				nil,
				nil,
			},

			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := proxy.Transform(tt.args.r, tt.args.handler)

			if err.Error() == "Unable to find DesiredResource in Config object" {
				t.Logf("\t%s\tShould receive a %s error.", succeed, "Unable to find DesiredResource in Config object")
			} else {
				t.Errorf("\t%s\tShould receive a %s error.", failed, "Unable to find DesiredResource in Config object")
			}

		})
	}
}

func TestTransformNoPre(t *testing.T) {

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type args struct {
		r       proxy.HandlerRequest
		handler *proxy.CustomHandler
	}
	tests := []struct {
		name         string
		args         args
		want         *proxy.ResourceHandlerRequest
		wantResource *MockHandlerNoPre
		wantErr      bool
	}{

		{"Transform CREATE response", args{*createRequest, proxy.New(NewMockNoPre(nil, nil))}, &proxy.ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandlerNoPre{},

			false},

		{"Transform UPDATE response", args{*updateRequest, proxy.New(NewMockNoPre(nil, nil))}, &proxy.ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandlerNoPre{},

			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := proxy.Transform(tt.args.r, tt.args.handler)

			if err.Error() == "Unable to find PreviousResource in Config object" {
				t.Logf("\t%s\tShould receive a %s error.", succeed, "Unable to find PreviousResource in Config object")
			} else {
				t.Errorf("\t%s\tShould receive a %s error.", failed, "Unable to find PreviousResource in Config object")
			}

		})
	}
}

func TestInjectCredentialsAndInvoke(t *testing.T) {
	type args struct {
		req request.Request
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InjectCredentialsAndInvoke(tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("InjectCredentialsAndInvoke() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
