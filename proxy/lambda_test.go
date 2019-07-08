package proxy

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy/internal/scheduler"
)

const succeed = "\u2713"
const failed = "\u2717"

//Helper function to load the request data from file.
func loadData(theRequest *HandlerRequest, path string) *HandlerRequest {

	dat, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(dat, &theRequest); err != nil {

		panic(err)
	}

	return theRequest

}

//Test lambda handler for invalid request. All request should cause a panic and return a fail response.
func TestHandleLambdaInvalidEvent(t *testing.T) {

	var f tFunc = func(cb json.RawMessage) *ProgressEvent {
		return nil
	}

	re := NewMockResourceHandler(f)
	StartWithOutLambda(re, metric.New(NewMockedMetrics(), "AWS::Test::TestModel"), scheduler.New(NewmockedEvents()), nil)

	emptyRequest := loadData(&HandlerRequest{}, "tests/data/errorRequest/empty.requst.json")
	emptyresponseEndpoint := loadData(&HandlerRequest{}, "tests/data/errorRequest/request.no.responseEndpoint.json")
	emptyPlatformCreds := loadData(&HandlerRequest{}, "tests/data/errorRequest/request.no.platformCreds.json")
	emptyBearToken := loadData(&HandlerRequest{}, "tests/data/errorRequest/request.no.bearToken.json")
	emptyRegion := loadData(&HandlerRequest{}, "tests/data/errorRequest/request.no.region.json")
	emptyResourceProperties := loadData(&HandlerRequest{}, "tests/data/errorRequest/create.request.no.props.json")
	type args struct {
		ctx   context.Context
		event HandlerRequest
	}
	tests := []struct {
		name    string
		args    args
		want    HandlerResponse
		wantErr bool
	}{
		{"Empty request", args{context.Background(), *emptyRequest}, HandlerResponse{
			Message:         "No request object received",
			OperationStatus: Failed,
			BearerToken:     "",
			ErrorCode:       InvalidRequest,
			ResourceModel:   emptyRequest.Data.ResourceProperties,
		}, false},

		{"Empty responseEndpoint", args{context.Background(), *emptyresponseEndpoint}, HandlerResponse{
			Message:         "No callback endpoint received",
			OperationStatus: Failed,
			BearerToken:     "123456",
			ErrorCode:       InvalidRequest,
			ResourceModel:   emptyresponseEndpoint.Data.ResourceProperties,
		}, false},
		{"Empty platform creds", args{context.Background(), *emptyPlatformCreds}, HandlerResponse{
			Message:         "Missing required platform credentials",
			OperationStatus: Failed,
			BearerToken:     "123456",
			ErrorCode:       InvalidRequest,
			ResourceModel:   emptyPlatformCreds.Data.ResourceProperties,
		}, false},
		{"Empty bear Token", args{context.Background(), *emptyBearToken}, HandlerResponse{
			Message:         "No BearerToken received",
			OperationStatus: Failed,
			BearerToken:     "",
			ErrorCode:       InvalidRequest,
			ResourceModel:   emptyBearToken.Data.ResourceProperties,
		}, false},
		{"Empty region request", args{context.Background(), *emptyRegion}, HandlerResponse{
			Message:         "Region was not provided.",
			OperationStatus: Failed,
			BearerToken:     "123456",
			ErrorCode:       InvalidRequest,
			ResourceModel:   emptyRegion.Data.ResourceProperties,
		}, false},
		{"Empty resource resource properties", args{context.Background(), *emptyResourceProperties}, HandlerResponse{
			Message:         "Invalid resource properties object received",
			OperationStatus: Failed,
			BearerToken:     "123456",
			ErrorCode:       InvalidRequest,
			ResourceModel:   emptyResourceProperties.Data.ResourceProperties,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HandleLambdaEvent(tt.args.ctx, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Logf("\t%s\tHandleLambdaEvent() Should return %v", succeed, tt.wantErr)
				return
			}
			if reflect.DeepEqual(got, tt.want) {
				t.Logf("\t%s\tHandleLambdaEvent() Should return %v", succeed, tt.want)
			} else {
				t.Errorf("\t%s\tHandleLambdaEvent() Should return %v got : %v", failed, got, tt.want)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandleLambdaEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

//Test lambda handler for timeout. All custom handlers shoud return within (60) or should cause a panic and return a fail response.
func TestCustomHandlerProcessInvocationFailedToRespond(t *testing.T) {

	var f tFunc = func(cb json.RawMessage) *ProgressEvent {
		time.Sleep(65 * time.Second)
		return nil
	}

	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")

	re := NewMockResourceHandler(f)
	StartWithOutLambda(re, metric.New(NewMockedMetrics(), createRequest.ResourceType), scheduler.New(NewmockedEvents()), nil)

	type args struct {
		cx  context.Context
		req HandlerRequest
	}
	tests := []struct {
		name                               string
		args                               args
		want                               HandlerResponse
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"Failed to reply", args{context.Background(), *createRequest}, HandlerResponse{
			Message:         "Handler failed to provide a response",
			OperationStatus: Failed,
			BearerToken:     "123456",
			ErrorCode:       InvalidRequest,
			ResourceModel:   nil,
		},
			false, 1, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m := metpub.Client.(*MockedMetrics)
			e := sch.Client.(*MockedEvents)
			got, err := HandleLambdaEvent(tt.args.cx, tt.args.req)

			if (err != nil) != tt.wantErr {
				t.Logf("\t%s\tHandleLambdaEvent() Should return %v", succeed, tt.wantErr)
				return
			}

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

//Test lambda handler for synchronous.
func TestCustomHandlerProcessInvocationsynchronous(t *testing.T) {

	type testcallBack struct {
		Count int
	}
	var f tFunc = func(cb json.RawMessage) *ProgressEvent {

		tc := testcallBack{}

		if len(cb) > 0 {
			if err := json.Unmarshal(cb, &tc); err != nil {
				panic(err)
			}
		}
		tc.Count = tc.Count + 1

		return &ProgressEvent{
			OperationStatus:      Complete,
			Message:              "Complete",
			CallbackDelaySeconds: 0,
			CallbackContext:      tc,
		}
	}

	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")
	deleteRequest := loadData(&HandlerRequest{}, "tests/data/delete.request.json")
	listRequest := loadData(&HandlerRequest{}, "tests/data/list.request.json")
	readRequest := loadData(&HandlerRequest{}, "tests/data/read.request.json")
	updateRequest := loadData(&HandlerRequest{}, "tests/data/update.request.json")
	malformedRequest := loadData(&HandlerRequest{}, "tests/data/errorRequest/malformed.request.json")

	re := NewMockResourceHandler(f)
	StartWithOutLambda(re, metric.New(NewMockedMetrics(), createRequest.ResourceType), scheduler.New(NewmockedEvents()), nil)

	type args struct {
		cx  context.Context
		req HandlerRequest
	}
	tests := []struct {
		name                               string
		args                               args
		want                               HandlerResponse
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"CREATE synchronous call", args{mockContext{}, *createRequest}, HandlerResponse{
			Message:         "Complete",
			OperationStatus: Complete,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
		{"DELETE synchronous call", args{mockContext{}, *deleteRequest}, HandlerResponse{
			Message:         "Complete",
			OperationStatus: Complete,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
		{"LIST synchronous call", args{mockContext{}, *listRequest}, HandlerResponse{
			Message:         "Complete",
			OperationStatus: Complete,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
		{"READ synchronous call", args{mockContext{}, *readRequest}, HandlerResponse{
			Message:         "Complete",
			OperationStatus: Complete,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
		{"UPDATE asynchronous call", args{mockContext{}, *updateRequest}, HandlerResponse{
			Message:         "Complete",
			OperationStatus: Complete,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 1, 1, 0, 0},
		{"Malformed request", args{context.Background(), *malformedRequest}, HandlerResponse{
			Message:         "Complete",
			OperationStatus: Complete,
			BearerToken:     "123456",
			ErrorCode:       "",
		}, false, 0, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m := metpub.Client.(*MockedMetrics)
			e := sch.Client.(*MockedEvents)
			got, err := HandleLambdaEvent(tt.args.cx, tt.args.req)
			t.Log(got.Message)

			if (err != nil) != tt.wantErr {
				t.Logf("\t%s\tHandleLambdaEvent() Should return %v", succeed, tt.wantErr)
				return
			}

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

			if e.CleanupCloudWatchEventsCount == tt.wantcleanupCloudWatchEvents {
				t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
			} else {
				t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
			}
		})
	}
}

//Test lambda handler for short asynchronous.
func TestCustomHandlerProcessInvocationShortAsynchronous(t *testing.T) {

	type testcallBack struct {
		Count int
	}
	var f tFunc = func(cb json.RawMessage) *ProgressEvent {

		tc := testcallBack{}

		if len(cb) > 0 {
			if err := json.Unmarshal(cb, &tc); err != nil {
				panic(err)
			}
		}
		tc.Count = tc.Count + 1

		if tc.Count < 3 {

			return &ProgressEvent{
				OperationStatus:      InProgress,
				Message:              "InProgress",
				CallbackDelaySeconds: 5,
				CallbackContext:      tc,
			}

		}

		return &ProgressEvent{
			OperationStatus:      Complete,
			Message:              "Complete",
			CallbackDelaySeconds: 5,
			CallbackContext:      tc,
		}
	}

	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")
	deleteRequest := loadData(&HandlerRequest{}, "tests/data/delete.request.json")
	listRequest := loadData(&HandlerRequest{}, "tests/data/list.request.json")
	readRequest := loadData(&HandlerRequest{}, "tests/data/read.request.json")
	updateRequest := loadData(&HandlerRequest{}, "tests/data/update.request.json")

	re := NewMockResourceHandler(f)
	StartWithOutLambda(re, metric.New(NewMockedMetrics(), createRequest.ResourceType), scheduler.New(NewmockedEvents()), nil)

	type args struct {
		cx  context.Context
		req HandlerRequest
	}
	tests := []struct {
		name                               string
		args                               args
		want                               HandlerResponse
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"Short CREATE asynchronous call", args{mockContext{}, *createRequest}, HandlerResponse{
			Message:         "Complete,",
			OperationStatus: Complete,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
		{"Short DELETE asynchronous call", args{mockContext{}, *deleteRequest}, HandlerResponse{
			Message:         "Complete,",
			OperationStatus: Complete,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
		{"Short LIST asynchronous call", args{mockContext{}, *listRequest}, HandlerResponse{
			Message:         "Complete,",
			OperationStatus: Complete,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
		{"Short READ asynchronous call", args{mockContext{}, *readRequest}, HandlerResponse{
			Message:         "Complete,",
			OperationStatus: Complete,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
		{"Short UPDATE asynchronous call", args{mockContext{}, *updateRequest}, HandlerResponse{
			Message:         "Complete,",
			OperationStatus: Complete,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m := metpub.Client.(*MockedMetrics)
			e := sch.Client.(*MockedEvents)
			got, err := HandleLambdaEvent(tt.args.cx, tt.args.req)
			t.Log(got.Message)

			if (err != nil) != tt.wantErr {
				t.Logf("\t%s\tHandleLambdaEvent() Should return %v", succeed, tt.wantErr)
				return
			}

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

//Test lambda handler for long asynchronous.
func TestCustomHandlerProcessInvocationLongAsynchronous(t *testing.T) {

	type testcallBack struct {
		Count int
	}
	var f tFunc = func(cb json.RawMessage) *ProgressEvent {

		tc := testcallBack{}

		if len(cb) > 0 {
			if err := json.Unmarshal(cb, &tc); err != nil {
				panic(err)
			}
		}
		tc.Count = tc.Count + 1

		if tc.Count < 3 {

			return &ProgressEvent{
				OperationStatus:      InProgress,
				Message:              "InProgress",
				CallbackDelaySeconds: 160,
				CallbackContext:      tc,
			}

		}

		return &ProgressEvent{
			OperationStatus:      Complete,
			Message:              "Complete",
			CallbackDelaySeconds: 5,
			CallbackContext:      tc,
		}
	}

	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")
	deleteRequest := loadData(&HandlerRequest{}, "tests/data/delete.request.json")
	listRequest := loadData(&HandlerRequest{}, "tests/data/list.request.json")
	readRequest := loadData(&HandlerRequest{}, "tests/data/read.request.json")
	updateRequest := loadData(&HandlerRequest{}, "tests/data/update.request.json")

	re := NewMockResourceHandler(f)
	StartWithOutLambda(re, metric.New(NewMockedMetrics(), createRequest.ResourceType), scheduler.New(NewmockedEvents()), nil)

	type args struct {
		cx  context.Context
		req HandlerRequest
	}
	tests := []struct {
		name                               string
		args                               args
		want                               HandlerResponse
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"Long CREATE asynchronous call", args{mockContext{}, *createRequest}, HandlerResponse{
			Message:         "InProgress,",
			OperationStatus: InProgress,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
		{"Long DELETE asynchronous call", args{mockContext{}, *deleteRequest}, HandlerResponse{
			Message:         "InProgress,",
			OperationStatus: InProgress,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
		{"Long LIST asynchronous call", args{mockContext{}, *listRequest}, HandlerResponse{
			Message:         "InProgress,",
			OperationStatus: InProgress,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
		{"Long READ asynchronous call", args{mockContext{}, *readRequest}, HandlerResponse{
			Message:         "InProgress,",
			OperationStatus: InProgress,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
		{"Long UPDATE asynchronous call", args{mockContext{}, *updateRequest}, HandlerResponse{
			Message:         "InProgress,",
			OperationStatus: InProgress,
			BearerToken:     "123456",
			ErrorCode:       "",
			ResourceModel:   nil,
		},
			false, 0, 3, 3, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m := metpub.Client.(*MockedMetrics)
			e := sch.Client.(*MockedEvents)
			got, err := HandleLambdaEvent(tt.args.cx, tt.args.req)
			t.Log(got.Message)

			if (err != nil) != tt.wantErr {
				t.Logf("\t%s\tHandleLambdaEvent() Should return %v", succeed, tt.wantErr)
				return
			}

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

			if e.CleanupCloudWatchEventsCount == tt.wantcleanupCloudWatchEvents {
				t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
			} else {
				t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
			}
		})
	}
}

/*
func TestTransform(t *testing.T) {

	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")
	updateRequest := loadData(&HandlerRequest{}, "tests/data/update.request.json")


	type args struct {
		r       HandlerRequest
		handler *CustomHandler
	}
	tests := []struct {
		name         string
		args         args
		want         *ResourceHandlerRequest
		wantResource *MockHandler
		wantErr      bool
	}{

		{"Transform CREATE response", args{*createRequest, New(NewMock(nil))}, &ResourceHandlerRequest{
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

		{"Transform UPDATE response", args{*updateRequest, proxy.New(NewMock(nil))}, &ResourceHandlerRequest{
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
			got := Transform(tt.args.r, tt.args.handler)
			r := tt.args.handler.CustomResource.(*MockHandler)

			//if (err != nil) != tt.wantErr {
			//	t.Errorf("Transform() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}

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

	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")
	updateRequest := loadData(&HandlerRequest{}, "tests/data/update.request.json")


	type args struct {
		r       HandlerRequest
		handler *CustomHandler
	}
	tests := []struct {
		name         string
		args         args
		want         *ResourceHandlerRequest
		wantResource *MockHandlerNoDesired
		wantErr      bool
	}{

		{"Transform CREATE response", args{*createRequest, New(NewMockNoDesired(nil, nil))}, &ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandlerNoDesired{},

			false},

		{"Transform UPDATE response", args{*updateRequest, New(NewMockNoDesired(nil, nil))}, &ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandlerNoDesired{
				MockHandlerResourceNoDesired{},
				nil,
			},

			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Transform(tt.args.r, tt.args.handler)

			//if err.Error() == "Unable to find DesiredResource in Config object" {
			//	t.Logf("\t%s\tShould receive a %s error.", succeed, "Unable to find DesiredResource in Config object")
			//} else {
			//	t.Errorf("\t%s\tShould receive a %s error.", failed, "Unable to find DesiredResource in Config object")
			//}

		})
	}
}

func TestTransformNoPre(t *testing.T) {

	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")
	updateRequest := loadData(&HandlerRequest{}, "tests/data/update.request.json")

	type args struct {
		r       HandlerRequest
		handler *CustomHandler
	}
	tests := []struct {
		name         string
		args         args
		want         *ResourceHandlerRequest
		wantResource *MockHandlerNoPre
		wantErr      bool
	}{

		{"Transform CREATE response", args{*createRequest, New(NewMockNoPre(nil, nil))}, &ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandlerNoPre{},

			false},

		{"Transform UPDATE response", args{*updateRequest, New(NewMockNoPre(nil, nil))}, &ResourceHandlerRequest{
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
			Transform(tt.args.r, tt.args.handler)

			//if err.Error() == "Unable to find PreviousResource in Config object" {
			//	t.Logf("\t%s\tShould receive a %s error.", succeed, "Unable to find PreviousResource in Config object")
			//} else {
			//	t.Errorf("\t%s\tShould receive a %s error.", failed, "Unable to find PreviousResource in Config object")
			//}

		})
	}

}
*/
