package errs

import "fmt"

type ResourceAlreadyExistsError struct {
	err                string //error description
	ResourceTypeName   string //resource type name
	ResourceIdentifier string //resource identifier

}

func (e *ResourceAlreadyExistsError) Error() string {
	return fmt.Sprintf("Resource of type '%s' with identifier '%s' already exists.\n%s", e.ResourceTypeName, e.ResourceIdentifier, e.err)
}
