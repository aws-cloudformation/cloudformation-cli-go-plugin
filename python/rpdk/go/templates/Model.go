package resource

import (
	"fmt"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy"
)


type {{ model_name|uppercase_first_letter }} struct {  
    {% for name, type in properties.items() %}
    {{ name|uppercase_first_letter }}          {{ type }} `json:"{{ name }}"`
        {% endfor %}
}

func (s {{ model_name|uppercase_first_letter }}) CreateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	r := request.DesiredResourceState.({{ model_name|uppercase_first_letter }})
	p := request.PreviousResourceState.({{ model_name|uppercase_first_letter }})

	fmt.Println(r)
	fmt.Println(p)
	return &proxy.ProgressEvent{
		ProgressStatus:       proxy.Complete,
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

func (s {{ model_name|uppercase_first_letter }}) DeleteRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return nil, nil
}

func (s {{ model_name|uppercase_first_letter }}) ListRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return nil, nil
}

func (s {{ model_name|uppercase_first_letter }}) ReadRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return nil, nil
}

func (s {{ model_name|uppercase_first_letter }}) UpdateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	return nil, nil
}
