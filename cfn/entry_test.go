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

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/action"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/handler"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/metrics"
)

func TestMarshalling(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		validEvent, err := openFixture("request.read.json")
		if err != nil {
			t.Fatalf("Unable to read fixture: %v", err)
		}
		evt := &Event{}

		if err := json.Unmarshal([]byte(validEvent), evt); err != nil {
			t.Fatalf("Marshaling error with event: %v", err)
		}

		if evt.Action != action.Read {
			t.Fatalf("Incorrect action (%v), expected: %v", evt.Action.String(), action.Read.String())
		}

		if evt.RequestData.LogicalResourceID != "myBucket" {
			t.Fatalf("Incorrect Logical Resource ID: %v", evt.RequestData.LogicalResourceID)
		}
	})

	t.Run("Invalid Body", func(t *testing.T) {
		invalidEvent, err := openFixture("request.read.invalid.json")
		if err != nil {
			t.Fatalf("Unable to read fixture: %v", err)
		}

		evt := &Event{}
		if err := json.Unmarshal([]byte(invalidEvent), evt); err == nil {
			t.Fatalf("Marshaling failed to throw an error: %#v", err)
		}
	})
}

func TestRouter(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		actions := []action.Action{
			action.Create,
			action.Read,
			action.Update,
			action.Delete,
			action.List,
		}

		for _, a := range actions {
			fn, err := Router(a, &EmptyHandlers{})
			if err != nil {
				t.Fatalf("Unable to select '%v' handler: %v", a.String(), err)
			}

			if fn == nil {
				t.Fatalf("Handler was not returned")
			}
		}
	})

	t.Run("Failed Path", func(t *testing.T) {
		fn, err := Router(action.Unknown, &EmptyHandlers{})
		cfnErr := err.(cfnerr.Error)
		if cfnErr != nil && cfnErr.Code() != InvalidRequestError {
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
		validEvent, err := openFixture("request.read.json")
		if err != nil {
			t.Fatalf("Unable to read fixture: %v", err)
		}

		evt := &Event{}

		if err := json.Unmarshal([]byte(validEvent), evt); err != nil {
			t.Fatalf("Marshaling error with event: %v", err)
		}

		if err := ValidateEvent(evt); err != nil {
			t.Fatalf("Failed to validate: %v", err)
		}
	})

	t.Run("Failed Validation", func(t *testing.T) {
		validEvent, err := openFixture("request.read.invalid.validation.json")
		if err != nil {
			t.Fatalf("Unable to read fixture: %v", err)
		}

		evt := &Event{}

		if err := json.Unmarshal([]byte(validEvent), evt); err != nil {
			t.Fatalf("Marshaling error with event: %v", err)
		}

		if err := ValidateEvent(evt); err == nil {
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

	type args struct {
		handlerFn        HandlerFunc
		request          Request
		reqContext       *RequestContext
		metricsPublisher *metrics.Publisher
		action           action.Action
	}
	tests := []struct {
		name      string
		args      args
		want      Response
		wantErr   bool
		wantCount int
	}{
		{"TestMaxTriesShouldReturnError ", args{func(ctx context.Context, request Request) (Response, error) {
			time.Sleep(2 * time.Hour)
			return nil, nil
		}, handler.NewRequest(nil, nil, "foo", "bar"), &RequestContext{}, mockPub, action.Create,
		}, handler.NewResponse(), true, 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Invoke(tt.args.handlerFn, tt.args.request, tt.args.reqContext, tt.args.metricsPublisher, tt.args.action)
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
