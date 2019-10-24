/*
Package handler contains types that are passed into and out of the methods
of a resource provider's implementation of cfn.Handler

Everything method of a handler receives a handler.Request and must return
a handler.ProgressEvent. An empty example Update method that returns
a blank ProgressEvent could look like this:

	func (h *MyHandler) Update(ctx context.Context, req handler.Request) handler.ProgressEvent {
		return handler.NewProgressEvent(req)
	}

The handler.Request can be used to access information about an existing resource:

	func (h *MyHandler) Update(ctx context.Context, req handler.Request) handler.ProgressEvent {
		var m := &MyModel{}

		err := req.Unmarshal(m)
		if err != nil {
			cfnErr := cfnerr.New(cfnerr.GeneralServiceException, "Unable to complete request", err)
			return handler.NewFailedEvent(req, cfnErr)
		}

		fmt.Println("Updating resource:", req.LogicalResourceID)
		return handler.NewProgressEvent(req)
	}

When a Handler method is successful, it should set the appropriate properties
of the ProgressEvent that it returns.

For example, an "in progress" response might look like this:

	e := handler.NewProgressEvent(req)
	e.OperationStatus = InProgress
	e.Message = "Creating resource " + req.LogicalResourceID

And a completed response like this:

	e := handler.NewProgressEvent(req)
	e.OperationStatus = Success
	e.Message = "Created resource " + req.LogicalResourceID
	e.ResourceModel = m  // Model populated from req.Unmarshal and modified
*/
package handler
