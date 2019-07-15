package injection

import "net/url"

type CloudFormationProvider struct {
	callbackEndpoint url.URL
}

func (c *CloudFormationProvider) New() CloudFormationProvider {
	
}


