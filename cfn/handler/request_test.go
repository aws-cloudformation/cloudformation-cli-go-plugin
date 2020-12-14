package handler

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/go-cmp/cmp"
)

func TestUnmarshal(t *testing.T) {
	type Detail struct {
		Build        *int
		IsProduction *bool
	}

	type Model struct {
		Name    *string
		Version *float64
		Detail  *Detail `json:"detail,omitempty"`
	}

	req := Request{
		LogicalResourceID:              "foo",
		previousResourcePropertiesBody: []byte(`{"Name":"bar","Version":"0.1","detail":{"Build":"57","IsProduction":"false"}}`),
		resourcePropertiesBody:         []byte(`{"Name":"baz","Version":"2.3","detail":{"Build":"69","IsProduction":"true"}}`),
	}

	expectedPrevious := Model{
		Name:    aws.String("bar"),
		Version: aws.Float64(0.1),
		Detail: &Detail{
			Build:        aws.Int(57),
			IsProduction: aws.Bool(false),
		},
	}

	expectedCurrent := Model{
		Name:    aws.String("baz"),
		Version: aws.Float64(2.3),
		Detail: &Detail{
			Build:        aws.Int(69),
			IsProduction: aws.Bool(true),
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
