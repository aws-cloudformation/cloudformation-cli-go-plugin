package errs

import "fmt"

type ResourceNotFoundError struct {
	ResourceTypeName   string //resource type name
	ResourceIdentifier string //resource identifier

}

func (e *ResourceNotFoundError) Error() string {
	return fmt.Sprintf("Resource of type '%s' with identifier '%s' was not found.", e.ResourceTypeName, e.ResourceIdentifier)
}
