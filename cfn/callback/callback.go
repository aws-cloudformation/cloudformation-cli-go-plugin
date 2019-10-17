package callback

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/avast/retry-go"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/errcode"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/operationstatus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

const (
	//ServiceInternalError ...
	ServiceInternalError string = "ServiceInternal"
	//MaxRetries is the number of retries allowed to report status.
	MaxRetries uint = 3
)

//CloudFormationCallbackAdapter used to report progress events back to CloudFormation.
type CloudFormationCallbackAdapter struct {
	client cloudformationiface.CloudFormationAPI
}

//New creates a CloudFormationCallbackAdapter and returns a pointer to the struct.
func New(client cloudformationiface.CloudFormationAPI) *CloudFormationCallbackAdapter {
	return &CloudFormationCallbackAdapter{
		client: client,
	}
}

//ReportProgress reports the current status back to the Cloudformation service.
func (c *CloudFormationCallbackAdapter) ReportProgress(bearerToken string, code errcode.Status, status operationstatus.Status, resourceModel interface{}, statusMessage string) error {

	b, err := json.Marshal(resourceModel)

	if err != nil {
		return cfnerr.New(ServiceInternalError, "Schedule error", err)
	}

	in := cloudformation.RecordHandlerProgressInput{
		BearerToken:     aws.String(bearerToken),
		OperationStatus: aws.String(TranslateOperationStatus(status.String())),
		StatusMessage:   aws.String(statusMessage),
		ResourceModel:   aws.String(string(b)),
		ErrorCode:       aws.String(TranslateErrorCode(code.String())),
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
			log.Println(s)

		}), retry.Attempts(MaxRetries),
	)

	if rerr != nil {
		return cfnerr.New(ServiceInternalError, "Callback Error error", rerr)
	}

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
