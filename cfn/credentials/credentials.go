/*
Package credentials provides helper functions for dealing with AWS credentials
passed in to resource providers from CloudFormation.
*/
package credentials

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

const (
	// CloudFormationCredentialsProviderName is the name of the credentials provider.
	CloudFormationCredentialsProviderName = "CloudFormationCredentialsProvider"
	// InvalidSessionError is returned when a session is invalid.
	InvalidSessionError = "InvalidSession"
)

// CloudFormationCredentialsProvider holds the credentials passed by CloudFormation.
type CloudFormationCredentialsProvider struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey  string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
}

// Retrieve implements aws.CredentialsProvider.
func (c *CloudFormationCredentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{
		AccessKeyID:     c.AccessKeyID,
		SecretAccessKey:  c.SecretAccessKey,
		SessionToken:    c.SessionToken,
		Source:          CloudFormationCredentialsProviderName,
	}, nil
}

// ConfigFromCredentialsProvider creates an aws.Config from a credentials provider.
func ConfigFromCredentialsProvider(provider *CloudFormationCredentialsProvider) aws.Config {
	cfg, _ := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				provider.AccessKeyID,
				provider.SecretAccessKey,
				provider.SessionToken,
			),
		),
	)
	return cfg
}
