package resource

import (
	"context"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
)

type Handler struct{}

func (r *Handler) Create(ctx context.Context, req handler.Request) handler.ProgressEvent {
	return wrap(ctx, req, createResource)
}

func (r *Handler) Read(ctx context.Context, req handler.Request) handler.ProgressEvent {
	return wrap(ctx, req, readResource)
}

func (r *Handler) Update(ctx context.Context, req handler.Request) handler.ProgressEvent {
	return wrap(ctx, req, updateResource)
}

func (r *Handler) Delete(ctx context.Context, req handler.Request) handler.ProgressEvent {
	return wrap(ctx, req, deleteResource)
}

// List handles the List event from the Cloudformation service.
func (r *Handler) List(ctx context.Context, req handler.Request) handler.ProgressEvent {
	return handler.NewFailedEvent(cfnerr.New(
		cfnerr.GeneralServiceException,
		"Not implemented",
		nil,
	))
}
