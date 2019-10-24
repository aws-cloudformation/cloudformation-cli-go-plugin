/*
Package cfn defines the Handler interface that must be implemented
by a resource provider and the Start function that invokes a Handler.

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

You must implement every method of the handler.
See the handler package for documentation on the Request and ProgressEvent types.

When creating your resource provider, you must also include a main package
and function that invokes cfn.Start, passing a pointer to your handler.

	func main() {
		cfn.Start(&MyHandler{})
	}
*/
package cfn
