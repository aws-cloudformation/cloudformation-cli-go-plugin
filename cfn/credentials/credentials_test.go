package credentials

import (
	"context"
	"testing"
)

func TestCredentials(t *testing.T) {
	t.Run("Retrieve", func(t *testing.T) {
		creds := &CloudFormationCredentialsProvider{
			AccessKeyID:    "a",
			SecretAccessKey: "b",
			SessionToken:   "c",
		}

		val, err := creds.Retrieve(context.Background())
		if err != nil {
			t.Fatalf("Unable to retrieve credentials: %v", err)
		}
		if val.AccessKeyID != "a" {
			t.Fatalf("Incorrect access key: %v", val.AccessKeyID)
		}
		if val.SecretAccessKey != "b" {
			t.Fatalf("Incorrect secret key: %v", val.SecretAccessKey)
		}
		if val.SessionToken != "c" {
			t.Fatalf("Incorrect session token: %v", val.SessionToken)
		}
	})
}

func TestConfigFromCredentialsProvider(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		provider := &CloudFormationCredentialsProvider{
			AccessKeyID:    "a",
			SecretAccessKey: "b",
			SessionToken:   "c",
		}
		cfg := ConfigFromCredentialsProvider(provider)

		creds, err := cfg.Credentials.Retrieve(context.Background())
		if err != nil {
			t.Fatalf("Unable to retrieve credentials from config: %v", err)
		}
		if creds.AccessKeyID != "a" {
			t.Fatalf("Incorrect access key: %v", creds.AccessKeyID)
		}
	})
}
