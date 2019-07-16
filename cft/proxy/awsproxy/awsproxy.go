package awsproxy

// InjectCredentialsAndInvoke consumes a "aws/request.Request" representing the
// client's request for a service action and injects caller credentials. The "output" return
// value will be populated with the request's response once the request completes
// successfully.
//
//
//// This method is useful when you want to inject credentials
// into the SDK's request.
//
//
//    // Example sending a request using the GetBucketReplicationRequest method.
//    req, resp := client.GetBucketReplicationRequest(params)
//    err := Wrapper.InjectCredentialsAndInvoke(req)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (w *Wrapper) InjectCredentialsAndInvoke(req request.Request) error {

	req.Config.Credentials = w.wrapperCreds
	err := req.Send()
	if err != nil {
		return err
	}

	return nil
}