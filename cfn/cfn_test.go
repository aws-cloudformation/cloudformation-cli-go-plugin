package cfn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/scheduler"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func TestReschedule(t *testing.T) {
	c := context.Background()

	p := handler.NewProgressEvent()
	p.CallbackContext = map[string]interface{}{"foo": true}
	e := &event{}

	s := scheduler.ScheduleIDS{
		Target:  "foo",
		Handler: "bar",
	}

	type args struct {
		ctx             context.Context
		invokeScheduler InvokeScheduler
		progEvt         handler.ProgressEvent
		event           *event
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"Test reschedule should return true", args{c, MockScheduler{Err: nil, Result: &scheduler.Result{ComputeLocal: true, IDS: s}}, p, e}, true, false},
		{"Test reschedule should return false", args{c, MockScheduler{Err: nil, Result: &scheduler.Result{ComputeLocal: false, IDS: s}}, p, e}, false, false},
		{"Test reschedule should return error", args{c, MockScheduler{Err: errors.New("error"), Result: &scheduler.Result{ComputeLocal: true, IDS: s}}, p, e}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reschedule(tt.args.ctx, tt.args.invokeScheduler, tt.args.progEvt, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("reschedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("reschedule() = %v, want %v", got, tt.want)
			}
			if len(e.RequestContext.CloudWatchEventsRuleName) == 0 {
				t.Errorf("RequestContext.CloudWatchEventsRuleName not set")
			}
			if len(e.RequestContext.CloudWatchEventsTargetID) == 0 {
				t.Errorf("RequestContext.CloudWatchEventsTargetID not set")
			}
			if len(e.RequestContext.CallbackContext) == 0 {
				t.Errorf("RequestContext.CallbackContext not set")
			}
		})
	}
}

func TestMakeEventFunc(t *testing.T) {
	start := time.Now()
	future := start.Add(time.Minute * 15)

	tc, cancel := context.WithDeadline(context.Background(), future)

	defer cancel()

	lc := lambdacontext.NewContext(tc, &lambdacontext.LambdaContext{})

	f1 := func(callback map[string]interface{}, s *session.Session) handler.ProgressEvent {
		return handler.ProgressEvent{}
	}

	f2 := func(callback map[string]interface{}, s *session.Session) handler.ProgressEvent {
		return handler.ProgressEvent{
			OperationStatus:      handler.InProgress,
			Message:              "In Progress",
			CallbackDelaySeconds: 130,
		}
	}

	f4 := func(callback map[string]interface{}, s *session.Session) handler.ProgressEvent {
		return handler.ProgressEvent{
			OperationStatus: handler.Failed,
		}
	}

	f3 := func(callback map[string]interface{}, s *session.Session) handler.ProgressEvent {

		if len(callback) == 1 {
			return handler.ProgressEvent{
				OperationStatus: handler.Success,
				Message:         "Success",
			}

		}
		return handler.ProgressEvent{
			OperationStatus:      handler.InProgress,
			Message:              "In Progress",
			CallbackDelaySeconds: 3,
			CallbackContext:      map[string]interface{}{"foo": "bar"},
		}
	}

	f5 := func(callback map[string]interface{}, s *session.Session) handler.ProgressEvent {

		if len(callback) == 1 {
			return handler.ProgressEvent{
				OperationStatus: handler.Failed,
				Message:         "Failed",
			}

		}
		return handler.ProgressEvent{
			OperationStatus:      handler.InProgress,
			Message:              "In Progress",
			CallbackDelaySeconds: 3,
			CallbackContext:      map[string]interface{}{"foo": "bar"},
		}
	}

	f6 := func(callback map[string]interface{}, s *session.Session) (response handler.ProgressEvent) {
		defer func() {
			// Catch any panics and return a failed ProgressEvent
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = errors.New(fmt.Sprint(r))
				}

				response = handler.NewFailedEvent(err)
			}
		}()
		panic("error")
	}

	type args struct {
		h     Handler
		ctx   context.Context
		event *event
	}
	tests := []struct {
		name    string
		args    args
		want    response
		wantErr bool
	}{
		{"Test simple CREATE", args{&MockHandler{f1}, lc, loadEvent("request.create.json", &event{})}, response{
			BearerToken: "123456",
		}, false},
		{"Test CREATE failed", args{&MockHandler{f4}, lc, loadEvent("request.create.json", &event{})}, response{
			OperationStatus: handler.Failed,
			BearerToken:     "123456",
		}, false},
		{"Test simple CREATE async", args{&MockHandler{f2}, lc, loadEvent("request.create.json", &event{})}, response{
			BearerToken:     "123456",
			Message:         "In Progress",
			OperationStatus: handler.InProgress,
		}, false},
		{"Test CREATE async local", args{&MockHandler{f3}, lc, loadEvent("request.create.json", &event{})}, response{
			BearerToken:     "123456",
			Message:         "Success",
			OperationStatus: handler.Success,
		}, false},
		{"Test CREATE async local failed", args{&MockHandler{f5}, lc, loadEvent("request.create.json", &event{})}, response{
			BearerToken:     "123456",
			Message:         "Failed",
			OperationStatus: handler.Failed,
		}, false},
		{"Test READ async should return err", args{&MockHandler{f2}, lc, loadEvent("request.read.json", &event{})}, response{
			OperationStatus: handler.Failed,
		}, true},
		{"Test account number should not error", args{&MockHandler{f1}, context.Background(), loadEvent("request.read.invalid.validation.json", &event{})}, response{
			BearerToken: "123456",
		}, false},
		{"Test invalid Action", args{&MockHandler{f1}, context.Background(), loadEvent("request.invalid.json", &event{})}, response{
			OperationStatus: handler.Failed,
		}, true},
		{"Test wrap panic", args{&MockHandler{f6}, context.Background(), loadEvent("request.create.json", &event{})}, response{
			OperationStatus: handler.Failed,
			ErrorCode:       cloudformation.HandlerErrorCodeGeneralServiceException,
			Message:         "Unable to complete request: error",
			BearerToken:     "123456",
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := makeEventFunc(tt.args.h)

			got, err := f(tt.args.ctx, tt.args.event)

			if (err != nil) != tt.wantErr {
				t.Errorf("makeEventFunc() = %v, wantErr %v", err, tt.wantErr)
				return
			}

			switch tt.wantErr {
			case true:
				if tt.want.OperationStatus != got.OperationStatus {
					t.Errorf("response = %v; want %v", got.OperationStatus, tt.want.OperationStatus)
				}

			case false:
				if !reflect.DeepEqual(tt.want, got) {
					t.Errorf("response = %v; want %v", got, tt.want)
				}

			}

		})
	}
}

