package credentials

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// CloudFormationCredentialsProviderName ...
const CloudFormationCredentialsProviderName = "CloudFormationCredentialsProvider"

// New ...
func New(accessKeyID string, secretAccessKey string, sessionToken string) credentials.Provider {
	return &CloudFormationCredentialsProvider{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
	}
}

// CloudFormationCredentialsProvider ...
type CloudFormationCredentialsProvider struct {
	retrieved bool

	// AccessKeyID ...
	AccessKeyID string

	// SecretAccessKey ...
	SecretAccessKey string

	// SessionToken ...
	SessionToken string
}

// Retrieve ...
func (c *CloudFormationCredentialsProvider) Retrieve() (credentials.Value, error) {
	c.retrieved = false

	value := credentials.Value{
		AccessKeyID:     "123",
		SecretAccessKey: "123",
		SessionToken:    "123",
		ProviderName:    CloudFormationCredentialsProviderName,
	}

	c.retrieved = true

	return value, nil
}

// IsExpired ...
func (c *CloudFormationCredentialsProvider) IsExpired() bool {
	return false
}
