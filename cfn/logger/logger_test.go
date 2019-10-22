package logger

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
)

type CloudWatchLogsAPI struct {
	cloudwatchlogsiface.CloudWatchLogsAPI
}

func (cwl CloudWatchLogsAPI) PutLogEventsWithContext(ctx aws.Context, input *cloudwatchlogs.PutLogEventsInput, opts ...request.Option) (*cloudwatchlogs.PutLogEventsOutput, error) {
	return nil, nil
}

type ErrorCloudWatchLogsAPI struct {
	cloudwatchlogsiface.CloudWatchLogsAPI
}

func (cwl ErrorCloudWatchLogsAPI) PutLogEventsWithContext(ctx aws.Context, input *cloudwatchlogs.PutLogEventsInput, opts ...request.Option) (*cloudwatchlogs.PutLogEventsOutput, error) {
	return nil, errors.New("zomg")
}

func TestCloudWatchProvider(t *testing.T) {
	t.Run("Setup With Defaults", func(t *testing.T) {
		api := CloudWatchLogsAPI{}
		provider := NewCloudWatchLogOutputProvider(
			api,
			"TestLogGroup",
		)

		if provider.logGroupName != "TestLogGroup" {
			t.Fatalf("Incorrect LogGroupName")
		}

		if len(provider.logStreamName) == 0 {
			t.Fatalf("Incorrect LogGroupName")
		}

		if provider.logBufferMaxSize != LogBufferMaxSize {
			t.Fatalf("Incorrect max buffer size")
		}
	})

	t.Run("Write", func(t *testing.T) {
		api := CloudWatchLogsAPI{}
		provider := NewCloudWatchLogOutputProvider(
			api,
			"TestLogGroup",
		)

		provider.Write([]byte("Pineapple Pizza is the best pizza"))

		if len(provider.logLines) != 1 {
			t.Fatalf("Not enough log lines")
		}
	})

	t.Run("Flush buffer", func(t *testing.T) {
		api := CloudWatchLogsAPI{}
		provider := NewCloudWatchLogOutputProvider(
			api,
			"TestLogGroup",
		)

		// set the buffer size to something super small
		provider.logBufferMaxSize = 20

		provider.Write([]byte("Pineapple Pizza is the best pizza"))

		if len(provider.logLines) != 0 {
			t.Fatalf("Log should have been flushed")
		}
	})

	t.Run("Flush larger buffer", func(t *testing.T) {
		api := CloudWatchLogsAPI{}
		provider := NewCloudWatchLogOutputProvider(
			api,
			"TestLogGroup",
		)

		// set the buffer size to something super small
		provider.logBufferMaxSize = 90

		provider.Write([]byte("Pineapple Pizza is the best pizza"))
		provider.Write([]byte("Pineapple Pizza is the best pizza"))

		if len(provider.logLines) != 2 {
			t.Fatalf("Should have two log lines")
		}

		provider.Write([]byte("Pineapple Pizza is the best pizza"))

		if len(provider.logLines) != 0 {
			t.Fatalf("Log should have been flushed, %v", provider.logLinesSize)
		}
	})

	t.Run("Flush Error", func(t *testing.T) {
		api := ErrorCloudWatchLogsAPI{}
		provider := NewCloudWatchLogOutputProvider(
			api,
			"TestLogGroup",
		)

		// set the buffer size to something super small
		provider.logBufferMaxSize = 10

		if _, err := provider.Write([]byte("Pineapple Pizza is the best pizza")); err == nil {
			t.Fatalf("Should have errored")
		}
	})
}
