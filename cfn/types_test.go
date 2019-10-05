package cfn

type EmptyHandlers struct{}

func (h *EmptyHandlers) Create(request Request, rc *RequestContext) (Response, error) {
	return nil, nil
}

func (h *EmptyHandlers) Read(request Request, rc *RequestContext) (Response, error) {
	return nil, nil
}

func (h *EmptyHandlers) Update(request Request, rc *RequestContext) (Response, error) {
	return nil, nil
}

func (h *EmptyHandlers) Delete(request Request, rc *RequestContext) (Response, error) {
	return nil, nil
}

func (h *EmptyHandlers) List(request Request, rc *RequestContext) (Response, error) {
	return nil, nil
}
