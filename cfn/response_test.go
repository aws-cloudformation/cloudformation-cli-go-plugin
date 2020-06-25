package cfn

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/google/go-cmp/cmp"

	"encoding/json"

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
		ErrorCode:   cloudformation.HandlerErrorCodeNotUpdatable,
		BearerToken: "xyzzy",
	}

	expected := `{"message":"foo","status":"SUCCESS","resourceModel":{"Name":"Douglas","Version":"42.1"},"errorCode":"NotUpdatable","bearerToken":"xyzzy"}`

	actual, err := json.Marshal(r)
	if err != nil {
		t.Errorf("Unexpected error marshaling response JSON: %s", err)
	}

	if diff := cmp.Diff(string(actual), expected); diff != "" {
		t.Errorf(diff)
	}
}
