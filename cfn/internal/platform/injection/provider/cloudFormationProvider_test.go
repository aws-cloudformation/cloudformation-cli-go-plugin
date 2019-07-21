package provider

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

func TestNewCloudFormationProvider(t *testing.T) {
	type args struct {
		credentialsProvider credentials.Provider
	}
	tests := []struct {
		name string
		args args
		want *CloudFormationProvider
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCloudFormationProvider(tt.args.credentialsProvider); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCloudFormationProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}
