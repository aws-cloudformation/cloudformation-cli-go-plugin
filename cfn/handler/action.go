package handler

type Action string

const (
	// Unknown ...
	UnknownAction Action = "UNKNOWN"

	// Create ...
	Create Action = "CREATE"

	// Read ...
	Read Action = "READ"

	// Update ...
	Update Action = "UPDATE"

	// Delete ...
	Delete Action = "DELETE"

	// List ...
	List Action = "LIST"
)
