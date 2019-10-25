/*
Package handler contains types that are passed into and out of the methods
of a resource provider's implementation of cfn.Handler

Every method of a handler receives a handler.Request and must return
a handler.ProgressEvent.

An empty Update method that returns a blank ProgressEvent could look like this:

	func (h *MyHandler) Update(ctx context.Context, req handler.Request) handler.ProgressEvent {
		return handler.NewProgressEvent()
	}

The handler.Request can be used to access information about an existing resource:

	func (h *MyHandler) Update(ctx context.Context, req handler.Request) handler.ProgressEvent {
		var m := &MyModel{}

		// Unmarshal the current properties of the resource
		err := req.Unmarshal(m)
		if err != nil {
			cfnErr := cfnerr.New(cfnerr.GeneralServiceException, "Unable to complete request", err)
			return handler.NewFailedEvent(cfnErr)
		}

		// Output the resource's logical ID
		fmt.Println("Updating resource:", req.LogicalResourceID)
		return handler.NewProgressEvent()
	}

When a Handler method is successful, it should set the appropriate properties
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
