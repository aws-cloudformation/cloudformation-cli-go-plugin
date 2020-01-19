package encoding_test

import (
	"testing"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/encoding"
	"github.com/google/go-cmp/cmp"
)

func TestStringify(t *testing.T) {
	type Struct struct {
		S string
	}

	s := "foo"
	b := true
	i := 42
	f := 3.14
	l := []interface{}{s, b, i, f}
	m := map[string]interface{}{
		"l": l,
	}
	o := Struct{S: s}
	var nilPointer *Struct

	for _, testCase := range []struct {
		data     interface{}
		expected interface{}
	}{
		{s, "foo"},
		{b, "true"},
		{i, "42"},
		{f, "3.14"},
		{l, []interface{}{"foo", "true", "42", "3.14"}},
		{m, map[string]interface{}{"l": []interface{}{"foo", "true", "42", "3.14"}}},
		{o, map[string]interface{}{"S": "foo"}},

		{&s, "foo"},
		{&b, "true"},
		{&i, "42"},
		{&f, "3.14"},
		{&l, []interface{}{"foo", "true", "42", "3.14"}},
		{&m, map[string]interface{}{"l": []interface{}{"foo", "true", "42", "3.14"}}},
		{&o, map[string]interface{}{"S": "foo"}},

		{nilPointer, nil},

		{nil, nil},
	} {
		actual, err := encoding.Stringify(testCase.data)
		if err != nil {
			t.Fatal(err)
		}

		if d := cmp.Diff(actual, testCase.expected); d != "" {
			t.Errorf(d)
		}
	}
}
