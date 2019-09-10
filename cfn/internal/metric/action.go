package metric

import (
	"strings"
)

// Action ...
type Action int

func (c Action) String() string {
	if c < Create || c > List {
		return "UNKNOWN"
	}

	actions := []string{
		"UNKNOWN",
		"CREATE",
		"READ",
		"UPDATE",
		"DELETE",
		"LIST",
	}

	return actions[c]
}

//Convert a string to an Action
func Convert(s string) Action {
	action := strings.ToUpper(s)

	actions := []string{
		"UNKNOWN",
		"CREATE",
		"READ",
		"UPDATE",
		"DELETE",
		"LIST",
	}

	for i, v := range actions {
		if action == v {
			return Action(i)
		}
	}

	return 0
}

const (
	// Create ...
	Create Action = iota + 1

	// Read ...
	Read

	// Update ...
	Update

	// Delete ...
	Delete

	// List ...
	List
)
