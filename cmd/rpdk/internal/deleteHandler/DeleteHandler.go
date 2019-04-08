package deletehandler

import (
	"github.com/aws-cloudformation-rpdk-go-plugin/internal/platform/proxy"
)

type DeleteHandler struct {
	model proxy.Model
}

func New(m proxy.Model) DeleteHandler {
	return DeleteHandler{
		model: m,
	}

}

func (d DeleteHandler) HandleRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	// TODO : put your code here

	//Set the model properties
	d.model.Property1 = "Hello"
	d.model.Property2 = 123

	p := proxy.ProgressEvent{}

	p.ResourceModel = d.model

	p.ProgressStatus = proxy.Complete

	p.Message = "Resource created!"

	return &p, nil

}
