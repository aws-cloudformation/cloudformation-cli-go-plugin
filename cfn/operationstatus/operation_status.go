package operationstatus

import (
	"strings"
)

// Status is the resource provider status
//
// The status will be reported back to the Resource Provider
// API in the form a ProgressEvent from the handler (Read/Update/etc)
type Status int

// Formats the status as a string
//
// Example
//
//  // Will return "InProgress"
//	operationstatus.InProgress.String()
func (c Status) String() string {
	if c < InProgress || c > Failed {
		return "Unknown"
	}

	statuses := []string{
		"Unknown",
		"In_Progress",
		"Success",
		"Failed",
	}

	return strings.ToUpper(statuses[c])
}

// Convert turns a string into a status
//
// Example
//
//  // will return operationstatus.InProgress
// 	operationstatus.Convert("InProgress")
func Convert(s string) Status {
	status := strings.ToUpper(s)

	statuses := []string{
		"Unknown",
		"In_Progress",
		"Success",
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
	// Unknown represents all states that aren't covered
	// elsewhere
	Unknown Status = iota

	// InProgress should be returned when a resource provider
	// is in the process of being operated on.
	InProgress

	// Success should be returned when the resource provider
	// has finished it's operation.
	Success

	// Failed should be returned when the resource provider
	// has failed
	Failed
)
