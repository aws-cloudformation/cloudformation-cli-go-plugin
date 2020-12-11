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
		Detail  *Detail
	}

	req := Request{
		LogicalResourceID:              "foo",
		previousResourcePropertiesBody: []byte(`{"Name":"bar","Version":"0.1","Detail":{"Build":"57","IsProduction":"false"}}`),
		resourcePropertiesBody:         []byte(`{"Name":"baz","Version":"2.3","Detail":{"Build":"69","IsProduction":"true"}}`),
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

func TestNestedUnmarshal(t *testing.T) {
	type Model struct {
		Name    *string
		Version *float64
		Detail  map[string]interface{}
	}

	req := Request{
		LogicalResourceID:              "foo",
		previousResourcePropertiesBody: []byte(`{"Name":"bar","Version":"0.1","Detail":{"Nested":{"Build":"57","IsProduction":"false"}}}`),
		resourcePropertiesBody:         []byte(`{"Name":"baz","Version":"2.3","Detail":{"Nested":{"Build":"69","IsProduction":"true"}}}`),
	}

	expectedPrevious := Model{
		Name:    aws.String("bar"),
		Version: aws.Float64(0.1),
		Detail: map[string]interface{}{
			"Nested": map[string]interface{}{
				"Build":        aws.Int(57),
				"IsProduction": aws.Bool(false),
			},
		},
	}

	expectedCurrent := Model{
		Name:    aws.String("baz"),
		Version: aws.Float64(2.3),
		Detail: map[string]interface{}{
			"Nested": map[string]interface{}{
				"Build":        aws.Int(69),
				"IsProduction": aws.Bool(true),
			},
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
