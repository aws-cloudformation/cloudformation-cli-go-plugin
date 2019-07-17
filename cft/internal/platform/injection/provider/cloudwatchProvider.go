package provider

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// CloudWatchProvider is a set of credentials which are set by the lambda request,
type CloudWatchProvider struct {
	creds  credentials.Provider
	region string
}

//NewCloudWatchProvider is a factory function that returns a new CloudWatchProvider.
func NewCloudWatchProvider(credentialsProvider credentials.Provider) *CloudWatchProvider {

	return &CloudWatchProvider{
		creds: credentialsProvider,
	}
}

//Get returns a new CloudWatch service sesson.
func (c *CloudWatchProvider) Get() (*cloudwatch.CloudWatch, error) {

	// Default Retry Condition of Retry Policy retries on Throttling and ClockSkew
	// Exceptions.
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(c.region),
		Credentials: credentials.NewCredentials(c.creds),
		MaxRetries:  aws.Int(16),
	})

	if err != nil {
		return nil, err
	}

	return cloudwatch.New(sess), nil
}
