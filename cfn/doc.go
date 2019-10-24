/*
Package cfn defines the Handler interface that must be implemented
by all resource provider and the Start function that invokes a handler.

An empty example Handler would look like this:

	type MyHandler struct{}

	func (m *MyHandler) Create(ctx context.Context, req handler.Request) handler.ProgressEvent {
		return handler.NewProgressEvent(req)
	}

	func (m *MyHandler) Read(ctx context.Context, req handler.Request) handler.ProgressEvent {
		return handler.NewProgressEvent(req)
	}

	func (m *MyHandler) Update(ctx context.Context, req handler.Request) handler.ProgressEvent {
		return handler.NewProgressEvent(req)
	}

	func (m *MyHandler) Delete(ctx context.Context, req handler.Request) handler.ProgressEvent {
		return handler.NewProgressEvent(req)
	}

	func (m *MyHandler) List(ctx context.Context, req handler.Request) handler.ProgressEvent {
		return handler.NewProgressEvent(req)
	}

You need to then implement code for each method of the handler.
See the handler package for documentation on the Request and ProgressEvent types.

When creating your resource provider, you must also include a main package
and function that invokes cfn.Start and passes a pointer to your handler.

	func main() {
		cfn.Start(&MyHandler{})
	}
*/
package cfn