//loadEvent is a helper function that unmarshal the event from a file.
func loadEvent(path string, evt *event) *event {
	validevent, err := openFixture(path)
	if err != nil {
		log.Fatalf("Unable to read fixture: %v", err)
	}

	if err := json.Unmarshal(validevent, evt); err != nil {
		log.Fatalf("Marshaling error with event: %v", err)
	}
	return evt
}

func TestMakeTestEventFunc(t *testing.T) {
	start := time.Now()
	future := start.Add(time.Minute * 15)

	tc, cancel := context.WithDeadline(context.Background(), future)

	defer cancel()

	lc := lambdacontext.NewContext(tc, &lambdacontext.LambdaContext{})

	f1 := func(callback map[string]interface{}, s *session.Session) handler.ProgressEvent {
		response := handler.ProgressEvent{
			OperationStatus: handler.Success,
			Message:         "Create complete",
		}
		return response
	}

	type args struct {
		h     Handler
		ctx   context.Context
		event *testEvent
	}
	tests := []struct {
		name    string
		args    args
		want    handler.ProgressEvent
		wantErr bool
	}{
		{"Test simple CREATE", args{&MockHandler{f1}, lc, loadTestEvent("test.create.json", &testEvent{})}, handler.ProgressEvent{
			OperationStatus: handler.Success,
			Message:         "Create complete",
		}, false},
		{"Test simple READ", args{&MockHandler{f1}, lc, loadTestEvent("test.read.json", &testEvent{})}, handler.ProgressEvent{
			OperationStatus: handler.Success,
			Message:         "Create complete",
		}, false},
		{"Test simple DELETE", args{&MockHandler{f1}, lc, loadTestEvent("test.delete.json", &testEvent{})}, handler.ProgressEvent{
			OperationStatus: handler.Success,
			Message:         "Create complete",
		}, false},
		{"Test simple INVALID", args{&MockHandler{f1}, lc, loadTestEvent("test.INVALID.json", &testEvent{})}, handler.ProgressEvent{
			OperationStatus: handler.Failed,
			Message:         "InvalidRequest: No action/invalid action specified",
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := makeTestEventFunc(tt.args.h)
			got, err := f(tt.args.ctx, tt.args.event)

			if (err != nil) != tt.wantErr {
				t.Errorf("makeEventFunc() = %v, wantErr %v", err, tt.wantErr)
				return
			}

			switch tt.wantErr {
			case true:
				if tt.want.OperationStatus != got.OperationStatus {
					t.Errorf("response = %v; want %v", got.OperationStatus, tt.want.OperationStatus)
				}

			case false:
				if !reflect.DeepEqual(tt.want, got) {
					t.Errorf("response = %v; want %v", got, tt.want)
				}

			}
		})
	}
}

//loadEvent is a helper function that unmarshal the event from a file.
func loadTestEvent(path string, evt *testEvent) *testEvent {
	validevent, err := openFixture(path)
	if err != nil {
		log.Fatalf("Unable to read fixture: %v", err)
	}

	if err := json.Unmarshal(validevent, evt); err != nil {
		log.Fatalf("Marshaling error with event: %v", err)
	}
	return evt
}
