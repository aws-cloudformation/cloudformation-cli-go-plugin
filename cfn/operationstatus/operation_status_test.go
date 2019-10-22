package operationstatus

import (
	"strings"
	"testing"
)

// TestConvert tests the string to action conversion.
func TestConvert(t *testing.T) {
	t.Run("String List", func(t *testing.T) {
		stringStatus := map[string]Status{
			"IN_PROGRESS": InProgress,
			"SUCCESS":     Success,
			"FAILED":      Failed,
		}

		for k, v := range stringStatus {
			if Convert(k) != v {
				t.Fatalf("Invalid return value: %v != %v", k, v)
			}
		}
	})

	t.Run("Unknown operationstatus string", func(t *testing.T) {
		if Convert("NotAnAction") != Unknown {
			t.Fatal("Invalid return value")
		}
	})
}

// TestString tests the conversion from Status to string
func TestString(t *testing.T) {
	t.Run("OperationStatus Strings", func(t *testing.T) {
		stringActions := map[Status]string{
			InProgress: "IN_PROGRESS",
			Success:    "SUCCESS",
			Failed:     "FAILED",
		}

		for k, v := range stringActions {
			if strings.ToUpper(k.String()) != strings.ToUpper(v) {
				t.Fatalf("Invalid return value: %v != %v", k, v)
			}
		}
	})
}
