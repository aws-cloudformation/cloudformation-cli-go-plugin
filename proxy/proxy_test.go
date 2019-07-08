package proxy

/*
func Test_processInvocationFailedResponse(t *testing.T) {
	re := NewMock(&proxy.ProgressEvent{
		Status:               proxy.Failed,
		HandlerErrorCode:     "Custom Fault",
		Message:              "Custom Fault",
		CallbackDelayMinutes: 0,
	}, nil)
	proxy.StartWithOutLambda(re)

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	readRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/read.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/delete.request.json")
	listRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/list.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type fields struct {
		in proxy.ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		fields                             fields
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"failed CREATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *createRequest, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
		{"failed DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
		{"failed LIST response", fields{proxy.ProcessInvocationInput{mockContext{}, *listRequest, metric.New(New(), listRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
		{"failed READ response", fields{proxy.ProcessInvocationInput{mockContext{}, *readRequest, metric.New(New(), readRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
		{"failed UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Failed,
			HandlerErrorCode:     "Custom Fault",
			Message:              "Custom Fault",
			CallbackDelayMinutes: 0,
		},
			false, 0, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(re)
			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation(&tt.fields.in)
			if got.Status == tt.want.Status {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.Status)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.Status, got.Status)
			}

			if m.HandlerExceptionCount == tt.wantHandlerExceptionCount {
				t.Logf("\t%s\tHandlerException metric should be invoked (%v) times.", succeed, tt.wantHandlerExceptionCount)
			} else {
				t.Errorf("\t%s\tHandlerException metric should be invoked (%v) times : %v", failed, tt.wantHandlerExceptionCount, m.HandlerExceptionCount)
			}

			if m.HandlerInvocationCount == tt.wantHandlerInvocationCount {
				t.Logf("\t%s\tHandlerInvocation metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationCount)
			} else {
				t.Errorf("\t%s\tHandlerInvocation metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationCount, m.HandlerInvocationCount)
			}

			if m.HandlerInvocationDurationCount == tt.wantHandlerInvocationDurationCount {
				t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationDurationCount)
			} else {
				t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationDurationCount, m.HandlerInvocationDurationCount)
			}
			if e.RescheduleAfterMinutesCount == tt.wantrescheduleAfterMinutesCount {
				t.Logf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times.", succeed, tt.wantrescheduleAfterMinutesCount)
			} else {
				t.Errorf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times : %v", failed, tt.wantrescheduleAfterMinutesCount, e.RescheduleAfterMinutesCount)
			}

			if e.CleanupCloudWatchEventsCount == tt.wantcleanupCloudWatchEvents {
				t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
			} else {
				t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
			}
		})
	}
}

func Test_processInvocationCompleteSynchronouslyResponse(t *testing.T) {

	re := NewMock(&proxy.ProgressEvent{
		Status:               proxy.Complete,
		CallbackDelayMinutes: 0,
	}, nil)
	proxy.StartWithOutLambda(re)

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	readRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/read.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/delete.request.json")
	listRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/list.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type fields struct {
		in proxy.ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		fields                             fields
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"complete synchronously CREATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *createRequest, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously READ response", fields{proxy.ProcessInvocationInput{mockContext{}, *readRequest, metric.New(New(), readRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously LIST response", fields{proxy.ProcessInvocationInput{mockContext{}, *listRequest, metric.New(New(), listRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(re)
			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation(&tt.fields.in)
			if got.Status == tt.want.Status {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.Status)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.Status, got.Status)
			}

			if m.HandlerExceptionCount == tt.wantHandlerExceptionCount {
				t.Logf("\t%s\tHandlerException metric should be invoked (%v) times.", succeed, tt.wantHandlerExceptionCount)
			} else {
				t.Errorf("\t%s\tHandlerException metric should be invoked (%v) times : %v", failed, tt.wantHandlerExceptionCount, m.HandlerExceptionCount)
			}

			if m.HandlerInvocationCount == tt.wantHandlerInvocationCount {
				t.Logf("\t%s\tHandlerInvocation metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationCount)
			} else {
				t.Errorf("\t%s\tHandlerInvocation metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationCount, m.HandlerInvocationCount)
			}

			if m.HandlerInvocationDurationCount == tt.wantHandlerInvocationDurationCount {
				t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationDurationCount)
			} else {
				t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationDurationCount, m.HandlerInvocationDurationCount)
			}
			if e.RescheduleAfterMinutesCount == tt.wantrescheduleAfterMinutesCount {
				t.Logf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times.", succeed, tt.wantrescheduleAfterMinutesCount)
			} else {
				t.Errorf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times : %v", failed, tt.wantrescheduleAfterMinutesCount, e.RescheduleAfterMinutesCount)
			}

			if e.CleanupCloudWatchEventsCount == tt.wantcleanupCloudWatchEvents {
				t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
			} else {
				t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
			}
		})
	}
}

func Test_processMalformedSynchronouslyResponse(t *testing.T) {

	re := NewMock(&proxy.ProgressEvent{
		Status:               proxy.Complete,
		CallbackDelayMinutes: 0,
	}, nil)
	proxy.StartWithOutLambda(re)

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/malformed.request.json")
	readRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/malformed.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/malformed.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/malformed.request.json")
	listRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/malformed.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type fields struct {
		in proxy.ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		fields                             fields
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"complete synchronously CREATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *createRequest, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously READ response", fields{proxy.ProcessInvocationInput{mockContext{}, *readRequest, metric.New(New(), readRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},

		{"complete synchronously LIST response", fields{proxy.ProcessInvocationInput{mockContext{}, *listRequest, metric.New(New(), listRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.Complete,
			CallbackDelayMinutes: 0,
		}, false, 0, 1, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(re)
			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation(&tt.fields.in)
			if got.Status == tt.want.Status {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.Status)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.Status, got.Status)
			}

			if m.HandlerExceptionCount == tt.wantHandlerExceptionCount {
				t.Logf("\t%s\tHandlerException metric should be invoked (%v) times.", succeed, tt.wantHandlerExceptionCount)
			} else {
				t.Errorf("\t%s\tHandlerException metric should be invoked (%v) times : %v", failed, tt.wantHandlerExceptionCount, m.HandlerExceptionCount)
			}

			if m.HandlerInvocationCount == tt.wantHandlerInvocationCount {
				t.Logf("\t%s\tHandlerInvocation metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationCount)
			} else {
				t.Errorf("\t%s\tHandlerInvocation metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationCount, m.HandlerInvocationCount)
			}

			if m.HandlerInvocationDurationCount == tt.wantHandlerInvocationDurationCount {
				t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationDurationCount)
			} else {
				t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationDurationCount, m.HandlerInvocationDurationCount)
			}
			if e.RescheduleAfterMinutesCount == tt.wantrescheduleAfterMinutesCount {
				t.Logf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times.", succeed, tt.wantrescheduleAfterMinutesCount)
			} else {
				t.Errorf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times : %v", failed, tt.wantrescheduleAfterMinutesCount, e.RescheduleAfterMinutesCount)
			}

			if e.CleanupCloudWatchEventsCount == tt.wantcleanupCloudWatchEvents {
				t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
			} else {
				t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
			}
		})
	}
}

func Test_processInvocationInProgressWithContextResponse(t *testing.T) {

	re := NewMock(&proxy.ProgressEvent{
		Status:               proxy.InProgress,
		CallbackDelayMinutes: 5,
	}, nil)
	proxy.StartWithOutLambda(re)

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.with-request-context.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.with-request-context.request.json")
	deleteRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/delete.with-request-context.request.json")

	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type fields struct {
		in proxy.ProcessInvocationInput
	}
	tests := []struct {
		name                               string
		fields                             fields
		want                               *proxy.ProgressEvent
		wantErr                            bool
		wantHandlerExceptionCount          int
		wantHandlerInvocationCount         int
		wantHandlerInvocationDurationCount int
		wantrescheduleAfterMinutesCount    int
		wantcleanupCloudWatchEvents        int
	}{
		{"in progress CREATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *createRequest, metric.New(New(), createRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.InProgress,
			CallbackDelayMinutes: 5,
		}, false, 0, 1, 1, 1, 1},

		{"in progress UPDATE response", fields{proxy.ProcessInvocationInput{mockContext{}, *updateRequest, metric.New(New(), updateRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.InProgress,
			CallbackDelayMinutes: 5,
		}, false, 0, 1, 1, 1, 1},

		{"in progress DELETE response", fields{proxy.ProcessInvocationInput{mockContext{}, *deleteRequest, metric.New(New(), deleteRequest.ResourceType), scheduler.New(NewmockedEvents()), nil}}, &proxy.ProgressEvent{
			Status:               proxy.InProgress,
			CallbackDelayMinutes: 5,
			ResourceModel:        deleteRequest.Data.ResourceProperties,
		}, false, 0, 1, 1, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proxy.New(re)
			m := tt.fields.in.Metric.Client.(*MockedMetrics)
			e := tt.fields.in.Sched.Client.(*MockedEvents)
			got := p.ProcessInvocation(&tt.fields.in)
			if got.Status == tt.want.Status {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want.Status)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want.Status, got.Status)
			}

			if m.HandlerExceptionCount == tt.wantHandlerExceptionCount {
				t.Logf("\t%s\tHandlerException metric should be invoked (%v) times.", succeed, tt.wantHandlerExceptionCount)
			} else {
				t.Errorf("\t%s\tHandlerException metric should be invoked (%v) times : %v", failed, tt.wantHandlerExceptionCount, m.HandlerExceptionCount)
			}

			if m.HandlerInvocationCount == tt.wantHandlerInvocationCount {
				t.Logf("\t%s\tHandlerInvocation metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationCount)
			} else {
				t.Errorf("\t%s\tHandlerInvocation metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationCount, m.HandlerInvocationCount)
			}

			if m.HandlerInvocationDurationCount == tt.wantHandlerInvocationDurationCount {
				t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantHandlerInvocationDurationCount)
			} else {
				t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantHandlerInvocationDurationCount, m.HandlerInvocationDurationCount)
			}
			if e.RescheduleAfterMinutesCount == tt.wantrescheduleAfterMinutesCount {
				t.Logf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times.", succeed, tt.wantrescheduleAfterMinutesCount)
			} else {
				t.Errorf("\t%s\tRescheduleAfterMinutesCount method should be invoked (%v) times : %v", failed, tt.wantrescheduleAfterMinutesCount, e.RescheduleAfterMinutesCount)
			}

			if e.CleanupCloudWatchEventsCount == tt.wantcleanupCloudWatchEvents {
				t.Logf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times.", succeed, tt.wantcleanupCloudWatchEvents)
			} else {
				t.Errorf("\t%s\tHandlerInvocationDuration metric should be invoked (%v) times : %v", failed, tt.wantcleanupCloudWatchEvents, e.CleanupCloudWatchEventsCount)
			}
		})
	}
}

func TestTransform(t *testing.T) {

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type args struct {
		r       proxy.HandlerRequest
		handler *proxy.CustomHandler
	}
	tests := []struct {
		name         string
		args         args
		want         *proxy.ResourceHandlerRequest
		wantResource *MockHandler
		wantErr      bool
	}{

		{"Transform CREATE response", args{*createRequest, proxy.New(NewMock(nil, nil))}, &proxy.ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandler{
				MockHandlerResource{"abc", 123},
				MockHandlerResource{},
				nil,
				nil,
			},

			false},

		{"Transform UPDATE response", args{*updateRequest, proxy.New(NewMock(nil, nil))}, &proxy.ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandler{
				MockHandlerResource{"abc", 123},
				MockHandlerResource{"cba", 321},
				nil,
				nil,
			},

			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := proxy.Transform(tt.args.r, tt.args.handler)
			r := tt.args.handler.CustomResource.(*MockHandler)

			if (err != nil) != tt.wantErr {
				t.Errorf("Transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if reflect.DeepEqual(got, tt.want) {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want, got)
			}

			if reflect.DeepEqual(got, tt.want) {
				t.Logf("\t%s\tShould receive a %s status code.", succeed, tt.want)
			} else {
				t.Errorf("\t%s\tShould receive a %s status code : %v", failed, tt.want, got)
			}

			if reflect.DeepEqual(r, tt.wantResource) {
				t.Logf("\t%s\tShould update resource.", succeed)
			} else {
				t.Errorf("\t%s\tShould update resource %v was : %v", failed, tt.wantResource, r)
			}

		})
	}
}

func TestTransformNoDesired(t *testing.T) {

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type args struct {
		r       proxy.HandlerRequest
		handler *proxy.CustomHandler
	}
	tests := []struct {
		name         string
		args         args
		want         *proxy.ResourceHandlerRequest
		wantResource *MockHandlerNoDesired
		wantErr      bool
	}{

		{"Transform CREATE response", args{*createRequest, proxy.New(NewMockNoDesired(nil, nil))}, &proxy.ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandlerNoDesired{},

			false},

		{"Transform UPDATE response", args{*updateRequest, proxy.New(NewMockNoDesired(nil, nil))}, &proxy.ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandlerNoDesired{
				MockHandlerResourceNoDesired{},
				nil,
				nil,
			},

			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy.Transform(tt.args.r, tt.args.handler)

			if err.Error() == "Unable to find DesiredResource in Config object" {
				t.Logf("\t%s\tShould receive a %s error.", succeed, "Unable to find DesiredResource in Config object")
			} else {
				t.Errorf("\t%s\tShould receive a %s error.", failed, "Unable to find DesiredResource in Config object")
			}

		})
	}
}

func TestTransformNoPre(t *testing.T) {

	createRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/create.request.json")
	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	updateRequest, err := loadData(&proxy.HandlerRequest{}, "tests/data/update.request.json")
	if err != nil {
		log.Fatalf("error loading data. :%v", err.Error())
	}

	type args struct {
		r       proxy.HandlerRequest
		handler *proxy.CustomHandler
	}
	tests := []struct {
		name         string
		args         args
		want         *proxy.ResourceHandlerRequest
		wantResource *MockHandlerNoPre
		wantErr      bool
	}{

		{"Transform CREATE response", args{*createRequest, proxy.New(NewMockNoPre(nil, nil))}, &proxy.ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandlerNoPre{},

			false},

		{"Transform UPDATE response", args{*updateRequest, proxy.New(NewMockNoPre(nil, nil))}, &proxy.ResourceHandlerRequest{
			AwsAccountID:        "123456789012",
			Region:              "us-east-1",
			ResourceType:        "AWS::Test::TestModel",
			ResourceTypeVersion: "1.0",
		},
			&MockHandlerNoPre{},

			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy.Transform(tt.args.r, tt.args.handler)

			if err.Error() == "Unable to find PreviousResource in Config object" {
				t.Logf("\t%s\tShould receive a %s error.", succeed, "Unable to find PreviousResource in Config object")
			} else {
				t.Errorf("\t%s\tShould receive a %s error.", failed, "Unable to find PreviousResource in Config object")
			}

		})
	}

}
*/
