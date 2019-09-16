package action

import (
	"strings"
	"testing"
)

// TestConvert tests the string to action conversion.
func TestConvert(t *testing.T) {
	t.Run("String List", func(t *testing.T) {
		stringActions := map[string]Action{
			"create": Create,
			"read":   Read,
			"update": Update,
			"delete": Delete,
			"list":   List,
		}

		for k, v := range stringActions {
			if Convert(k) != v {
				t.Fatal("Invalid return value")
			}
		}
	})

	t.Run("Unknown action string", func(t *testing.T) {
		if Convert("NotAnAction") != Unknown {
			t.Fatal("Invalid return value")
		}
	})
}

// TestString tests the conversion from Action to string
func TestString(t *testing.T) {
	t.Run("Action Strings", func(t *testing.T) {
		stringActions := map[Action]string{
			Create: "create",
			Read:   "read",
			Update: "update",
			Delete: "delete",
			List:   "list",
		}

		for k, v := range stringActions {
			if k.String() != strings.ToUpper(v) {
				t.Fatal("Invalid return value")
			}
		}
	})
}
