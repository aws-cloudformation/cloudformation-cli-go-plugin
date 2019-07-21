package provider

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

// CloudWatchEventsProvider is a set of credentials which are set by the lambda request,
type CloudWatchEventsProvider struct {
	creds  credentials.Provider
	region string
}

//NewCloudWatchEventsProvider is a factory function that returns a new CloudWatchEventsProvider.
func NewCloudWatchEventsProvider(credentialsProvider credentials.Provider) *CloudWatchEventsProvider {

	return &CloudWatchEventsProvider{
		creds: credentialsProvider,
	}
}

//Get returns a new CloudWatchEvents service sesson.
func (c *CloudWatchEventsProvider) Get() (*cloudwatchevents.CloudWatchEvents, error) {

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

	return cloudwatchevents.New(sess), nil
}
