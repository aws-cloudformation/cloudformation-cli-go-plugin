package cfn 

type EmptyHandlers struct{}

func (h *EmptyHandlers) Create(request Request) (Response, error) {
	return nil, nil
}

func (h *EmptyHandlers) Read(request Request) (Response, error) {
	return nil, nil
}

func (h *EmptyHandlers) Update(request Request) (Response, error) {
	return nil, nil
}

func (h *EmptyHandlers) Delete(request Request) (Response, error) {
	return nil, nil
}

func (h *EmptyHandlers) List(request Request) (Response, error) {
	return nil, nil
}
