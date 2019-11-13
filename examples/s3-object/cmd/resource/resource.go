package resource

import (
	"bytes"
	"errors"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Create handles the Create event from the Cloudformation service.
func Create(prevModel *Model, currentModel *Model, session *session.Session) (handler.ProgressEvent, error) {
	// Create the object
	_, err := s3.New(session).PutObject(&s3.PutObjectInput{
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
func Read(prevModel *Model, currentModel *Model, session *session.Session) (handler.ProgressEvent, error) {
	return handler.ProgressEvent{}, errors.New("Not implemented: Read")
}

// Update handles the Update event from the Cloudformation service.
func Update(prevModel *Model, currentModel *Model, session *session.Session) (handler.ProgressEvent, error) {
	_, err := Delete(prevModel, prevModel, session)
	if err != nil {
		return handler.ProgressEvent{}, err
	}

	_, err = Create(prevModel, currentModel, session)
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
func Delete(prevModel *Model, currentModel *Model, session *session.Session) (handler.ProgressEvent, error) {
	// Delete the object
	_, err := s3.New(session).DeleteObject(&s3.DeleteObjectInput{
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
func List(prevModel *Model, currentModel *Model, session *session.Session) (handler.ProgressEvent, error) {
	return handler.ProgressEvent{}, errors.New("Not implemented: List")
}
