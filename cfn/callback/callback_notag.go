// +build !callback

package callback

import (
	"log"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

//CloudFormationCallbackAdapter used to report progress events back to CloudFormation.
type CloudFormationCallbackAdapter struct {
	logger *log.Logger
	client cloudformationiface.CloudFormationAPI
}

//New creates a CloudFormationCallbackAdapter and returns a pointer to the struct.
func New(client cloudformationiface.CloudFormationAPI) *CloudFormationCallbackAdapter {
	return &CloudFormationCallbackAdapter{
		logger: logging.New("callback: "),
		client: client,
	}
}

//ReportProgress reports the current status back to the Cloudformation service.
func (c *CloudFormationCallbackAdapter) ReportProgress(bearerToken string, code string, operationStatus string, currentOperationStatus string, resourceModel string, statusMessage string) error {

	in := cloudformation.RecordHandlerProgressInput{
		BearerToken:     aws.String(bearerToken),
		OperationStatus: aws.String(TranslateOperationStatus(operationStatus)),
	}

	if len(statusMessage) != 0 {
		in.SetStatusMessage(statusMessage)
	}

	if len(resourceModel) != 0 {
		in.SetResourceModel(resourceModel)
	}

	if len(code) != 0 {
		in.SetErrorCode(TranslateErrorCode(code))
	}

	if len(currentOperationStatus) != 0 {
		in.SetCurrentOperationStatus(currentOperationStatus)
	}

	c.logger.Printf("Record progress: %v", &in)

	return nil
}

//TranslateErrorCode : Translate the error code into a standard Cloudformation error
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

//TranslateOperationStatus Translate the operation Status into a standard Cloudformation error
func TranslateOperationStatus(operationStatus string) string {

	switch operationStatus {
	case "SUCCESS":
		return cloudformation.OperationStatusSuccess
	case "FAILED":
		return cloudformation.OperationStatusFailed
	case "IN_PROGRESS":
		return cloudformation.OperationStatusInProgress
	default:
		// default will be to fail on unknown status
		return cloudformation.OperationStatusFailed
	}

}
