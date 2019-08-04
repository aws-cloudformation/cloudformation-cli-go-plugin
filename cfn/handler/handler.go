package handler

import (
	"encoding/json"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/action"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/operationstatus"
)

// Request ...
type Request struct {
	action          action.Action
	CallbackContext interface{}       `json:"callbackContext"`
	Credentials     map[string]string `json:"credentials"`
	Body            interface{}       `json:"request"`
}

// Action returns the action to be performed
func (r *Request) Action() action.Action {
	return r.action
}

// UnmarshalJSON ...
func (r *Request) UnmarshalJSON(data []byte) error {
	var handlerIntermediate struct {
		Action string
	}

	if err := json.Unmarshal(data, &handlerIntermediate); err != nil {
		return err
	}

	r.action = action.Convert(handlerIntermediate.Action)

	return nil
}

// Response ...
type Response struct {
	Message         string
	OperationStatus operationstatus.Status
	ResourceModel   string
	BearerToken     string
	ErrorCode       error
}
