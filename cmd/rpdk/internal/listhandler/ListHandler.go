package listhandler

import (
	"github.com/aws-cloudformation-rpdk-go-plugin/internal/platform/proxy"
)

type ListHandler struct {
	model proxy.Model
}

func New(m proxy.Model) ListHandler {
	return ListHandler{
		model: m,
	}

}

func (l ListHandler) HandleRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	// TODO : put your code here

	//Set the model properties
	l.model.Property1 = "Hello"
	l.model.Property2 = 123

	p := proxy.ProgressEvent{}

	p.ResourceModel = l.model

	p.ProgressStatus = proxy.Complete

	p.Message = "Resource created!"

	return &p, nil

}
