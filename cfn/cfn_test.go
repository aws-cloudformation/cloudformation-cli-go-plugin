package cfn

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/scheduler"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

func Test_reschedule(t *testing.T) {

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

func Test_makeEventFuncFailedResponse(t *testing.T) {
	f1 := func() handler.ProgressEvent {
		return handler.ProgressEvent{}
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
		{"Test invalid READ", args{&MockHandler{f1}, context.Background(), loadEvent("request.read.invalid.validation.json", &event{})}, response{
			OperationStatus: handler.Failed,
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := makeEventFunc(tt.args.h)

			got, err := f(tt.args.ctx, tt.args.event)

			if (err != nil) != tt.wantErr {
				t.Errorf("makeEventFunc() = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want.OperationStatus != got.OperationStatus {
				t.Errorf("response = %v; want %v", got.OperationStatus, tt.want.OperationStatus)
			}

		})
	}
}

func Test_makeEventFunc(t *testing.T) {
	start := time.Now()
	future := start.Add(time.Minute * 15)

	tc, cancel := context.WithDeadline(context.Background(), future)

	defer cancel()

	lc := lambdacontext.NewContext(tc, &lambdacontext.LambdaContext{})

	f1 := func() handler.ProgressEvent {
		return handler.ProgressEvent{}
	}

	f2 := func() handler.ProgressEvent {
		return handler.ProgressEvent{
			OperationStatus:      handler.InProgress,
			Message:              "In Progress",
			CallbackDelaySeconds: 130,
		}
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
		{"Test simple CREATE async", args{&MockHandler{f2}, lc, loadEvent("request.create.json", &event{})}, response{
			BearerToken:     "123456",
			Message:         "In Progress",
			OperationStatus: handler.InProgress,
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
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("response = %v; want %v", got, tt.want)
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
