package errorCode

import (
	"strings"
	"testing"
)

// TestConvert tests the string to action conversion.
func TestConvert(t *testing.T) {
	t.Run("String List", func(t *testing.T) {
		stringStatus := map[string]Status{
			"NotUpdatable":            NotUpdatable,
			"InvalidRequest":          InvalidRequest,
			"AccessDenied":            AccessDenied,
			"InvalidCredentials":      InvalidCredentials,
			"AlreadyExists":           AlreadyExists,
			"NotFound":                NotFound,
			"ResourceConflict":        ResourceConflict,
			"Throttling":              Throttling,
			"ServiceLimitExceeded":    ServiceLimitExceeded,
			"NotStabilized":           NotStabilized,
			"GeneralServiceException": GeneralServiceException,
			"ServiceInternalError":    ServiceInternalError,
			"NetworkFailure":          NetworkFailure,
			"InternalFailure":         InternalFailure,
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
			NotUpdatable:            "NotUpdatable",
			InvalidRequest:          "InvalidRequest",
			AccessDenied:            "AccessDenied",
			InvalidCredentials:      "InvalidCredentials",
			AlreadyExists:           "AlreadyExists",
			NotFound:                "NotFound",
			ResourceConflict:        "ResourceConflict",
			Throttling:              "Throttling",
			ServiceLimitExceeded:    "ServiceLimitExceeded",
			NotStabilized:           "NotStabilized",
			GeneralServiceException: "GeneralServiceException",
			ServiceInternalError:    "ServiceInternalError",
			NetworkFailure:          "NetworkFailure",
			InternalFailure:         "InternalFailure",
		}

		for k, v := range stringActions {
			if strings.ToUpper(k.String()) != strings.ToUpper(v) {
				t.Fatalf("Invalid return value: %v != %v", k, v)
			}
		}
	})
}
