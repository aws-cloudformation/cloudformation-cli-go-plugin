package provider

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// PlatformCredentialsProviderName provides a name of Static provider
const PlatformCredentialsProviderName = "PlatformCredentialsProvider"

var (
	// ErrStaticCredentialsEmpty is emitted when static credentials are empty.
	ErrStaticCredentialsEmpty = awserr.New("EmptyStaticCreds", "static credentials are empty", nil)
)

// PlatformCredentialsProvider is a set of credentials which are set by the lambda request,
// and will never expire.
type PlatformCredentialsProvider struct {
	credentials.Value
}

//NewPlatformCredentialsProvider is a factory function that returns a new PlatformCredentialsProvider.
func NewPlatformCredentialsProvider(id, secret, token string) *PlatformCredentialsProvider {
	return &PlatformCredentialsProvider{credentials.Value{
		AccessKeyID:     id,
		SecretAccessKey: secret,
		SessionToken:    token,
	}}
}

// Retrieve returns the credentials or error if the credentials are invalid.
func (p *PlatformCredentialsProvider) Retrieve() (credentials.Value, error) {
	if p.AccessKeyID == "" || p.SecretAccessKey == "" {
		return credentials.Value{ProviderName: PlatformCredentialsProviderName}, ErrStaticCredentialsEmpty
	}

	if len(p.Value.ProviderName) == 0 {
		p.Value.ProviderName = PlatformCredentialsProviderName
	}
	return p.Value, nil
}

// IsExpired returns if the credentials are expired.
//
// For PlatformCredentialsProvider, the credentials never expired.
func (p *PlatformCredentialsProvider) IsExpired() bool {
	return false
}
