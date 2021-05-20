package encoding_test

import (
	"testing"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/encoding"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/go-cmp/cmp"
)

func TestUnstringifyStruct(t *testing.T) {
	type Model struct {
		S  string
		SP *string
		B  bool
		BP *bool
		I  int
		IP *int
		F  float64
		FP *float64
	}

	expected := Model{
		S:  "foo",
		SP: aws.String("bar"),
		B:  true,
		BP: aws.Bool(true),
		I:  2147483648, // math.MaxInt32 + 1
		IP: aws.Int(2147483648),
		F:  3.14,
		FP: aws.Float64(22),
	}

	t.Run("Convert strings", func(t *testing.T) {
		var actual Model

		err := encoding.Unstringify(map[string]interface{}{
			"S":  "foo",
			"SP": "bar",
			"B":  "true",
			"BP": "true",
			"I":  "2147483648",
			"IP": "2147483648",
			"F":  "3.14",
			"FP": "22",
		}, &actual)

		if err != nil {
			t.Fatal(err)
		}

		if d := cmp.Diff(actual, expected); d != "" {
			t.Error(d)
		}
	})

	t.Run("Original types", func(t *testing.T) {
		var actual Model

		err := encoding.Unstringify(map[string]interface{}{
			"S":  "foo",
			"SP": "bar",
			"B":  true,
			"BP": true,
			"I":  2147483648,
			"IP": 2147483648,
			"F":  3.14,
			"FP": 22.0,
		}, &actual)

		if err != nil {
			t.Fatal(err)
		}

		if d := cmp.Diff(actual, expected); d != "" {
			t.Error(d)
		}
	})

	t.Run("Compatible types", func(t *testing.T) {
		var actual Model

		err := encoding.Unstringify(map[string]interface{}{
			"S":  "foo",
			"SP": "bar",
			"B":  true,
			"BP": true,
			"I":  float64(2147483648),
			"IP": float64(2147483648),
			"F":  3.14,
			"FP": int(22),
		}, &actual)

		if err != nil {
			t.Fatal(err)
		}

		if d := cmp.Diff(actual, expected); d != "" {
			t.Error(d)
		}
	})
}

func TestUnstringifySlices(t *testing.T) {
	type Model struct {
		S  []string
		SP []*string
		B  []bool
		BP []*bool
		I  []int
		IP []*int
		F  []float64
		FP []*float64
	}

	expected := Model{
		S:  []string{"foo"},
		SP: []*string{aws.String("bar")},
		B:  []bool{true},
		BP: []*bool{aws.Bool(true)},
		I:  []int{42},
		IP: []*int{aws.Int(42)},
		F:  []float64{3.14},
		FP: []*float64{aws.Float64(22)},
	}

	t.Run("Convert strings", func(t *testing.T) {
		var actual Model

		err := encoding.Unstringify(map[string]interface{}{
			"S":  []interface{}{"foo"},
			"SP": []interface{}{"bar"},
			"B":  []interface{}{"true"},
			"BP": []interface{}{"true"},
			"I":  []interface{}{"42"},
			"IP": []interface{}{"42"},
			"F":  []interface{}{"3.14"},
			"FP": []interface{}{"22"},
		}, &actual)

		if err != nil {
			t.Fatal(err)
		}

		if d := cmp.Diff(actual, expected); d != "" {
			t.Error(d)
		}
	})

	t.Run("Original types", func(t *testing.T) {
		var actual Model

		err := encoding.Unstringify(map[string]interface{}{
			"S":  []interface{}{"foo"},
			"SP": []interface{}{"bar"},
			"B":  []interface{}{true},
			"BP": []interface{}{true},
			"I":  []interface{}{42},
			"IP": []interface{}{42},
			"F":  []interface{}{3.14},
			"FP": []interface{}{22.0},
		}, &actual)

		if err != nil {
			t.Fatal(err)
		}

		if d := cmp.Diff(actual, expected); d != "" {
			t.Error(d)
		}
	})

	t.Run("Compatible types", func(t *testing.T) {
		var actual Model

		err := encoding.Unstringify(map[string]interface{}{
			"S":  []interface{}{"foo"},
			"SP": []interface{}{"bar"},
			"B":  []interface{}{true},
			"BP": []interface{}{true},
			"I":  []interface{}{float64(42)},
			"IP": []interface{}{float64(42)},
			"F":  []interface{}{3.14},
			"FP": []interface{}{int(22)},
		}, &actual)

		if err != nil {
			t.Fatal(err)
		}

		if d := cmp.Diff(actual, expected); d != "" {
			t.Error(d)
		}
	})
}

