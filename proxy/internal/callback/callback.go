package callback

import (
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

//CloudFormationCallbackAdapter used to report progress events back to CloudFormation.
type CloudFormationCallbackAdapter struct {
	Client cloudformationiface.CloudFormationAPI
}

//New creates a CloudFormationCallbackAdapter and returns a pointer to the struct.
func New(client cloudformationiface.CloudFormationAPI) *CloudFormationCallbackAdapter {
	return &CloudFormationCallbackAdapter{
		Client: client,
	}
}


//ReportProgress reports the current status back to the Cloudformation service
func (c *CloudFormationCallbackAdapter) ReportProgress(bearerToken string, errorCode, operationStatus string, resourceModel interface{}, statusMessage string) {

	b, err := json.Marshal(resourceModel)

	if err != nil {
		log.Fatalf("error:  %s", err)
	}

	in := cloudformation.RecordHandlerProgressInput{
		BearerToken:     aws.String(bearerToken),
		OperationStatus: aws.String(TranslateOperationStatus(operationStatus)),
		StatusMessage:   aws.String(statusMessage),
		ResourceModel:   aws.String(string(b)),
	}

	if errorCode != "" {
		in.ErrorCode = aws.String(TranslateErrorCode(errorCode))
	}

	// TODO: be far more fault tolerant, do retries, emit logs and metrics, etc.
	res, err := c.Client.RecordHandlerProgress(&in)

	if err != nil {
		log.Fatalf("error: %s", err)
	}

	log.Printf("Record Handler Progress with %s", res)

}

//TranslateErrorCode : Translate the error code into a standard Cloudformation error
func TranslateErrorCode(errorCode string) string {

	switch errorCode {
	case "AccessDenied":
		return cloudformation.HandlerErrorCodeAccessDenied
	case "InternalFailure":
		return cloudformation.HandlerErrorCodeInternalFailure
	case "InvalidCredentials":
		return cloudformation.HandlerErrorCodeInvalidCredentials
	case "InvalidRequest":
		return cloudformation.HandlerErrorCodeInvalidRequest
	case "NetworkFailure":
		return cloudformation.HandlerErrorCodeNetworkFailure
	case "NoOperationToPerform":
		return cloudformation.HandlerErrorCodeNoOperationToPerform
	case "NotFound":
		return cloudformation.HandlerErrorCodeNotFound
	case "NotReady":
		return cloudformation.HandlerErrorCodeNotReady
	case "NotUpdatable":
		return cloudformation.HandlerErrorCodeNotUpdatable
	case "ServiceException":
		return cloudformation.HandlerErrorCodeServiceException
	case "ServiceLimitExceeded":
		return cloudformation.HandlerErrorCodeServiceLimitExceeded
	case "ServiceTimeout":
		return cloudformation.HandlerErrorCodeServiceTimeout
	case "Throttling":
		return cloudformation.HandlerErrorCodeServiceTimeout
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

