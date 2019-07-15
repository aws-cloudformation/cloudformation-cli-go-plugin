package proxy

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

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
func TestInvokeHandlerinvalidRequest(t *testing.T) {
	var empty interface{}

	emptyPayload := loadData(&HandlerRequest{}, "tests/data/empty.request.json")
	emptyResourceProperties := loadData(&HandlerRequest{}, "tests/data/empty.resource.request.json")
	//updateRequest := loadData(&HandlerRequest{}, "tests/data/update.request.json")
	//deleteRequest := loadData(&HandlerRequest{}, "tests/data/delete.request.json")
	//listRequest := loadData(&HandlerRequest{}, "tests/data/list.request.json")
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
		{"EmptyPayload: ReturnsFailure", args{context.Background(), *emptyPayload}, HandlerResponse{
			Message:         "Invalid request object received",
			OperationStatus: FAILED,
			ErrorCode:       "InternalFailure",
			ResourceModel:   empty,
		}, false, 1, 0, 0, 0, 0},

		{"EmptyResourceProperties: ReturnsFailure", args{context.Background(), *emptyResourceProperties}, HandlerResponse{
			Message:         "Invalid resource properties object received",
			OperationStatus: FAILED,
			BearerToken:     "123456",
			ErrorCode:       "InternalFailure",
			ResourceModel:   empty,
		}, false, 1, 0, 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var f tFunc = func(cb json.RawMessage) (*ProgressEvent, error) {
				return nil, nil
			}

			re := NewMockResourceHandler(f)

			p := Proxy{
				customResource: re,
				sch:            scheduler.New(NewmockedEvents()),
				metpub:         metric.New(NewMockedMetrics(), emptyResourceProperties.ResourceType),
				cbak:           nil,
			}

			m := p.metpub.Client.(*MockedMetrics)
			e := p.sch.Client.(*MockedEvents)
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

			if reflect.DeepEqual(got, tt.want) {
				t.Logf("\t%s\tHandler Response should match.", succeed)
			} else {
				t.Errorf("\t%s\tHandler Response should match %v : %v", failed, tt.want, got)
			}

		})
	}

}
