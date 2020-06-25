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

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/encoding"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

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

	f3 := func(callback map[string]interface{}, s *session.Session) handler.ProgressEvent {
		return handler.ProgressEvent{
			OperationStatus: handler.Failed,
		}
	}

	f4 := func(callback map[string]interface{}, s *session.Session) (response handler.ProgressEvent) {
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
		{"Test CREATE failed", args{&MockHandler{f3}, lc, loadEvent("request.create.json", &event{})}, response{
			OperationStatus: handler.Failed,
			BearerToken:     "123456",
		}, false},
		{"Test simple CREATE async", args{&MockHandler{f2}, lc, loadEvent("request.create.json", &event{})}, response{
			BearerToken:          "123456",
			Message:              "In Progress",
			OperationStatus:      handler.InProgress,
			CallbackDelaySeconds: 130,
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
		{"Test wrap panic", args{&MockHandler{f4}, context.Background(), loadEvent("request.create.json", &event{})}, response{
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

func TestMakeEventFuncModel(t *testing.T) {
	start := time.Now()
	future := start.Add(time.Minute * 15)
	tc, cancel := context.WithDeadline(context.Background(), future)
	defer cancel()
	lc := lambdacontext.NewContext(tc, &lambdacontext.LambdaContext{})
	f1 := func(r handler.Request) handler.ProgressEvent {
		m := MockModel{}
		if len(r.CallbackContext) == 1 {
			if err := r.Unmarshal(&m); err != nil {
				return handler.NewFailedEvent(err)
			}
			return handler.ProgressEvent{
				OperationStatus: handler.Success,
				ResourceModel:   &m,
				Message:         "Success",
			}
		}
		if err := r.Unmarshal(&m); err != nil {
			return handler.NewFailedEvent(err)
		}
		m.Property2 = aws.String("change")
		return handler.ProgressEvent{
			OperationStatus:      handler.InProgress,
			Message:              "In Progress",
			CallbackDelaySeconds: 3,
			CallbackContext:      map[string]interface{}{"foo": "bar"},
			ResourceModel:        &m,
		}
	}
	type args struct {
		h     Handler
		ctx   context.Context
		event *event
	}
	tests := []struct {
		name string
		args args
		want MockModel
	}{
		{"Test CREATE async local with model change", args{&MockModelHandler{f1}, lc, loadEvent("request.create2.json", &event{})}, MockModel{
			Property1: aws.String("abc"),
			Property2: aws.String("change"),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := makeEventFunc(tt.args.h)
			got, err := f(tt.args.ctx, tt.args.event)
			if err != nil {
				t.Errorf("TestMakeEventFuncModel() = %v", err)
				return
			}
			model, err := encoding.Stringify(got.ResourceModel)
			if err != nil {
				t.Errorf("TestMakeEventFuncModel() = %v", err)
			}
			wantrModel, err := encoding.Stringify(tt.want)
			if err != nil {
				t.Errorf("TestMakeEventFuncModel() = %v", err)
			}
			if wantrModel != model {
				t.Errorf("response = %v; want %v", model, wantrModel)
			}

		})
	}
}
