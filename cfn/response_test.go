package cfn

import (
	"encoding/json"
	"errors"
	"reflect"
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

	expected := `{"message":"foo","operationStatus":"SUCCESS","resourceModel":"bar","errorCode":"baz - quux: mooz","bearerToken":"xyzzy"}`

	actual, err := json.Marshal(r)

	if err != nil {
		t.Errorf("Unexpected error marshaling response JSON: %s", err)
	}

	if string(actual) != expected {
		t.Errorf("Incorrect JSON: %s", string(actual))
	}
}

func TestStringifyModel(t *testing.T) {
	testCases := []struct{ Foo interface{} }{
		{Foo: "bar"},
		{Foo: 42},
		{Foo: []interface{}{"bar", 42}},
		{Foo: map[string]interface{}{"bar": "baz", "quux": 42}},
	}

	expecteds := []map[string]interface{}{
		{"Foo": "bar"},
		{"Foo": "42"},
		{"Foo": []interface{}{"bar", "42"}},
		{"Foo": map[string]interface{}{"bar": "baz", "quux": "42"}},
	}

	for i, testCase := range testCases {
		actual := stringifyModel(testCase)
		expected := expecteds[i]

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%#v != %#v", actual, expected)
		}
	}
}
