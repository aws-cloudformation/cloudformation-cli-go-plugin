package callback

import (
	"context"
	"errors"
	"testing"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/logging"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

var MockModel = []byte("{\"foo\": \"bar\"}")

type MockedCallback struct {
	errCount int
}

func NewMockedCallback(errCount int) *MockedCallback {
	return &MockedCallback{errCount: errCount}
}

func (m *MockedCallback) RecordHandlerProgress(ctx context.Context, in *cloudformation.RecordHandlerProgressInput, optFns ...func(*cloudformation.Options)) (*cloudformation.RecordHandlerProgressOutput, error) {
	if m.errCount > 0 {
		m.errCount--
		return nil, errors.New("error")
	}
	return nil, nil
}

func TestTranslateOperationStatus(t *testing.T) {
	tests := []struct {
		name string
		in   Status
		want string
	}{
		{"SUCCESS", Success, string(cftypes.OperationStatusSuccess)},
		{"FAILED", Failed, string(cftypes.OperationStatusFailed)},
		{"IN_PROGRESS", InProgress, string(cftypes.OperationStatusInProgress)},
		{"PENDING", Pending, string(cftypes.OperationStatusPending)},
		{"Unknown", "Foo", string(cftypes.OperationStatusFailed)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TranslateOperationStatus(tt.in); got != tt.want {
				t.Errorf("TranslateOperationStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranslateErrorCode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"NotUpdatable", "NotUpdatable", string(cftypes.HandlerErrorCodeNotUpdatable)},
		{"InvalidRequest", "InvalidRequest", string(cftypes.HandlerErrorCodeInvalidRequest)},
		{"AccessDenied", "AccessDenied", string(cftypes.HandlerErrorCodeAccessDenied)},
		{"InvalidCredentials", "InvalidCredentials", string(cftypes.HandlerErrorCodeInvalidCredentials)},
		{"AlreadyExists", "AlreadyExists", string(cftypes.HandlerErrorCodeAlreadyExists)},
		{"NotFound", "NotFound", string(cftypes.HandlerErrorCodeNotFound)},
		{"ResourceConflict", "ResourceConflict", string(cftypes.HandlerErrorCodeResourceConflict)},
		{"Throttling", "Throttling", string(cftypes.HandlerErrorCodeThrottling)},
		{"ServiceLimitExceeded", "ServiceLimitExceeded", string(cftypes.HandlerErrorCodeServiceLimitExceeded)},
		{"GeneralServiceException", "GeneralServiceException", string(cftypes.HandlerErrorCodeGeneralServiceException)},
		{"ServiceInternalError", "ServiceInternalError", string(cftypes.HandlerErrorCodeServiceInternalError)},
		{"NetworkFailure", "NetworkFailure", string(cftypes.HandlerErrorCodeNetworkFailure)},
		{"InternalFailure", "InternalFailure", string(cftypes.HandlerErrorCodeInternalFailure)},
		{"Unknown", "foo", string(cftypes.HandlerErrorCodeInternalFailure)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TranslateErrorCode(tt.in); got != tt.want {
				t.Errorf("TranslateErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCloudFormationCallbackAdapterReportProgress(t *testing.T) {
	c := &CloudFormationCallbackAdapter{
		client:      NewMockedCallback(0),
		logger:      logging.New("callback: "),
		bearerToken: "123456",
	}
	if err := c.reportProgress("AccessDenied", Failed, InProgress, MockModel, "retry"); err != nil {
		t.Errorf("reportProgress() error = %v", err)
	}
}
