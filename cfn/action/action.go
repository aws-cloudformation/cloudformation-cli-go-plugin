package action

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

	return Unknown
}

const (
	// Unknown ...
	Unknown Action = iota

	// Create ...
	Create

	// Read ...
	Read

	// Update ...
	Update

	// Delete ...
	Delete

	// List ...
	List
)
