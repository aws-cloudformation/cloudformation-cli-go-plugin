package createhandler

import (
	"github.com/aws-cloudformation-rpdk-go-plugin/internal/platform/proxy"
)

type CreateHandler struct {
	model proxy.Model
}

func New(m proxy.Model) CreateHandler {
	return CreateHandler{
		model: m,
	}

}

func (c CreateHandler) HandleRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	// TODO : put your code here

	//Set the model properties
	c.model.Property1 = "Hello"
	c.model.Property2 = 123

	p := proxy.ProgressEvent{}

	p.ResourceModel = c.model

	p.ProgressStatus = proxy.Complete

	p.Message = "Resource created!"

	return &p, nil

}
