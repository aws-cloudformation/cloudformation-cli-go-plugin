package cfn

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/errs"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/internal/metric"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/internal/scheduler"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/proxy"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

const succeed = "\u2713"
const failed = "\u2717"

//loadData is a helper function that creates a HandlerRequest form a  data from file.
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

//Test lambda handler for invalid request. All request should return a fail response.
func TestInvokeHandlerinvalidRequestReturnFailure(t *testing.T) {
	var empty interface{}

	emptyPayload := loadData(&HandlerRequest{}, "tests/data/empty.request.json")
	emptyResourceProperties := loadData(&HandlerRequest{}, "tests/data/empty.resource.request.json")
	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")
	noEndPointRequest := loadData(&HandlerRequest{}, "tests/data/no-response-endpoint.request.json")
	withoutPlatformCredentialsRequest := loadData(&HandlerRequest{}, "tests/data/create.request-without-platform-credentials.json")
	type args struct {
		ctx     context.Context
		event   HandlerRequest
		pr      *proxy.ProgressEvent
		prError error
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
		{"EmptyPayload: ReturnsFailure", args{context.Background(), *emptyPayload, nil, nil}, HandlerResponse{
			Message:         "Invalid request object received",
			OperationStatus: proxy.FAILED,
			ErrorCode:       "InternalFailure",
			ResourceModel:   empty,
		}, false, 1, 0, 0, 0, 0},

		{"EmptyResourceProperties: ReturnsFailure", args{context.Background(), *emptyResourceProperties, nil, nil}, HandlerResponse{
			Message:         "Invalid resource properties object received",
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			ErrorCode:       "InternalFailure",
			ResourceModel:   empty,
		}, false, 1, 0, 0, 0, 0},

		{"Create Resource: Returns ResourceNotFound Error", args{context.Background(), *createRequest, nil, &errs.ResourceNotFoundError{ResourceTypeName: "AWS::Test::TestModel", ResourceIdentifier: "id-1234"}}, HandlerResponse{
			Message:         "Resource of type 'AWS::Test::TestModel' with identifier 'id-1234' was not found.",
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			ErrorCode:       proxy.NotFound,
			ResourceModel:   empty,
		}, false, 1, 1, 1, 0, 0},

		{"Create Resource: Returns AlreadyExists Error", args{context.Background(), *createRequest, nil, &errs.ResourceAlreadyExistsError{ResourceTypeName: "AWS::Test::TestModel", ResourceIdentifier: "id-1234"}}, HandlerResponse{
			Message:         "Resource of type 'AWS::Test::TestModel' with identifier 'id-1234' already exists.",
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			ErrorCode:       proxy.AlreadyExists,
			ResourceModel:   empty,
		}, false, 1, 1, 1, 0, 0},

		{"Create Resource: Returns AmazonService Error", args{context.Background(), *createRequest, nil, awserr.New("GeneralServiceException", "some error", nil)}, HandlerResponse{
			Message:         "GeneralServiceException: some error",
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			ErrorCode:       proxy.GeneralServiceException,
			ResourceModel:   empty,
		}, false, 1, 1, 1, 0, 0},

		{"Create Resource: Returns missing ResponseEndpoint Error", args{context.Background(), *noEndPointRequest, nil, nil}, HandlerResponse{
			Message:         "No callback endpoint received",
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			ErrorCode:       proxy.InternalFailure,
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 1, 0, 0, 0, 0},
		{"Create Resource: Returns Missing required platform credentials Error", args{context.Background(), *withoutPlatformCredentialsRequest, nil, nil}, HandlerResponse{
			Message:         "Missing required platform credentials",
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			ErrorCode:       proxy.InternalFailure,
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 1, 0, 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var f tFunc = func(resource mockCustomResource) (*proxy.ProgressEvent, error) {
				return tt.args.pr, tt.args.prError
			}

			re := NewMockResourceHandler(f)
			var buf bytes.Buffer

			e := NewMockedEvents()
			m := NewMockedMetrics()

			p := Wrapper{
				customResource: re,
				sch:            scheduler.New(NewMockCloudWatchEventsProvider(e)),
				metpub:         metric.New(NewMockCloudWatchProvider(m)),
				cbak:           nil,
				logger:         log.New(&buf, "INFO: ", log.Lshortfile),
			}

			p.sch.RefreshClient()
			p.metpub.RefreshClient()
			//p.cbak.RefreshClient()
			got, err := p.HandleLambdaEvent(tt.args.ctx, tt.args.event)

			if err != nil {
				panic(err)
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

			if got.OperationStatus == tt.want.OperationStatus {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.OperationStatus)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.OperationStatus, got.OperationStatus)
			}
			gotResult, _ := json.MarshalIndent(got, "", "  ")
			wantResult, _ := json.MarshalIndent(tt.want, "", "  ")
			if reflect.DeepEqual(gotResult, wantResult) {
				t.Logf("\t%s\tHandler Response should match.", string(succeed))
			} else {
				t.Errorf("\t%s\tHandler Response should match \nGot:\n%v\nWant:\n%v", failed, string(gotResult), string(wantResult))
			}

		})
	}

}

