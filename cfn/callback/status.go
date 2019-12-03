package callback

//Status represents the status of the handler during invocation.
type Status string

const (
	// UnknownStatus represents all states that aren't covered
	// elsewhere.
	UnknownStatus Status = "UNKNOWN"

	// InProgress is when a resource provider
	// is in the process of being operated on.
	InProgress Status = "IN_PROGRESS"

	// Success is when the resource provider
	// has finished it's operation.
	Success Status = "SUCCESS"

	// Failed is when the resource provider
	// has failed.
	Failed Status = "FAILED"

	// Pending is the resource provider
	// initial state.
	Pending Status = "PENDING"
)
