package encoding_test

import (
	"testing"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/encoding"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/go-cmp/cmp"
)

func TestUnstringifyStrings(t *testing.T) {
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
		I:  42,
		IP: aws.Int(42),
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
			"I":  "42",
			"IP": "42",
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
			"I":  42,
			"IP": 42,
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
			"I":  float64(42),
			"IP": float64(42),
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
