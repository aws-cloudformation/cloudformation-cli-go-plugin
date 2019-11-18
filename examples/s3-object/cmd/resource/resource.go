package resource

import (
	"bytes"
	"errors"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Create handles the Create event from the Cloudformation service.
func Create(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	// Create the object
	_, err := s3.New(req.Session).PutObject(&s3.PutObjectInput{
		ACL:    currentModel.ACL.Value(),
		Body:   bytes.NewReader([]byte(*currentModel.Content.Value())),
		Bucket: currentModel.BucketName.Value(),
		Key:    currentModel.Key.Value(),
	})

	if err != nil {
		return handler.ProgressEvent{}, err
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Create complete",
		ResourceModel:   currentModel,
	}, nil
}

// Read handles the Read event from the Cloudformation service.
func Read(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	return handler.ProgressEvent{}, errors.New("Not implemented: Read")
}

// Update handles the Update event from the Cloudformation service.
func Update(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	_, err := Delete(req, prevModel, prevModel)
	if err != nil {
		return handler.ProgressEvent{}, err
	}

	_, err = Create(req, prevModel, currentModel)
	if err != nil {
		return handler.ProgressEvent{}, err
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Update complete",
		ResourceModel:   currentModel,
	}, nil
}

// Delete handles the Delete event from the Cloudformation service.
func Delete(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	// Delete the object
	_, err := s3.New(req.Session).DeleteObject(&s3.DeleteObjectInput{
		Bucket: currentModel.BucketName.Value(),
		Key:    currentModel.Key.Value(),
	})

	if err != nil {
		return handler.ProgressEvent{}, err
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Delete complete",
	}, nil
}

// List handles the List event from the Cloudformation service.
func List(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	return handler.ProgressEvent{}, errors.New("Not implemented: List")
}
