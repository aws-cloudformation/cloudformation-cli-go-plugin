package handler

import (
	"testing"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/encoding"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshal(t *testing.T) {
	type Detail struct {
		Build        *encoding.Int
		IsProduction *encoding.Bool
	}

	type Model struct {
		Name    *encoding.String
		Version *encoding.Float
		Detail  Detail
	}

	req := Request{
		LogicalResourceID:              "foo",
		previousResourcePropertiesBody: []byte(`{"Name":"bar","Version":"0.1","Detail":{"Build":"57","IsProduction":"false"}}`),
		resourcePropertiesBody:         []byte(`{"Name":"baz","Version":"2.3","Detail":{"Build":"69","IsProduction":"true"}}`),
	}

	expectedPrevious := Model{
		Name:    encoding.NewString("bar"),
		Version: encoding.NewFloat(0.1),
		Detail: Detail{
			Build:        encoding.NewInt(57),
			IsProduction: encoding.NewBool(false),
		},
	}

	expectedCurrent := Model{
		Name:    encoding.NewString("baz"),
		Version: encoding.NewFloat(2.3),
		Detail: Detail{
			Build:        encoding.NewInt(69),
			IsProduction: encoding.NewBool(true),
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
