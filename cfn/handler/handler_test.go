package handler

import (
	"testing"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
)

type Props struct {
	Color string `json:"color"`
}

func TestNewRequest(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		prev := Props{}
		curr := Props{}

		req := NewRequest([]byte(`{"color": "red"}`), []byte(`{"color": "green"}`), "foo", "123")

		if err := req.UnmarshalPrevious(&prev); err != nil {
			t.Fatalf("Unable to unmarshal props: %v", err)
		}

		if prev.Color != "red" {
			t.Fatalf("Previous Properties don't match: %v", prev.Color)
		}

		if err := req.Unmarshal(&curr); err != nil {
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

			err := req.Unmarshal(&invalid)
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

			err := req.Unmarshal(&invalid)
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

			err := req.UnmarshalPrevious(&invalid)
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

			err := req.UnmarshalPrevious(&invalid)
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
