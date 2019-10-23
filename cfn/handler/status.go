package handler

// Status is the resource provider status
//
// The status will be reported back to the Resource Provider
// API in the form a ProgressEvent from the handler (Read/Update/etc)
type Status string

const (
	// Unknown represents all states that aren't covered
	// elsewhere
	UnknownStatus Status = "UNKNOWN"

	// InProgress should be returned when a resource provider
	// is in the process of being operated on.
	InProgress Status = "IN_PROGRESS"

	// Success should be returned when the resource provider
	// has finished it's operation.
	Success Status = "SUCCESS"

	// Failed should be returned when the resource provider
	// has failed
	Failed Status = "FAILED"
)