func TestUnstringifyMaps(t *testing.T) {
	type Model struct {
		S  map[string]string
		SP map[string]*string
		B  map[string]bool
		BP map[string]*bool
		I  map[string]int
		IP map[string]*int
		F  map[string]float64
		FP map[string]*float64
	}

	expected := Model{
		S:  map[string]string{"Val": "foo"},
		SP: map[string]*string{"Val": aws.String("bar")},
		B:  map[string]bool{"Val": true},
		BP: map[string]*bool{"Val": aws.Bool(true)},
		I:  map[string]int{"Val": 42},
		IP: map[string]*int{"Val": aws.Int(42)},
		F:  map[string]float64{"Val": 3.14},
		FP: map[string]*float64{"Val": aws.Float64(22)},
	}

	t.Run("Convert strings", func(t *testing.T) {
		var actual Model

		err := encoding.Unstringify(map[string]interface{}{
			"S":  map[string]interface{}{"Val": "foo"},
			"SP": map[string]interface{}{"Val": "bar"},
			"B":  map[string]interface{}{"Val": "true"},
			"BP": map[string]interface{}{"Val": "true"},
			"I":  map[string]interface{}{"Val": "42"},
			"IP": map[string]interface{}{"Val": "42"},
			"F":  map[string]interface{}{"Val": "3.14"},
			"FP": map[string]interface{}{"Val": "22"},
		}, &actual)

		if err != nil {
			t.Fatal(err)
		}

		if d := cmp.Diff(actual, expected); d != "" {
			t.Error(d)
		}
	})

	t.Run("Original types", func(t *testing.T) {
		var actual Model

		err := encoding.Unstringify(map[string]interface{}{
			"S":  map[string]interface{}{"Val": "foo"},
			"SP": map[string]interface{}{"Val": "bar"},
			"B":  map[string]interface{}{"Val": true},
			"BP": map[string]interface{}{"Val": true},
			"I":  map[string]interface{}{"Val": 42},
			"IP": map[string]interface{}{"Val": 42},
			"F":  map[string]interface{}{"Val": 3.14},
			"FP": map[string]interface{}{"Val": 22.0},
		}, &actual)

		if err != nil {
			t.Fatal(err)
		}

		if d := cmp.Diff(actual, expected); d != "" {
			t.Error(d)
		}
	})

	t.Run("Compatible types", func(t *testing.T) {
		var actual Model

		err := encoding.Unstringify(map[string]interface{}{
			"S":  map[string]interface{}{"Val": "foo"},
			"SP": map[string]interface{}{"Val": "bar"},
			"B":  map[string]interface{}{"Val": true},
			"BP": map[string]interface{}{"Val": true},
			"I":  map[string]interface{}{"Val": float64(42)},
			"IP": map[string]interface{}{"Val": float64(42)},
			"F":  map[string]interface{}{"Val": 3.14},
			"FP": map[string]interface{}{"Val": int(22)},
		}, &actual)

		if err != nil {
			t.Fatal(err)
		}

		if d := cmp.Diff(actual, expected); d != "" {
			t.Error(d)
		}
	})
}

func TestUnstringifyPointers(t *testing.T) {
	type Model struct {
		SSP *[]string
		SMP *map[string]string
	}

	expected := Model{
		SSP: &[]string{"foo"},
		SMP: &map[string]string{"bar": "baz"},
	}

	var actual Model

	err := encoding.Unstringify(map[string]interface{}{
		"SSP": []interface{}{"foo"},
		"SMP": map[string]interface{}{"bar": "baz"},
	}, &actual)

	if err != nil {
		t.Fatal(err)
	}

	if d := cmp.Diff(actual, expected); d != "" {
		t.Error(d)
	}
}
