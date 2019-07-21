package provider

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

func TestNewCloudWatchProvider(t *testing.T) {
	type args struct {
		credentialsProvider credentials.Provider
	}
	tests := []struct {
		name string
		args args
		want *CloudWatchProvider
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCloudWatchProvider(tt.args.credentialsProvider); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCloudWatchProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}
