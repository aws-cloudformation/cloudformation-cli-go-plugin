package cfn

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/action"
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
			t.Fatalf("Marshaling error uncaught")
		}
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
