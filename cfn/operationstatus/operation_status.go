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
		if status == v {
			return Status(i)
		}
	}

	return 0
}

const (
	// Create ...
	Create Status = iota + 1

	// InProgress ...
	InProgress

	// Complete ...
	Complete

	// Failed ...
	Failed
)
