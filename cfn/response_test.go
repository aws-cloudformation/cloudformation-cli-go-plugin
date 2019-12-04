package cfn

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"encoding/json"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/encoding"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
)

func TestMarshalJSON(t *testing.T) {
	type Model struct {
		Name    *encoding.String
		Version *encoding.Float
	}

	r := response{
		Message:         "foo",
		OperationStatus: handler.Success,
		ResourceModel: Model{
			Name:    encoding.NewString("Douglas"),
			Version: encoding.NewFloat(42.1),
		},
		ErrorCode:   cfnerr.New("baz", "quux", errors.New("mooz")),
		BearerToken: "xyzzy",
	}

	expected := `{"message":"foo","operationStatus":"SUCCESS","resourceModel":{"Name":"Douglas","Version":"42.1"},"errorCode":"baz","bearerToken":"xyzzy"}`

	actual, err := json.Marshal(r)
	if err != nil {
		t.Errorf("Unexpected error marshaling response JSON: %s", err)
	}

	if diff := cmp.Diff(string(actual), expected); diff != "" {
		t.Errorf(diff)
	}
}