func TestCustomHandlerProcessInvocationsynchronousReturnsSuccess(t *testing.T) {

	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")
	deleteRequest := loadData(&HandlerRequest{}, "tests/data/delete.request.json")
	listRequest := loadData(&HandlerRequest{}, "tests/data/list.request.json")
	readRequest := loadData(&HandlerRequest{}, "tests/data/read.request.json")
	updateRequest := loadData(&HandlerRequest{}, "tests/data/update.request.json")

	type args struct {
		ctx   context.Context
		event HandlerRequest
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
		{"Create: Returns Success", args{mockContext{}, *createRequest}, HandlerResponse{
			OperationStatus: proxy.Complete,
			BearerToken:     "123456",
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"Delete: Returns Success", args{mockContext{}, *deleteRequest}, HandlerResponse{
			OperationStatus: proxy.Complete,
			BearerToken:     "123456",
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"List: Returns Success", args{mockContext{}, *listRequest}, HandlerResponse{
			OperationStatus: proxy.Complete,
			BearerToken:     "123456",
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"Read: Returns Success", args{mockContext{}, *readRequest}, HandlerResponse{
			OperationStatus: proxy.Complete,
			BearerToken:     "123456",
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"Update: Returns Success", args{mockContext{}, *updateRequest}, HandlerResponse{
			OperationStatus: proxy.Complete,
			BearerToken:     "123456",
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var f tFunc = func(resource mockCustomResource) (*proxy.ProgressEvent, error) {
				return &proxy.ProgressEvent{OperationStatus: proxy.Complete, ResourceModel: resource}, nil
			}

			re := NewMockResourceHandler(f)
			var buf bytes.Buffer

			e := NewMockedEvents()
			m := NewMockedMetrics()

			p := Wrapper{
				customResource: re,
				sch:            scheduler.New(NewMockCloudWatchEventsProvider(e)),
				metpub:         metric.New(NewMockCloudWatchProvider(m)),
				cbak:           nil,
				logger:         log.New(&buf, "INFO: ", log.Lshortfile),
			}

			p.sch.RefreshClient()
			p.metpub.RefreshClient()

			got, err := p.HandleLambdaEvent(tt.args.ctx, tt.args.event)

			if err != nil {
				panic(err)
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
				t.Logf("\t%s\tCleanupCloudWatchEventsCount metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
			} else {
				t.Errorf("\t%s\tCleanupCloudWatchEventsCount should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
			}

			if got.OperationStatus == tt.want.OperationStatus {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.OperationStatus)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.OperationStatus, got.OperationStatus)
			}

			gotResult, _ := json.MarshalIndent(got, "", "  ")
			wantResult, _ := json.MarshalIndent(tt.want, "", "  ")
			if reflect.DeepEqual(gotResult, wantResult) {
				t.Logf("\t%s\tHandler Response should match.", string(succeed))
			} else {
				t.Errorf("\t%s\tHandler Response should match \nGot:\n%v\nWant:\n%v", failed, string(gotResult), string(wantResult))
			}

		})
	}

}

func TestCustomHandlerProcessInvocationNoLambdaContextReturnsFailed(t *testing.T) {

	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")

	type args struct {
		ctx   context.Context
		event HandlerRequest
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
		{"Create: Returns Success", args{nil, *createRequest}, HandlerResponse{
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			ErrorCode:       proxy.InternalFailure,
			Message:         "Internal error",
		}, false, 1, 0, 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var f tFunc = func(resource mockCustomResource) (*proxy.ProgressEvent, error) {
				return &proxy.ProgressEvent{OperationStatus: proxy.InProgress, ResourceModel: resource, CallbackDelaySeconds: 120}, nil
			}

			re := NewMockResourceHandler(f)
			var buf bytes.Buffer

			e := NewMockedEvents()
			m := NewMockedMetrics()

			p := Wrapper{
				customResource: re,
				sch:            scheduler.New(NewMockCloudWatchEventsProvider(e)),
				metpub:         metric.New(NewMockCloudWatchProvider(m)),
				cbak:           nil,
				logger:         log.New(&buf, "INFO: ", log.Lshortfile),
			}

			p.sch.RefreshClient()
			p.metpub.RefreshClient()

			got, err := p.HandleLambdaEvent(tt.args.ctx, tt.args.event)

			if err != nil {
				panic(err)
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
				t.Logf("\t%s\tCleanupCloudWatchEventsCount metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
			} else {
				t.Errorf("\t%s\tCleanupCloudWatchEventsCount should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
			}

			if got.OperationStatus == tt.want.OperationStatus {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.OperationStatus)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.OperationStatus, got.OperationStatus)
			}

			gotResult, _ := json.MarshalIndent(got, "", "  ")
			wantResult, _ := json.MarshalIndent(tt.want, "", "  ")
			if reflect.DeepEqual(gotResult, wantResult) {
				t.Logf("\t%s\tHandler Response should match.", string(succeed))
			} else {
				t.Errorf("\t%s\tHandler Response should match \nGot:\n%v\nWant:\n%v", failed, string(gotResult), string(wantResult))
			}

		})
	}

}

func TestCustomHandlerProcessInvocationsynchronousReturnsInprogress(t *testing.T) {

	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")
	deleteRequest := loadData(&HandlerRequest{}, "tests/data/delete.request.json")
	listRequest := loadData(&HandlerRequest{}, "tests/data/list.request.json")
	readRequest := loadData(&HandlerRequest{}, "tests/data/read.request.json")
	updateRequest := loadData(&HandlerRequest{}, "tests/data/update.request.json")

	type args struct {
		ctx   context.Context
		event HandlerRequest
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
		{"Create: Returns Success", args{mockContext{}, *createRequest}, HandlerResponse{
			OperationStatus: proxy.InProgress,
			BearerToken:     "123456",
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 1, 0},

		{"Delete: Returns Success", args{mockContext{}, *deleteRequest}, HandlerResponse{
			OperationStatus: proxy.InProgress,
			BearerToken:     "123456",
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 1, 0},

		{"List: Returns Success", args{mockContext{}, *listRequest}, HandlerResponse{
			OperationStatus: proxy.InProgress,
			BearerToken:     "123456",
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 1, 0},

		{"Read: Returns Success", args{mockContext{}, *readRequest}, HandlerResponse{
			OperationStatus: proxy.InProgress,
			BearerToken:     "123456",
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 1, 0},

		{"Update: Returns Success", args{mockContext{}, *updateRequest}, HandlerResponse{
			OperationStatus: proxy.InProgress,
			BearerToken:     "123456",
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var f tFunc = func(resource mockCustomResource) (*proxy.ProgressEvent, error) {
				return &proxy.ProgressEvent{OperationStatus: proxy.InProgress, ResourceModel: resource, CallbackDelaySeconds: 120}, nil
			}

			re := NewMockResourceHandler(f)
			var buf bytes.Buffer

			e := NewMockedEvents()
			m := NewMockedMetrics()

			p := Wrapper{
				customResource: re,
				sch:            scheduler.New(NewMockCloudWatchEventsProvider(e)),
				metpub:         metric.New(NewMockCloudWatchProvider(m)),
				cbak:           nil,
				logger:         log.New(&buf, "INFO: ", log.Lshortfile),
			}

			p.sch.RefreshClient()
			p.metpub.RefreshClient()

			got, err := p.HandleLambdaEvent(tt.args.ctx, tt.args.event)

			if err != nil {
				panic(err)
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
				t.Logf("\t%s\tCleanupCloudWatchEventsCount metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
			} else {
				t.Errorf("\t%s\tCleanupCloudWatchEventsCount should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
			}

			if got.OperationStatus == tt.want.OperationStatus {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.OperationStatus)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.OperationStatus, got.OperationStatus)
			}

			gotResult, _ := json.MarshalIndent(got, "", "  ")
			wantResult, _ := json.MarshalIndent(tt.want, "", "  ")
			if reflect.DeepEqual(gotResult, wantResult) {
				t.Logf("\t%s\tHandler Response should match.", string(succeed))
			} else {
				t.Errorf("\t%s\tHandler Response should match \nGot:\n%v\nWant:\n%v", failed, string(gotResult), string(wantResult))
			}

		})
	}

}

func TestCustomHandlerProcessInvocationsynchronousReturnsFailure(t *testing.T) {

	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")
	deleteRequest := loadData(&HandlerRequest{}, "tests/data/delete.request.json")
	listRequest := loadData(&HandlerRequest{}, "tests/data/list.request.json")
	readRequest := loadData(&HandlerRequest{}, "tests/data/read.request.json")
	updateRequest := loadData(&HandlerRequest{}, "tests/data/update.request.json")

	type args struct {
		ctx   context.Context
		event HandlerRequest
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
		{"Create: Returns Failure", args{mockContext{}, *createRequest}, HandlerResponse{
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			Message:         "Custom Fault",
			ErrorCode:       proxy.InternalFailure,
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"Delete: Returns Failure", args{mockContext{}, *deleteRequest}, HandlerResponse{
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			Message:         "Custom Fault",
			ErrorCode:       proxy.InternalFailure,
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"List: Returns Failure", args{mockContext{}, *listRequest}, HandlerResponse{
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			Message:         "Custom Fault",
			ErrorCode:       proxy.InternalFailure,
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"Read: Returns Failure", args{mockContext{}, *readRequest}, HandlerResponse{
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			Message:         "Custom Fault",
			ErrorCode:       proxy.InternalFailure,
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},

		{"Update: Returns Failure", args{mockContext{}, *updateRequest}, HandlerResponse{
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			Message:         "Custom Fault",
			ErrorCode:       proxy.InternalFailure,
			ResourceModel:   mockCustomResource{Property1: "abc", Property2: 123},
		}, false, 0, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var f tFunc = func(resource mockCustomResource) (*proxy.ProgressEvent, error) {
				return &proxy.ProgressEvent{OperationStatus: proxy.FAILED, ResourceModel: resource, HandlerErrorCode: proxy.InternalFailure, Message: "Custom Fault"}, nil
			}

			re := NewMockResourceHandler(f)
			var buf bytes.Buffer

			e := NewMockedEvents()
			m := NewMockedMetrics()

			p := Wrapper{
				customResource: re,
				sch:            scheduler.New(NewMockCloudWatchEventsProvider(e)),
				metpub:         metric.New(NewMockCloudWatchProvider(m)),
				cbak:           nil,
				logger:         log.New(&buf, "INFO: ", log.Lshortfile),
			}

			p.sch.RefreshClient()
			p.metpub.RefreshClient()

			got, err := p.HandleLambdaEvent(tt.args.ctx, tt.args.event)

			if err != nil {
				panic(err)
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
				t.Logf("\t%s\tCleanupCloudWatchEventsCount metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
			} else {
				t.Errorf("\t%s\tCleanupCloudWatchEventsCount should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
			}

			if got.OperationStatus == tt.want.OperationStatus {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.OperationStatus)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.OperationStatus, got.OperationStatus)
			}

			gotResult, _ := json.MarshalIndent(got, "", "  ")
			wantResult, _ := json.MarshalIndent(tt.want, "", "  ")
			if reflect.DeepEqual(gotResult, wantResult) {
				t.Logf("\t%s\tHandler Response should match.", string(succeed))
			} else {
				t.Errorf("\t%s\tHandler Response should match \nGot:\n%v\nWant:\n%v", failed, string(gotResult), string(wantResult))
			}

		})
	}

}

func TestCustomHandlerProcessInvocatioNullResponseReturnsFailure(t *testing.T) {

	createRequest := loadData(&HandlerRequest{}, "tests/data/create.request.json")
	deleteRequest := loadData(&HandlerRequest{}, "tests/data/delete.request.json")
	listRequest := loadData(&HandlerRequest{}, "tests/data/list.request.json")
	readRequest := loadData(&HandlerRequest{}, "tests/data/read.request.json")
	updateRequest := loadData(&HandlerRequest{}, "tests/data/update.request.json")

	type args struct {
		ctx   context.Context
		event HandlerRequest
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
		{"Create: Returns Failure", args{mockContext{}, *createRequest}, HandlerResponse{
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			Message:         "Handler returned null",
			ErrorCode:       proxy.InternalFailure,
		}, false, 1, 1, 1, 0, 0},

		{"Delete: Returns Failure", args{mockContext{}, *deleteRequest}, HandlerResponse{
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			Message:         "Handler returned null",
			ErrorCode:       proxy.InternalFailure,
		}, false, 1, 1, 1, 0, 0},

		{"List: Returns Failure", args{mockContext{}, *listRequest}, HandlerResponse{
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			Message:         "Handler returned null",
			ErrorCode:       proxy.InternalFailure,
		}, false, 1, 1, 1, 0, 0},

		{"Read: Returns Failure", args{mockContext{}, *readRequest}, HandlerResponse{
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			Message:         "Handler returned null",
			ErrorCode:       proxy.InternalFailure,
		}, false, 1, 1, 1, 0, 0},

		{"Update: Returns Failure", args{mockContext{}, *updateRequest}, HandlerResponse{
			OperationStatus: proxy.FAILED,
			BearerToken:     "123456",
			Message:         "Handler returned null",
			ErrorCode:       proxy.InternalFailure,
		}, false, 1, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var f tFunc = func(resource mockCustomResource) (*proxy.ProgressEvent, error) {
				return nil, nil
			}

			re := NewMockResourceHandler(f)
			var buf bytes.Buffer

			e := NewMockedEvents()
			m := NewMockedMetrics()

			p := Wrapper{
				customResource: re,
				sch:            scheduler.New(NewMockCloudWatchEventsProvider(e)),
				metpub:         metric.New(NewMockCloudWatchProvider(m)),
				cbak:           nil,
				logger:         log.New(&buf, "INFO: ", log.Lshortfile),
			}

			p.sch.RefreshClient()
			p.metpub.RefreshClient()

			got, err := p.HandleLambdaEvent(tt.args.ctx, tt.args.event)

			if err != nil {
				panic(err)
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
				t.Logf("\t%s\tCleanupCloudWatchEventsCount metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
			} else {
				t.Errorf("\t%s\tCleanupCloudWatchEventsCount should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
			}

			if got.OperationStatus == tt.want.OperationStatus {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.OperationStatus)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.OperationStatus, got.OperationStatus)
			}

			gotResult, _ := json.MarshalIndent(got, "", "  ")
			wantResult, _ := json.MarshalIndent(tt.want, "", "  ")
			if reflect.DeepEqual(gotResult, wantResult) {
				t.Logf("\t%s\tHandler Response should match.", string(succeed))
			} else {
				t.Errorf("\t%s\tHandler Response should match \nGot:\n%v\nWant:\n%v", failed, string(gotResult), string(wantResult))
			}

		})
	}

}
