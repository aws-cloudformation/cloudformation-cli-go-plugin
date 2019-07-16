package callback

import (
	"encoding/json"
	"log"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cft/internal/platform/injection/provider"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

//CloudFormationCallbackAdapter used to report progress events back to CloudFormation.
type CloudFormationCallbackAdapter struct {
	Client     cloudformationiface.CloudFormationAPI
	cfProvider provider.CloudFormationProvider
}

//New creates a CloudFormationCallbackAdapter and returns a pointer to the struct.
func New(cloudFormationProvider provider.CloudFormationProvider) *CloudFormationCallbackAdapter {
	return &CloudFormationCallbackAdapter{
		cfProvider: cloudFormationProvider,
	}
}

func (c *CloudFormationCallbackAdapter) RefreshClient() error {

	p, err := c.cfProvider.Get()

	if err != nil {
		return err
	}
	c.Client = p

	return nil
}

//ReportProgress reports the current status back to the Cloudformation service
func (c *CloudFormationCallbackAdapter) ReportProgress(bearerToken string, errorCode, operationStatus string, resourceModel interface{}, statusMessage string) {

	b, err := json.Marshal(resourceModel)

	if err != nil {
		//log.Fatalf("error:  %s", err)
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
