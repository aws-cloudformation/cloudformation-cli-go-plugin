package cfn

import (
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/encoding"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// response represents a response to the
// cloudformation service from a resource handler.
type response struct {
	Message              string                 `json:"message,omitempty"`
	OperationStatus      handler.Status         `json:"status,omitempty"`
	ResourceModel        interface{}            `json:"resourceModel,omitempty"`
	ErrorCode            string                 `json:"errorCode,omitempty"`
	BearerToken          string                 `json:"bearerToken,omitempty"`
	ResourceModels       []interface{}          `json:"resourceModels"`
	NextToken            string                 `json:"nextToken,omitempty"`
	CallbackContext      map[string]interface{} `json:"callbackContext,omitempty"`
	CallbackDelaySeconds int64                  `json:"callbackDelaySeconds,omitempty"`
}

func newFailedResponse(err error, bearerToken string) response {
	return response{
		OperationStatus: handler.Failed,
		ErrorCode:       string(cftypes.HandlerErrorCodeInternalFailure),
		Message:         err.Error(),
		BearerToken:     bearerToken,
	}
}

func newResponse(pevt *handler.ProgressEvent, bearerToken string) (response, error) {
	model, err := encoding.Stringify(pevt.ResourceModel)
	if err != nil {
		return response{}, err
	}

	var models []interface{}
	if pevt.ResourceModels != nil {
		models = make([]interface{}, len(pevt.ResourceModels))
		for i := range pevt.ResourceModels {
			m, err := encoding.Stringify(pevt.ResourceModels[i])
			if err != nil {
				return response{}, err
			}
			models[i] = m
		}
	}

	resp := response{
		BearerToken:          bearerToken,
		Message:              pevt.Message,
		OperationStatus:      pevt.OperationStatus,
		ResourceModel:        model,
		ResourceModels:       models,
		NextToken:            pevt.NextToken,
		CallbackContext:      pevt.CallbackContext,
		CallbackDelaySeconds: pevt.CallbackDelaySeconds,
	}

	if pevt.HandlerErrorCode != "" {
		resp.ErrorCode = pevt.HandlerErrorCode
	}

	return resp, nil
}
