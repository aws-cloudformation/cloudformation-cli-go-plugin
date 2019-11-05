package encoding_test

import (
	"encoding/json"
	"testing"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/encoding"

	"github.com/google/go-cmp/cmp"
)

type Model struct {
	ModelString encoding.String  `json:"modelString"`
	ModelBool   encoding.Bool    `json:"modelBool"`
	ModelInt    encoding.Int     `json:"modelInt"`
	ModelFloat  encoding.Float   `json:"modelFloat"`
	ModelSlice  []Inner          `json:"modelSlice"`
	ModelMap    map[string]Inner `json:"modelMap"`
	ModelNested json.RawMessage  `json:"embedded"`
}

type Inner struct {
	InnerString encoding.String `json:"innerString,omitempty"`
	InnerBool   encoding.Bool   `json:"innerBool"`
	InnerInt    encoding.Int    `json:"innerInt"`
	InnerFloat  encoding.Float  `json:"innerFloat"`
}

var model = Model{
	ModelString: "foo",
	ModelBool:   false,
	ModelInt:    42,
	ModelFloat:  3.14,
	ModelSlice: []Inner{
		{
			InnerString: "bar",
			InnerBool:   true,
			InnerInt:    43,
			InnerFloat:  6.28,
		},
	},
	ModelMap: map[string]Inner{
		"ModelMapInner": {
			InnerString: "baz",
			InnerBool:   false,
			InnerInt:    44,
			InnerFloat:  9.42,
		},
	},
	ModelNested: []byte(`{"innerBool":"true","innerFloat":"12.56","innerInt":"45"}`),
}

var stringified = map[string]interface{}{
	"modelString": "foo",
	"modelBool":   "false",
	"modelInt":    "42",
	"modelFloat":  "3.14",
	"modelSlice": []interface{}{
		map[string]interface{}{
			"innerString": "bar",
			"innerBool":   "true",
			"innerInt":    "43",
			"innerFloat":  "6.28",
		},
	},
	"modelMap": map[string]interface{}{
		"ModelMapInner": map[string]interface{}{
			"innerString": "baz",
			"innerBool":   "false",
			"innerInt":    "44",
			"innerFloat":  "9.42",
		},
	},
	"embedded": map[string]interface{}{
		"innerBool":  "true",
		"innerInt":   "45",
		"innerFloat": "12.56",
	},
}

func TestMarshal(t *testing.T) {
	data, err := json.Marshal(model)
	if err != nil {
		t.Error(err)
	}

	actual := make(map[string]interface{})
	err = json.Unmarshal(data, &actual)
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(actual, stringified); diff != "" {
		t.Error(diff)
	}
}

func TestUnmarshal(t *testing.T) {
	data, err := json.Marshal(stringified)
	if err != nil {
		panic(err)
	}

	actual := Model{}

	err = json.Unmarshal(data, &actual)
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(actual, model); diff != "" {
		t.Error(diff)
	}
}
