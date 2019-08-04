package cfnerr

// New base error
func New(code string, message string, origErr error) Error {
	var errs []error
	if origErr != nil {
		errs = append(errs, origErr)
	}
	return newBaseError(code, message, errs)
}

// NewBatchError groups one or more errors together for processing
func NewBatchError(code string, message string, origErrs []error) BatchedErrors {
	return newBaseError(code, message, origErrs)
}

// An Error wraps lower level errors with code, message and an original error.
// The underlying concrete error type may also satisfy other interfaces which
// can be to used to obtain more specific information about the error.
//
// Calling Error() or String() will always include the full information about
// an error based on its underlying type.
type Error interface {
	// inherit the base error interface
	error

	// Returns an error code
	Code() string

	// Returns the error message
	Message() string

	// Returns the original error
	OrigErr() error
}

// BatchedErrors is a batch of errors which also wraps lower level errors with
// code, message, and original errors. Calling Error() will include all errors
// that occurred in the batch.
type BatchedErrors interface {
	error

	// Returns all original errors
	OrigErrs() []error
}
