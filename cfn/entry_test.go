package cfn

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/action"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/cfnerr"
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
	t.Run("Happy Path", func(t *testing.T) {

	})
}

// helper func to load fixtures from the disk
func openFixture(name string) ([]byte, error) {
	d, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(filepath.Join(d, "test", "data", name))
}
