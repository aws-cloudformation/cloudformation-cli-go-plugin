package handler

// Status represents the status of the handler.
type Status string

const (
	// UnknownStatus represents all states that aren't covered
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
