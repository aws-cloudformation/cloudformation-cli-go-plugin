package cfn

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/metrics"
)

func TestMarshalling(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		validevent, err := openFixture("request.read.json")
		if err != nil {
			t.Fatalf("Unable to read fixture: %v", err)
		}
		evt := &event{}

		if err := json.Unmarshal([]byte(validevent), evt); err != nil {
			t.Fatalf("Marshaling error with event: %v", err)
		}

		if evt.Action != readAction {
			t.Fatalf("Incorrect action (%v), expected: %v", evt.Action, readAction)
		}

		if evt.RequestData.LogicalResourceID != "myBucket" {
			t.Fatalf("Incorrect Logical Resource ID: %v", evt.RequestData.LogicalResourceID)
		}
	})

	t.Run("Invalid Body", func(t *testing.T) {
		invalidevent, err := openFixture("request.read.invalid.json")
		if err != nil {
			t.Fatalf("Unable to read fixture: %v", err)
		}

		evt := &event{}
		if err := json.Unmarshal([]byte(invalidevent), evt); err == nil {
			t.Fatalf("Marshaling failed to throw an error: %#v", err)
		}
	})
}

func TestRouter(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		actions := []string{
			createAction,
			readAction,
			updateAction,
			deleteAction,
			listAction,
		}

		for _, a := range actions {
			fn, err := router(a, &EmptyHandler{})
			if err != nil {
				t.Fatalf("Unable to select '%v' handler: %v", a, err)
			}

			if fn == nil {
				t.Fatalf("Handler was not returned")
			}
		}
	})

	t.Run("Failed Path", func(t *testing.T) {
		fn, err := router(unknownAction, &EmptyHandler{})
		cfnErr := err.(cfnerr.Error)
		if cfnErr != nil && cfnErr.Code() != invalidRequestError {
			t.Errorf("Unspecified error returned: %v", err)
		} else if err == nil {
			t.Errorf("There should have been an error")
		}

		if fn != nil {
			t.Fatalf("Handler should be nil")
		}
	})
}

func TestValidateEvent(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		validevent, err := openFixture("request.read.json")
		if err != nil {
			t.Fatalf("Unable to read fixture: %v", err)
		}

		evt := &event{}

		if err := json.Unmarshal([]byte(validevent), evt); err != nil {
			t.Fatalf("Marshaling error with event: %v", err)
		}

		if err := validateEvent(evt); err != nil {
			t.Fatalf("Failed to validate: %v", err)
		}
	})

	t.Run("Failed Validation", func(t *testing.T) {
		validevent, err := openFixture("request.read.invalid.validation.json")
		if err != nil {
			t.Fatalf("Unable to read fixture: %v", err)
		}

		evt := &event{}

		if err := json.Unmarshal([]byte(validevent), evt); err != nil {
			t.Fatalf("Marshaling error with event: %v", err)
		}

		if err := validateEvent(evt); err == nil {
			t.Fatalf("Failed to validate: %v", err)
		}
	})
}

func TestHandler(t *testing.T) {
	// no-op
}

func TestInvoke(t *testing.T) {
	mockClient := NewMockedMetrics()
	mockPub := metrics.New(mockClient)
	mockPub.SetResourceTypeName("dsf::fs::sfa")

	// For test purposes, set the timeout low
	Timeout = time.Second

	type args struct {
		handlerFn        handlerFunc
		request          handler.Request
		reqContext       *requestContext
		metricsPublisher *metrics.Publisher
		action           string
	}
	tests := []struct {
		name      string
		args      args
		want      handler.ProgressEvent
		wantErr   bool
		wantCount int
	}{
		{"TestMaxTriesShouldReturnError ", args{func(ctx context.Context, request handler.Request) handler.ProgressEvent {
			time.Sleep(2 * time.Hour)
			return handler.ProgressEvent{}
		}, handler.NewRequest(nil, nil, "foo"), &requestContext{}, mockPub, createAction,
		}, handler.ProgressEvent{}, true, 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := invoke(tt.args.handlerFn, tt.args.request, tt.args.reqContext, tt.args.metricsPublisher, tt.args.action)
			if (err != nil) != tt.wantErr {
				t.Errorf("Invoke() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {

			}

			if mockClient.HandlerInvocationCount != tt.wantCount {
				t.Errorf("InvocationCount= %v, wantCount %v", mockClient.HandlerInvocationCount, tt.wantCount)
			}
		})
	}
}

// helper func to load fixtures from the disk
func openFixture(name string) ([]byte, error) {
	d, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(filepath.Join(d, "test", "data", name))
}
