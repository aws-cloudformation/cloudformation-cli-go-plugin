package resource

import (
	"bytes"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Create handles the Create event from the Cloudformation service.
func createResource(prevModel, currentModel *Model, session *session.Session) (response handler.ProgressEvent, err error) {
	// Create the object
	_, err = s3.New(session).PutObject(&s3.PutObjectInput{
		ACL:    currentModel.ACL.Value(),
		Body:   bytes.NewReader([]byte(*currentModel.Content.Value())),
		Bucket: currentModel.BucketName.Value(),
		Key:    currentModel.Key.Value(),
	})

	if err == nil {
		response.OperationStatus = handler.Success
		response.Message = "Create complete"
		response.ResourceModel = currentModel
	}

	return
}

// Read handles the Read event from the Cloudformation service.
func readResource(prevModel, currentModel *Model, session *session.Session) (response handler.ProgressEvent, err error) {
	// Noop

	response.OperationStatus = handler.Success
	response.Message = "Read complete"
	response.ResourceModel = currentModel

	return
}

// Update handles the Update event from the Cloudformation service.
func updateResource(prevModel, currentModel *Model, session *session.Session) (response handler.ProgressEvent, err error) {
	_, err = deleteResource(prevModel, prevModel, session)
	if err != nil {
		return response, err
	}

	_, err = createResource(prevModel, currentModel, session)
	if err != nil {
		return response, err
	}

	response.OperationStatus = handler.Success
	response.Message = "Update complete"
	response.ResourceModel = currentModel

	return
}

// Delete handles the Delete event from the Cloudformation service.
func deleteResource(prevModel, currentModel *Model, session *session.Session) (response handler.ProgressEvent, err error) {
	// Delete the object
	_, err = s3.New(session).DeleteObject(&s3.DeleteObjectInput{
		Bucket: currentModel.BucketName.Value(),
		Key:    currentModel.Key.Value(),
	})

	if err == nil {
		response.OperationStatus = handler.Success
		response.Message = "Delete complete"
	}

	return
}
