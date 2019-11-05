package handler

import (
	"testing"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/encoding"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshal(t *testing.T) {
	type Detail struct {
		Build        encoding.Int
		IsProduction encoding.Bool
	}

	type Model struct {
		Name    string
		Version encoding.Float
		Detail  Detail
	}

	req := Request{
		LogicalResourceID:              "foo",
		previousResourcePropertiesBody: []byte(`{"Name":"bar","Version":"0.1","Detail":{"Build":"57","IsProduction":"false"}}`),
		resourcePropertiesBody:         []byte(`{"Name":"baz","Version":"2.3","Detail":{"Build":"69","IsProduction":"true"}}`),
	}

	expectedPrevious := Model{
		Name:    "bar",
		Version: 0.1,
		Detail: Detail{
			Build:        57,
			IsProduction: false,
		},
	}

	expectedCurrent := Model{
		Name:    "baz",
		Version: 2.3,
		Detail: Detail{
			Build:        69,
			IsProduction: true,
		},
	}

	actual := Model{}

	// Previous body
	err := req.UnmarshalPrevious(&actual)
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(actual, expectedPrevious); diff != "" {
		t.Errorf(diff)
	}

	// Current body
	err = req.Unmarshal(&actual)
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(actual, expectedCurrent); diff != "" {
		t.Errorf(diff)
	}
}
