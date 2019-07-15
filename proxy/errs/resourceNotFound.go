package errs

import "fmt"

type ResourceNotFoundError struct {
	Err                string //error description
	ResourceTypeName   string //resource type name
	ResourceIdentifier string //resource identifier

}

func (e *ResourceNotFoundError) Error() string {
	return fmt.Sprintf("Resource of type '%s' with identifier '%s' was not found.\n%s", e.ResourceTypeName, e.ResourceIdentifier, e.Err)
}
