package provider

import "net/url"
import "github.com/aws/aws-sdk-go/aws/session"
import "github.com/aws/aws-sdk-go/aws"
import "github.com/aws/aws-sdk-go/aws/credentials"
import "github.com/aws/aws-sdk-go/service/cloudformation"

// CloudFormationProvider is a set of credentials which are set by the lambda request,
type CloudFormationProvider struct {
	callbackEndpoint *url.URL
	creds            *credentials.Provider
	region           string
}

//NewCloudFormationProvider is a factory function that returns a new CloudFormationProvider.
func NewCloudFormationProvider(credentialsProvider *credentials.Provider) *CloudFormationProvider {

	return &CloudFormationProvider{
		creds: credentialsProvider,
	}
}

//SetCallbackEndpoint sets the call back URL of the CloudFormationProvider.
func (c *CloudFormationProvider) SetCallbackEndpoint(callback *url.URL) {

	c.callbackEndpoint = callback
}

//Get returns a new cloudforamtion service sesson.
func (c *CloudFormationProvider) Get() (*cloudformation.CloudFormation, error) {

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(c.region),
		Credentials: credentials.NewCredentials(*c.creds),
		Endpoint:    aws.String(c.callbackEndpoint.String()),
		MaxRetries:  aws.Int(16),
	})

	if err != nil {
		return nil, err
	}

	return cloudformation.New(sess), nil
}
