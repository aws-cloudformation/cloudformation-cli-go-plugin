package resource

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Create handles the Create event from the Cloudformation service.
func Create(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	var err error
	// Create S3 service client
	svc := s3.New(req.Session)
	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(*currentModel.BucketName),
	})

	// We catch all errors. But if the handler panics, it is caught and a error progress event is returned.
	if err != nil {
		// Construct a new handler.ProgressEvent and return it as an error
		response := handler.ProgressEvent{
			OperationStatus: handler.Failed,

			HandlerErrorCode: cloudformation.HandlerErrorCodeInvalidRequest,
			Message:          fmt.Sprintf("Unable to create bucket %q, %v", *currentModel.BucketName, err),
			ResourceModel:    currentModel,
		}
		return response, nil
	}

	// Wait until bucket is created before finishing
	fmt.Printf("Waiting for bucket %q to be created...\n", *currentModel.BucketName)

	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(*currentModel.BucketName),
	})

	if err != nil {
		response := handler.ProgressEvent{
			OperationStatus:  handler.Failed,
			HandlerErrorCode: cloudformation.HandlerErrorCodeInvalidRequest,
			Message:          fmt.Sprintf("Unable to create bucket %q, %v", *currentModel.BucketName, err),
			ResourceModel:    currentModel,
		}
		return response, nil
	}

	// Use the standard log package to output logs to Cloudwatch logs.
	log.Printf("Bucket %q successfully created\n", *currentModel.BucketName)

	// We need to set the bucket id because it is the primary Identifier
	currentModel.Id = currentModel.BucketName

	// We return success
	response := handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Create complete",
		ResourceModel:   currentModel,
	}
	return response, nil
}

// Read handles the Read event from the Cloudformation service.
func Read(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	// Noop
	return handler.ProgressEvent{}, errors.New("Not implemented: Read")
}

// Update handles the Update event from the Cloudformation service.
func Update(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	// In this example, we do not replace the bucket.
	// So we just return success if the handler is called.

	// Construct a new handler.ProgressEvent and return it
	response := handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Update complete",
		ResourceModel:   currentModel,
	}

	return response, nil
}

// Delete handles the Delete event from the Cloudformation service.
func Delete(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	var err error
	// Create S3 service client
	svc := s3.New(req.Session)

	// Delete the S3 Bucket
	// It must be empty or else the call fails
	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(*currentModel.Id),
	})
	if err != nil {
		response := handler.ProgressEvent{
			OperationStatus:  handler.Failed,
			HandlerErrorCode: cloudformation.HandlerErrorCodeInvalidRequest,
			Message:          fmt.Sprintf("Unable to delete bucket %q, %v", *currentModel.BucketName, err),
			ResourceModel:    currentModel,
		}
		return response, nil
	}

	// Wait until bucket is deleted before finishing
	fmt.Printf("Waiting for bucket %q to be deleted...\n", *currentModel.BucketName)

	err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: aws.String(*currentModel.BucketName),
	})
	if err != nil {
		response := handler.ProgressEvent{
			OperationStatus:  handler.Failed,
			HandlerErrorCode: cloudformation.HandlerErrorCodeInvalidRequest,
			Message:          fmt.Sprintf("Error occurred while waiting for %q to be deleted: %v", *currentModel.BucketName, err),
			ResourceModel:    currentModel,
		}
		return response, nil
	}

	// Use the standard log package to output logs to Cloudwatch logs.
	log.Printf("Bucket %q successfully deleted\n", *currentModel.BucketName)

	// We return success
	response := handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Create complete",
		ResourceModel:   currentModel,
	}
	return response, nil
}

// List handles the List event from the Cloudformation service.
func List(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	// Noop
	return handler.ProgressEvent{}, errors.New("Not implemented: Read")
}
