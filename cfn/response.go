package cfn

import (
	"fmt"
	"reflect"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
)

// response represents a response to the
// cloudformation service from a resource handler.
// The zero value is ready to use.
type response struct {
	Message         string         `json:"message,omitempty"`
	OperationStatus handler.Status `json:"operationStatus,omitempty"`
	ResourceModel   interface{}    `json:"resourceModel,omitempty"`
	ErrorCode       cfnerr.Error   `json:"errorCode,omitempty"`
	BearerToken     string         `json:"bearerToken,omitempty"`
}

// newFailedResponse returns a response pre-filled with the supplied error
func newFailedResponse(err error, bearerToken string) response {
	return response{
		OperationStatus: handler.Failed,
		ErrorCode:       cfnerr.New(cfnerr.InternalFailure, "Unpexected error", err),
		Message:         err.Error(),
		BearerToken:     bearerToken,
	}
}

// newResponse converts a progress event into a useable reponse
// for the CloudFormation Resource Provider service to understand.
func newResponse(pevt *handler.ProgressEvent, bearerToken string) (response, error) {
	resp := response{
		OperationStatus: pevt.OperationStatus,
		Message:         pevt.Message,
		BearerToken:     bearerToken,
	}

	if pevt.HandlerErrorCode == "" {
		resp.ErrorCode = cfnerr.New(pevt.HandlerErrorCode, pevt.Message, nil)
	}

	resp.ResourceModel = stringifyModel(pevt.ResourceModel)

	fmt.Printf("RESP: %#v\n", resp)

	return resp, nil
}

func stringifyModel(in interface{}) interface{} {
	val := reflect.ValueOf(in)

	switch val.Kind() {
	case reflect.Struct:
		t := val.Type()
		out := make(map[string]interface{})
		for i := 0; i < val.NumField(); i++ {
			out[t.Field(i).Name] = stringifyModel(val.Field(i).Interface())
		}
		return out
	case reflect.Array, reflect.Slice:
		out := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			out[i] = stringifyModel(val.Index(i).Interface())
		}
		return out
	case reflect.Map:
		out := make(map[string]interface{})
		for _, key := range val.MapKeys() {
			out[key.String()] = stringifyModel(val.MapIndex(key).Interface())
		}
		return out
	default:
		return fmt.Sprint(val.Interface())
	}
}
