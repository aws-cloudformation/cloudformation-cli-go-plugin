package updatehandler

import (
	"github.com/aws-cloudformation-rpdk-go-plugin/internal/platform/proxy"
)

type UpdateHandler struct {
	model proxy.Model
}

func New(m proxy.Model) UpdateHandler {
	return UpdateHandler{
		model: m,
	}

}

func (u UpdateHandler) HandleRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	// TODO : put your code here

	//Set the model properties
	u.model.Property1 = "Hello"
	u.model.Property2 = 123

	p := proxy.ProgressEvent{}

	p.ResourceModel = u.model

	p.ProgressStatus = proxy.Complete

	p.Message = "Resource created!"

	return &p, nil

}
