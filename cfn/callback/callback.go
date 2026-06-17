/*
Package callback provides functions for creating resource providers
that may need to be called multiple times while waiting
for resources to settle.
*/
package callback

import (
	"context"
	"log"

	"github.com/avast/retry-go"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

const (
	ServiceInternalError string = "ServiceInternal"
	MaxRetries           uint   = 3
)

// CloudFormationClient is the subset of the CloudFormation API used by this package.
type CloudFormationClient interface {
	RecordHandlerProgress(ctx context.Context, params *cloudformation.RecordHandlerProgressInput, optFns ...func(*cloudformation.Options)) (*cloudformation.RecordHandlerProgressOutput, error)
}

// CloudFormationCallbackAdapter used to report progress events back to CloudFormation.
type CloudFormationCallbackAdapter struct {
	client      CloudFormationClient
	bearerToken string
	logger      *log.Logger
}

// New creates a CloudFormationCallbackAdapter and returns a pointer to the struct.
func New(client CloudFormationClient, bearerToken string) *CloudFormationCallbackAdapter {
	return &CloudFormationCallbackAdapter{
		client:      client,
		bearerToken: bearerToken,
		logger:      logging.New("callback"),
	}
}

// ReportStatus reports the status back to the CloudFormation service.
func (c *CloudFormationCallbackAdapter) ReportStatus(operationStatus Status, model []byte, message string, errCode string) error {
	return c.reportProgress(errCode, operationStatus, InProgress, model, message)
}

// ReportInitialStatus reports the initial status back to the CloudFormation service.
func (c *CloudFormationCallbackAdapter) ReportInitialStatus() error {
	return c.reportProgress("", InProgress, Pending, []byte(""), "")
}

// ReportFailureStatus reports the failure status back to the CloudFormation service.
func (c *CloudFormationCallbackAdapter) ReportFailureStatus(model []byte, errCode string, handlerError error) error {
	return c.reportProgress(errCode, Failed, InProgress, model, handlerError.Error())
}

func (c *CloudFormationCallbackAdapter) reportProgress(errCode string, operationStatus Status, currentOperationStatus Status, resourceModel []byte, statusMessage string) error {
	in := &cloudformation.RecordHandlerProgressInput{
		BearerToken:     aws.String(c.bearerToken),
		OperationStatus: cftypes.OperationStatus(TranslateOperationStatus(operationStatus)),
	}

	if len(statusMessage) != 0 {
		in.StatusMessage = aws.String(statusMessage)
	}

	if len(resourceModel) != 0 {
		in.ResourceModel = aws.String(string(resourceModel))
	}

	if len(errCode) != 0 {
		in.ErrorCode = cftypes.HandlerErrorCode(TranslateErrorCode(errCode))
	}

	if len(currentOperationStatus) != 0 {
		in.CurrentOperationStatus = cftypes.OperationStatus(TranslateOperationStatus(currentOperationStatus))
	}

	rerr := retry.Do(
		func() error {
			_, err := c.client.RecordHandlerProgress(context.Background(), in)
			return err
		},
		retry.OnRetry(func(n uint, err error) {
			c.logger.Printf("Failed to record progress: try:#%d: %s\n", n+1, err)
		}),
		retry.Attempts(MaxRetries),
	)

	if rerr != nil {
		return cfnerr.New(ServiceInternalError, "Callback ReportProgress Error", rerr)
	}

	return nil
}

// TranslateErrorCode translates an error code into a standard CloudFormation error.
func TranslateErrorCode(errorCode string) string {
	switch cftypes.HandlerErrorCode(errorCode) {
	case cftypes.HandlerErrorCodeNotUpdatable,
		cftypes.HandlerErrorCodeInvalidRequest,
		cftypes.HandlerErrorCodeAccessDenied,
		cftypes.HandlerErrorCodeInvalidCredentials,
		cftypes.HandlerErrorCodeAlreadyExists,
		cftypes.HandlerErrorCodeNotFound,
		cftypes.HandlerErrorCodeResourceConflict,
		cftypes.HandlerErrorCodeThrottling,
		cftypes.HandlerErrorCodeServiceLimitExceeded,
		cftypes.HandlerErrorCodeServiceTimeout,
		cftypes.HandlerErrorCodeGeneralServiceException,
		cftypes.HandlerErrorCodeServiceInternalError,
		cftypes.HandlerErrorCodeNetworkFailure,
		cftypes.HandlerErrorCodeInternalFailure:
		return errorCode
	default:
		return string(cftypes.HandlerErrorCodeInternalFailure)
	}
}

// TranslateOperationStatus translates an operation status.
func TranslateOperationStatus(operationStatus Status) string {
	switch operationStatus {
	case Success:
		return string(cftypes.OperationStatusSuccess)
	case Failed:
		return string(cftypes.OperationStatusFailed)
	case InProgress:
		return string(cftypes.OperationStatusInProgress)
	case Pending:
		return string(cftypes.OperationStatusPending)
	default:
		return string(cftypes.OperationStatusFailed)
	}
}


