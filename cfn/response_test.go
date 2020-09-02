package cfn

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/google/go-cmp/cmp"

	"encoding/json"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/encoding"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
)

func TestResponseMarshalJSON(t *testing.T) {
	type Model struct {
		Name    *encoding.String
		Version *encoding.Float
	}

	for _, tt := range []struct {
		name     string
		response response
		expected string
	}{
		{
			name: "updated failed",
			response: response{
				Message:         "foo",
				OperationStatus: handler.Failed,
				ResourceModel: Model{
					Name:    encoding.NewString("Douglas"),
					Version: encoding.NewFloat(42.1),
				},
				ErrorCode:   cloudformation.HandlerErrorCodeNotUpdatable,
				BearerToken: "xyzzy",
			},
			expected: `{"message":"foo","status":"FAILED","resourceModel":{"Name":"Douglas","Version":"42.1"},"errorCode":"NotUpdatable","bearerToken":"xyzzy","resourceModels":null}`,
		},
		{
			name: "list with 1 result",
			response: response{
				OperationStatus: handler.Success,
				ResourceModels: []interface{}{
					Model{
						Name:    encoding.NewString("Douglas"),
						Version: encoding.NewFloat(42.1),
					},
				},
				BearerToken: "xyzzy",
			},
			expected: `{"status":"SUCCESS","bearerToken":"xyzzy","resourceModels":[{"Name":"Douglas","Version":"42.1"}]}`,
		},
		{
			name: "list with empty array",
			response: response{
				OperationStatus: handler.Success,
				ResourceModels:  []interface{}{},
				BearerToken:     "xyzzy",
			},
			expected: `{"status":"SUCCESS","bearerToken":"xyzzy","resourceModels":[]}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {

			actual, err := json.Marshal(tt.response)
			if err != nil {
				t.Errorf("Unexpected error marshaling response JSON: %s", err)
			}

			if diff := cmp.Diff(string(actual), tt.expected); diff != "" {
				t.Errorf(diff)
			}
		})
	}

}
