package readhandler

import (
	"github.com/aws-cloudformation-rpdk-go-plugin/internal/platform/proxy"
)

type ReadHandler struct {
	model proxy.Model
}

func New(m proxy.Model) ReadHandler {
	return ReadHandler{
		model: m,
	}

}

func (r ReadHandler) HandleRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	// TODO : put your code here

	//Set the model properties
	r.model.Property1 = "Hello"
	r.model.Property2 = 123

	p := proxy.ProgressEvent{}

	p.ResourceModel = r.model

	p.ProgressStatus = proxy.Complete

	p.Message = "Resource created!"

	return &p, nil

}
