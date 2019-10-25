package cfn

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
)

func TestMarshalJSON(t *testing.T) {
	r := response{
		Message:         "foo",
		OperationStatus: handler.Success,
		ResourceModel:   "bar",
		ErrorCode:       cfnerr.New("baz", "quux", errors.New("mooz")),
		BearerToken:     "xyzzy",
	}

	expected := `{"message":"foo","operationStatus":"SUCCESS","resourceModel":"bar","errorCode":"baz","bearerToken":"xyzzy"}`

	actual, err := json.Marshal(r)

	if err != nil {
		t.Errorf("Unexpected error marshaling response JSON: %s", err)
	}

	if string(actual) != expected {
		t.Errorf("Incorrect JSON: %s", string(actual))
	}
}
