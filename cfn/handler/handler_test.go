package handler

import (
	"encoding/json"
	"testing"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/operationstatus"
)

type Props struct {
	Color string `json:"color"`
}

func TestNewRequest(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		prev := Props{}
		curr := Props{}

		req := NewRequest([]byte(`{"color": "red"}`), []byte(`{"color": "green"}`), "foo", "123")

		if err := req.PreviousResourceProperties(&prev); err != nil {
			t.Fatalf("Unable to unmarshal props: %v", err)
		}

		if prev.Color != "red" {
			t.Fatalf("Previous Properties don't match: %v", prev.Color)
		}

		if err := req.ResourceProperties(&curr); err != nil {
			t.Fatalf("Unable to unmarshal props: %v", err)
		}

		if curr.Color != "green" {
			t.Fatalf("Properties don't match: %v", curr.Color)
		}

		if req.BearerToken() != "123" {
			t.Fatalf("Invalid Bearer Token: %v", req.BearerToken())
		}

		if req.LogicalResourceID() != "foo" {
			t.Fatalf("Invalid Logical Resource ID: %v", req.LogicalResourceID())
		}

	})

	t.Run("ResourceProps", func(t *testing.T) {
		t.Run("Invalid Body", func(t *testing.T) {
			req := NewRequest([]byte(``), []byte(``), "foo", "123")

			invalid := struct {
				Color int `json:"color"`
			}{}

			err := req.ResourceProperties(&invalid)
			if err == nil {
				t.Fatalf("Didn't throw an error")
			}

			cfnErr := err.(cfnerr.Error)
			if cfnErr.Code() != BodyEmptyError {
				t.Fatalf("Wrong error returned: %v", err)
			}
		})

		t.Run("Invalid Marshal", func(t *testing.T) {
			req := NewRequest([]byte(`{"color": "red"}`), []byte(`{"color": "green"}`), "foo", "123")

			invalid := struct {
				Color int `json:"color"`
			}{}

			err := req.ResourceProperties(&invalid)
			if err == nil {
				t.Fatalf("Didn't throw an error")
			}

			cfnErr := err.(cfnerr.Error)
			if cfnErr.Code() != MarshalingError {
				t.Fatalf("Wrong error returned: %v", err)
			}
		})
	})

	t.Run("PreviousResourceProps", func(t *testing.T) {
		t.Run("Invalid Body", func(t *testing.T) {
			req := NewRequest([]byte(``), []byte(``), "foo", "123")

			invalid := struct {
				Color int `json:"color"`
			}{}

			err := req.PreviousResourceProperties(&invalid)
			if err == nil {
				t.Fatalf("Didn't throw an error")
			}

			cfnErr := err.(cfnerr.Error)
			if cfnErr.Code() != BodyEmptyError {
				t.Fatalf("Wrong error returned: %v", err)
			}
		})

		t.Run("Invalid Marshal", func(t *testing.T) {
			req := NewRequest([]byte(`{"color": "red"}`), []byte(`{"color": "green"}`), "foo", "123")

			invalid := struct {
				Color int `json:"color"`
			}{}

			err := req.PreviousResourceProperties(&invalid)
			if err == nil {
				t.Fatalf("Didn't throw an error")
			}

			cfnErr := err.(cfnerr.Error)
			if cfnErr.Code() != MarshalingError {
				t.Fatalf("Wrong error returned: %v", err)
			}
		})
	})
}

func TestNewResponse(t *testing.T) {
	t.Run("New Response", func(t *testing.T) {
		NewResponse()
	})
}

func TestNewFailedResponse(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		err := cfnerr.New("SomeFailure", "Some internal failure", nil)
		resp := NewFailedResponse(err)

		internalErr := resp.Error().(cfnerr.Error)
		if internalErr.Code() != "SomeFailure" {
			t.Fatalf("Wrong error returned: %v", err.Code())
		}

		if resp.Message() != "SomeFailure: Some internal failure" {
			t.Fatalf("Wrong message: %v", resp.Message())
		}

		if resp.OperationStatus() != operationstatus.Failed {
			t.Fatalf("Wrong operationstatus: %v", resp.OperationStatus())
		}
	})

}

func TestResponseMarshaling(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		colour := struct {
			Name        string
			Description string
		}{
			Name:        "Red",
			Description: "The colour red",
		}

		resp := NewResponse()
		resp.operationStatus = operationstatus.Success
		resp.message = "Completed"
		resp.resourceModel = colour

		b, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Problem marshaling: %v", err)
		}

		valid := `{"message":"Completed","operationStatus":"SUCCESS","resourceModel":{"Name":"Red","Description":"The colour red"}}`

		if valid != string(b) {
			t.Fatalf("Response is not the same")
		}
	})
}
