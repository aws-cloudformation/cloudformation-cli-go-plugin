package handler

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/google/go-cmp/cmp"

	"encoding/json"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/encoding"
)

func TestProgressEventMarshalJSON(t *testing.T) {
	type Model struct {
		Name    *encoding.String
		Version *encoding.Float
	}

	for _, tt := range []struct {
		name     string
		event    ProgressEvent
		expected string
	}{
		{
			name: "not updatable",
			event: ProgressEvent{
				Message:         "foo",
				OperationStatus: Failed,
				ResourceModel: Model{
					Name:    encoding.NewString("Douglas"),
					Version: encoding.NewFloat(42.1),
				},
				HandlerErrorCode: cloudformation.HandlerErrorCodeNotUpdatable,
			},
			expected: `{"status":"FAILED","errorCode":"NotUpdatable","message":"foo","resourceModel":{"Name":"Douglas","Version":"42.1"},"resourceModels":null}`,
		},
		{
			name: "list with 1 result",
			event: ProgressEvent{
				OperationStatus: Success,
				ResourceModels: []interface{}{
					Model{
						Name:    encoding.NewString("Douglas"),
						Version: encoding.NewFloat(42.1),
					},
				},
			},
			expected: `{"status":"SUCCESS","resourceModels":[{"Name":"Douglas","Version":"42.1"}]}`,
		},
		{
			name: "list with empty array",
			event: ProgressEvent{
				OperationStatus: Success,
				ResourceModels:  []interface{}{},
			},
			expected: `{"status":"SUCCESS","resourceModels":[]}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {

			actual, err := json.Marshal(tt.event)
			if err != nil {
				t.Errorf("Unexpected error marshaling event JSON: %s", err)
			}

			if diff := cmp.Diff(string(actual), tt.expected); diff != "" {
				t.Errorf(diff)
			}
		})
	}

}
