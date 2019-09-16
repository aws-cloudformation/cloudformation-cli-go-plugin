package operationstatus

import (
	"strings"
)

// Status ...
type Status int

func (c Status) String() string {
	if c < InProgress || c > Failed {
		return "Unknown"
	}

	statuses := []string{
		"Unknown",
		"InProgress",
		"Complete",
		"Failed",
	}

	return statuses[c]
}

func Convert(s string) Status {
	status := strings.ToUpper(s)

	statuses := []string{
		"Unknown",
		"InProgress",
		"Complete",
		"Failed",
	}

	for i, v := range statuses {
		if status == strings.ToUpper(v) {
			return Status(i)
		}
	}

	return Unknown
}

const (
	// Unknown ...
	Unknown Status = iota

	// InProgress ...
	InProgress

	// Complete ...
	Complete

	// Failed ...
	Failed
)
