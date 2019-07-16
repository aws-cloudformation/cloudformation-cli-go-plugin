package errs

import "fmt"

type HandlerPanicError struct {
	Stacktrace                 string //Error Stack trace
	CustomerFacingErrorMessage string //Customer facing errorMessage
	ResourceTypeName           string //resource type name
	ResourceIdentifier         string //resource identifier

}

func (e *HandlerPanicError) Error() string {
	return fmt.Sprintf("Resource of type '%s' with identifier '%s' return %s.\n%s", e.ResourceTypeName, e.ResourceIdentifier, e.CustomerFacingErrorMessage, e.Stacktrace)
}
