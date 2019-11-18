/*
Package handler contains types that are passed into and out of a
resource provider's CRUDL functions.

A handler function receives a handler.Request, the current resource model,
and if appropriate, the previous resource model.

When a handler function is successful, it should set the appropriate properties
of the ProgressEvent that it returns.

For example, an "in progress" response might look like this:

	e := handler.NewProgressEvent()

	// Change status to "In progress"
	e.OperationStatus = InProgress

	// Explain the status change
	e.Message = "Creating resource " + req.LogicalResourceID

And a completed response like this:

	e := handler.NewProgressEvent()

	// Change status to "Succes"
	e.OperationStatus = Success

	// Explain the status change
	e.Message = "Created resource " + req.LogicalResourceID

	// Add the updated resource model
	e.ResourceModel = m
*/
package handler
