package cfn

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMarshalling(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		validevent, err := openFixture("request.read.json")
		if err != nil {
			t.Fatalf("Unable to read fixture: %v", err)
		}

		evt := &event{}

		if err := json.Unmarshal(validevent, evt); err != nil {
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
		if err := json.Unmarshal(invalidevent, evt); err == nil {
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
		if err != nil && err.Code() != invalidRequestError {
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

		if err := json.Unmarshal(validevent, evt); err != nil {
			t.Fatalf("Marshaling error with event: %v", err)
		}

		if err := validateEvent(evt); err != nil {
			t.Fatalf("Failed to validate: %v", err)
		}
	})

	t.Run("Failed Validation", func(t *testing.T) {
		invalidevent, err := openFixture("request.read.invalid.validation.json")
		if err != nil {
			t.Fatalf("Unable to read fixture: %v", err)
		}

		evt := &event{}

		if err := json.Unmarshal(invalidevent, evt); err != nil {
			t.Fatalf("Marshaling error with event: %v", err)
		}

		if err := validateEvent(evt); err != nil {
			t.Fatalf("Failed to validate: %v", err)
		}
	})
}

func TestHandler(t *testing.T) {
	// no-op
}

// helper func to load fixtures from the disk
func openFixture(name string) ([]byte, error) {
	d, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return os.ReadFile(filepath.Join(d, "test", "data", name))
}
