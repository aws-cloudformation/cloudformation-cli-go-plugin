package logging

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type mockCWLClient struct {
	DescribeLogGroupsFn func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	CreateLogGroupFn    func(ctx context.Context, params *cloudwatchlogs.CreateLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogGroupOutput, error)
	CreateLogStreamFn   func(ctx context.Context, params *cloudwatchlogs.CreateLogStreamInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error)
	PutLogEventsFn      func(ctx context.Context, params *cloudwatchlogs.PutLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutLogEventsOutput, error)
}

func (m *mockCWLClient) DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	return m.DescribeLogGroupsFn(ctx, params, optFns...)
}

func (m *mockCWLClient) CreateLogGroup(ctx context.Context, params *cloudwatchlogs.CreateLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogGroupOutput, error) {
	return m.CreateLogGroupFn(ctx, params, optFns...)
}

func (m *mockCWLClient) CreateLogStream(ctx context.Context, params *cloudwatchlogs.CreateLogStreamInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error) {
	return m.CreateLogStreamFn(ctx, params, optFns...)
}

func (m *mockCWLClient) PutLogEvents(ctx context.Context, params *cloudwatchlogs.PutLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutLogEventsOutput, error) {
	return m.PutLogEventsFn(ctx, params, optFns...)
}

func TestCloudWatchLogProvider(t *testing.T) {
	t.Run("Init", func(t *testing.T) {
		client := &mockCWLClient{
			DescribeLogGroupsFn: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []cwltypes.LogGroup{
						{LogGroupName: params.LogGroupNamePrefix},
					},
				}, nil
			},
			CreateLogGroupFn: func(ctx context.Context, params *cloudwatchlogs.CreateLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},
			CreateLogStreamFn: func(ctx context.Context, params *cloudwatchlogs.CreateLogStreamInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, nil
			},
			PutLogEventsFn: func(ctx context.Context, params *cloudwatchlogs.PutLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutLogEventsOutput, error) {
				return &cloudwatchlogs.PutLogEventsOutput{
					NextSequenceToken: aws.String("seq1"),
				}, nil
			},
		}

		_, err := NewCloudWatchLogsProvider(client, "pineapple-pizza")
		if err != nil {
			t.Fatalf("Error returned: %v", err)
		}
	})

	t.Run("Init Error Exists", func(t *testing.T) {
		client := &mockCWLClient{
			DescribeLogGroupsFn: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return nil, errors.New("access denied")
			},
			CreateLogGroupFn: func(ctx context.Context, params *cloudwatchlogs.CreateLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},
			CreateLogStreamFn: func(ctx context.Context, params *cloudwatchlogs.CreateLogStreamInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, nil
			},
			PutLogEventsFn: func(ctx context.Context, params *cloudwatchlogs.PutLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutLogEventsOutput, error) {
				return &cloudwatchlogs.PutLogEventsOutput{
					NextSequenceToken: aws.String("seq1"),
				}, nil
			},
		}

		_, err := NewCloudWatchLogsProvider(client, "pineapple-pizza")
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})

	t.Run("Init Error Unable to Create", func(t *testing.T) {
		client := &mockCWLClient{
			DescribeLogGroupsFn: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []cwltypes.LogGroup{},
				}, nil
			},
			CreateLogGroupFn: func(ctx context.Context, params *cloudwatchlogs.CreateLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, errors.New("cannot create log group")
			},
			CreateLogStreamFn: func(ctx context.Context, params *cloudwatchlogs.CreateLogStreamInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, nil
			},
			PutLogEventsFn: func(ctx context.Context, params *cloudwatchlogs.PutLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutLogEventsOutput, error) {
				return &cloudwatchlogs.PutLogEventsOutput{
					NextSequenceToken: aws.String("seq1"),
				}, nil
			},
		}

		_, err := NewCloudWatchLogsProvider(client, "pineapple-pizza")
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})

	t.Run("Write", func(t *testing.T) {
		client := &mockCWLClient{
			DescribeLogGroupsFn: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []cwltypes.LogGroup{
						{LogGroupName: params.LogGroupNamePrefix},
					},
				}, nil
			},
			CreateLogGroupFn: func(ctx context.Context, params *cloudwatchlogs.CreateLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},
			CreateLogStreamFn: func(ctx context.Context, params *cloudwatchlogs.CreateLogStreamInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, nil
			},
			PutLogEventsFn: func(ctx context.Context, params *cloudwatchlogs.PutLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutLogEventsOutput, error) {
				return &cloudwatchlogs.PutLogEventsOutput{
					NextSequenceToken: aws.String("seq1"),
				}, nil
			},
		}

		p, err := NewCloudWatchLogsProvider(client, "pineapple-pizza")
		if err != nil {
			t.Fatalf("Error returned: %v", err)
		}

		line := "Eric loves pineapple pizza"
		c, err := p.Write([]byte(line))
		if err != nil {
			t.Fatalf("Error returned: %v", err)
		}
		if c != len(line) {
			t.Fatalf("Didn't write the same content")
		}
	})

	t.Run("Write Error", func(t *testing.T) {
		writeCount := 0
		client := &mockCWLClient{
			DescribeLogGroupsFn: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []cwltypes.LogGroup{
						{LogGroupName: params.LogGroupNamePrefix},
					},
				}, nil
			},
			CreateLogGroupFn: func(ctx context.Context, params *cloudwatchlogs.CreateLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},
			CreateLogStreamFn: func(ctx context.Context, params *cloudwatchlogs.CreateLogStreamInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, nil
			},
			PutLogEventsFn: func(ctx context.Context, params *cloudwatchlogs.PutLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutLogEventsOutput, error) {
				if writeCount == 0 {
					writeCount++
					return &cloudwatchlogs.PutLogEventsOutput{
						NextSequenceToken: aws.String("seq1"),
					}, nil
				}
				return nil, errors.New("put log events failed")
			},
		}

		p, err := NewCloudWatchLogsProvider(client, "pineapple-pizza")
		if err != nil {
			t.Fatalf("Error returned: %v", err)
		}

		line := "Eric loves pineapple pizza"
		c, err := p.Write([]byte(line))
		if err == nil && c != 0 {
			t.Fatalf("Error not returned")
		}
	})
}
