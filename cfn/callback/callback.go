// +build callback

/*
Package callback provides functions for creating resource providers
that may need to be called multiple times while waiting
for resources to settle.
*/
package callback

import (
	"fmt"
	"log"

	"github.com/avast/retry-go"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

const (
	// ServiceInternalError ...
	ServiceInternalError string = "ServiceInternal"
	// MaxRetries is the number of retries allowed to report status.
	MaxRetries uint = 3
)

// CloudFormationCallbackAdapter used to report progress events back to CloudFormation.
type CloudFormationCallbackAdapter struct {
	client      cloudformationiface.CloudFormationAPI
	bearerToken string
	logger      *log.Logger
}

// New creates a CloudFormationCallbackAdapter and returns a pointer to the struct.
func New(client cloudformationiface.CloudFormationAPI, bearerToken string) *CloudFormationCallbackAdapter {
	return &CloudFormationCallbackAdapter{
		client:      client,
		bearerToken: bearerToken,
		logger:      logging.New("callback"),
	}
}

// ReportStatus reports the status back to the Cloudformation service of a handler
// that has moved from Pending to In_Progress
func (c *CloudFormationCallbackAdapter) ReportStatus(operationStatus Status, model []byte, message string, errCode string) error {
	if err := c.reportProgress(errCode, operationStatus, InProgress, model, message); err != nil {
		return err
	}
	return nil
}

// ReportInitialStatus reports the initial status back to the Cloudformation service.
func (c *CloudFormationCallbackAdapter) ReportInitialStatus() error {
	if err := c.reportProgress("", InProgress, Pending, []byte(""), ""); err != nil {
		return err
	}
	return nil
}

//ReportFailureStatus reports the failure status back to the Cloudformation service.
func (c *CloudFormationCallbackAdapter) ReportFailureStatus(model []byte, errCode string, handlerError error) error {
	if err := c.reportProgress(errCode, Failed, InProgress, model, handlerError.Error()); err != nil {
		return err
	}
	return nil
}

// ReportProgress reports the current status back to the Cloudformation service.
func (c *CloudFormationCallbackAdapter) reportProgress(errCode string, operationStatus Status, currentOperationStatus Status, resourceModel []byte, statusMessage string) error {

	in := cloudformation.RecordHandlerProgressInput{
		BearerToken:     aws.String(c.bearerToken),
		OperationStatus: aws.String(TranslateOperationStatus(operationStatus)),
	}

	if len(statusMessage) != 0 {
		in.SetStatusMessage(statusMessage)
	}

	if len(resourceModel) != 0 {
		in.SetResourceModel(string(resourceModel))
	}

	if len(errCode) != 0 {
		in.SetErrorCode(TranslateErrorCode(errCode))
	}

	if len(currentOperationStatus) != 0 {
		in.SetCurrentOperationStatus(TranslateOperationStatus(currentOperationStatus))
	}

	// Do retries and emit logs.
	rerr := retry.Do(
		func() error {
			_, err := c.client.RecordHandlerProgress(&in)
			if err != nil {
				return err
			}
			return nil
		}, retry.OnRetry(func(n uint, err error) {
			s := fmt.Sprintf("Failed to record progress: try:#%d: %s\n ", n+1, err)
			c.logger.Println(s)

		}), retry.Attempts(MaxRetries),
	)

	if rerr != nil {
		return cfnerr.New(ServiceInternalError, "Callback ReportProgress Error", rerr)
	}

	return nil
}

// TranslateErrorCode : Translate the error code into a standard Cloudformation error
func TranslateErrorCode(errorCode string) string {

	switch errorCode {
	case "NotUpdatable":
		return cloudformation.HandlerErrorCodeNotUpdatable
	case "InvalidRequest":
		return cloudformation.HandlerErrorCodeInvalidRequest
	case "AccessDenied":
		return cloudformation.HandlerErrorCodeAccessDenied
	case "InvalidCredentials":
		return cloudformation.HandlerErrorCodeInvalidCredentials
	case "AlreadyExists":
		return cloudformation.HandlerErrorCodeAlreadyExists
	case "NotFound":
		return cloudformation.HandlerErrorCodeNotFound
	case "ResourceConflict":
		return cloudformation.HandlerErrorCodeResourceConflict
	case "Throttling":
		return cloudformation.HandlerErrorCodeThrottling
	case "ServiceLimitExceeded":
		return cloudformation.HandlerErrorCodeServiceLimitExceeded
	case "NotStabilized":
		return cloudformation.HandlerErrorCodeNotStabilized
	case "GeneralServiceException":
		return cloudformation.HandlerErrorCodeGeneralServiceException
	case "ServiceInternalError":
		return cloudformation.HandlerErrorCodeServiceInternalError
	case "NetworkFailure":
		return cloudformation.HandlerErrorCodeNetworkFailure
	case "InternalFailure":
		return cloudformation.HandlerErrorCodeInternalFailure
	default:
		// InternalFailure is CloudFormation's fallback error code when no more specificity is there
		return cloudformation.HandlerErrorCodeInternalFailure
	}
}

// TranslateOperationStatus Translate the operation Status into a standard Cloudformation error
func TranslateOperationStatus(operationStatus Status) string {

	switch operationStatus {
	case Success:
		return cloudformation.OperationStatusSuccess
	case Failed:
		return cloudformation.OperationStatusFailed
	case InProgress:
		return cloudformation.OperationStatusInProgress
	case Pending:
		return cloudformation.OperationStatusPending
	default:
		// default will be to fail on unknown status
		return cloudformation.OperationStatusFailed
	}

}
