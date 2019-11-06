package resource

import (
	"bytes"
	"context"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var sess *session.Session
var client *s3.S3

func init() {
	sess = session.Must(session.NewSession())
	client = s3.New(sess)
}

// Handler implements the cfn.Handler interface.
// The zero value is ready to use.
type Handler struct {
}

// Create handles the Create event from the Cloudformation service.
func (r *Handler) Create(ctx context.Context, req handler.Request) handler.ProgressEvent {
	//***Add code here: Make your API call, modify the model, etc..
	m := &Model{}
	if err := req.Unmarshal(m); err != nil {
		cfnErr := cfnerr.New(cfnerr.GeneralServiceException, "Unable to complete request", err)
		return handler.NewFailedEvent(cfnErr)
	}

	// Create the object
	_, err := client.PutObject(&s3.PutObjectInput{
		ACL:    m.ACL.Value(),
		Body:   bytes.NewReader([]byte(*m.Content.Value())),
		Bucket: m.BucketName.Value(),
		Key:    m.Key.Value(),
	})

	if err != nil {
		return handler.NewFailedEvent(cfnerr.New(
			cfnerr.GeneralServiceException,
			"Unable to create object: "+err.Error(),
			err,
		))
	}

	p := handler.NewProgressEvent()
	p.ResourceModel = m
	p.OperationStatus = handler.Success
	p.Message = "Completed"

	// return the status
	return p
}

// Read handles the Read event from the Cloudformation service.
func (r *Handler) Read(ctx context.Context, req handler.Request) handler.ProgressEvent {
	//***Add code here: Make your API call, modify the model, etc..
	m := &Model{}
	if err := req.Unmarshal(m); err != nil {
		cfnErr := cfnerr.New(cfnerr.GeneralServiceException, "Unable to complete request", err)
		return handler.NewFailedEvent(cfnErr)
	}

	p := handler.NewProgressEvent()
	p.ResourceModel = m
	p.OperationStatus = handler.Success
	p.Message = "Completed"

	// return the status
	return p
}

// Update handles the Update event from the Cloudformation service.
func (r *Handler) Update(ctx context.Context, req handler.Request) handler.ProgressEvent {
	//***Add code here: Make your API call, modify the model, etc..
	m := &Model{}
	if err := req.Unmarshal(m); err != nil {
		cfnErr := cfnerr.New(cfnerr.GeneralServiceException, "Unable to complete request", err)
		return handler.NewFailedEvent(cfnErr)
	}

	// Create the object
	_, err := client.PutObject(&s3.PutObjectInput{
		ACL:    m.ACL.Value(),
		Body:   bytes.NewReader([]byte(*m.Content.Value())),
		Bucket: m.BucketName.Value(),
		Key:    m.Key.Value(),
	})

	if err != nil {
		return handler.NewFailedEvent(cfnerr.New(
			cfnerr.GeneralServiceException,
			"Unable to create object: "+err.Error(),
			err,
		))
	}

	p := handler.NewProgressEvent()
	p.ResourceModel = m
	p.OperationStatus = handler.Success
	p.Message = "Completed"

	// return the status
	return p
}

// Delete handles the Delete event from the Cloudformation service.
func (r *Handler) Delete(ctx context.Context, req handler.Request) handler.ProgressEvent {
	//***Add code here: Make your API call, modify the model, etc..
	m := &Model{}
	if err := req.Unmarshal(m); err != nil {
		cfnErr := cfnerr.New(cfnerr.GeneralServiceException, "Unable to complete request", err)
		return handler.NewFailedEvent(cfnErr)
	}

	// Create the object
	_, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: m.BucketName.Value(),
		Key:    m.Key.Value(),
	})

	if err != nil {
		return handler.NewFailedEvent(cfnerr.New(
			cfnerr.GeneralServiceException,
			"Unable to create object: "+err.Error(),
			err,
		))
	}

	p := handler.NewProgressEvent()
	p.ResourceModel = m
	p.OperationStatus = handler.Success
	p.Message = "Completed"

	// return the status
	return p
}

// List handles the List event from the Cloudformation service.
func (r *Handler) List(ctx context.Context, req handler.Request) handler.ProgressEvent {
	//***Add code here: Make your API call, modify the model, etc..
	m := &Model{}
	if err := req.Unmarshal(m); err != nil {
		cfnErr := cfnerr.New(cfnerr.GeneralServiceException, "Unable to complete request", err)
		return handler.NewFailedEvent(cfnErr)
	}

	p := handler.NewProgressEvent()
	p.ResourceModel = m
	p.OperationStatus = handler.Success
	p.Message = "Completed"

	// return the status
	return p
}
