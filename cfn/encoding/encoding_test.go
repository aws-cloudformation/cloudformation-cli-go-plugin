package encoding_test

import (
	"encoding/json"
	"testing"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/encoding"

	"github.com/google/go-cmp/cmp"
)

type Model struct {
	ModelString *encoding.String `json:"modelString,omitempty"`
	ModelBool   *encoding.Bool   `json:"modelBool,omitempty"`
	ModelInt    *encoding.Int    `json:"modelInt,omitempty"`
	ModelFloat  *encoding.Float  `json:"modelFloat,omitempty"`
	ModelSlice  []Inner          `json:"modelSlice,omitempty"`
	ModelMap    map[string]Inner `json:"modelMap,omitempty"`
	ModelNested json.RawMessage  `json:"embedded,omitempty"`
}

type Inner struct {
	InnerString *encoding.String `json:"innerString,omitempty"`
	InnerBool   *encoding.Bool   `json:"innerBool,omitempty"`
	InnerInt    *encoding.Int    `json:"innerInt,omitempty"`
	InnerFloat  *encoding.Float  `json:"innerFloat"` // No omitempty
}

var model = Model{
	ModelBool:  encoding.NewBool(false),
	ModelInt:   encoding.NewInt(42),
	ModelFloat: encoding.NewFloat(3.14),
	ModelSlice: []Inner{
		{
			InnerString: encoding.NewString("bar"),
			InnerInt:    encoding.NewInt(43),
			InnerFloat:  encoding.NewFloat(6.28),
		},
	},
	ModelMap: map[string]Inner{
		"ModelMapInner": {
			InnerString: encoding.NewString("baz"),
			InnerBool:   encoding.NewBool(false),
			InnerFloat:  encoding.NewFloat(9.42),
		},
	},
	ModelNested: []byte(`{"innerBool":"true","innerFloat":null,"innerInt":"45"}`),
}

var stringified = map[string]interface{}{
	"modelBool":  "false",
	"modelInt":   "42",
	"modelFloat": "3.14",
	"modelSlice": []interface{}{
		map[string]interface{}{
			"innerString": "bar",
			"innerInt":    "43",
			"innerFloat":  "6.28",
		},
	},
	"modelMap": map[string]interface{}{
		"ModelMapInner": map[string]interface{}{
			"innerString": "baz",
			"innerBool":   "false",
			"innerFloat":  "9.42",
		},
	},
	"embedded": map[string]interface{}{
		"innerBool":  "true",
		"innerInt":   "45",
		"innerFloat": nil,
	},
}

func TestString(t *testing.T) {
	v := "Hello, world!"
	s := encoding.NewString(v)

	// Value
	if *s.Value() != v {
		t.Errorf("Value failed: %v", s.Value())
	}

	// Marshal
	data, err := json.Marshal(s)
	if err != nil {
		t.Error(err)
	}

	if string(data) != `"Hello, world!"` {
		t.Error("Marshal failed: " + string(data))
	}

	// Unmarshal value
	v = "Unmarshal me"
	data, err = json.Marshal(v)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, s)
	if err != nil {
		t.Error(err)
	}

	if *s.Value() != v {
		t.Errorf("Unmarshal value failed: %v", s.Value())
	}
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

	if diff := cmp.Diff(stringified, actual); diff != "" {
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

	if diff := cmp.Diff(model, actual); diff != "" {
		t.Error(diff)
	}
}
