package log

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
)

func TestCloudWatchLogProvider(t *testing.T) {
	t.Run("Init", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			DescribeLogGroupsFn: func(ctx context.Context, input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []*cloudwatchlogs.LogGroup{
						&cloudwatchlogs.LogGroup{LogGroupName: input.LogGroupNamePrefix},
					},
				}, nil
			},

			CreateLogGroupFn: func(ctx context.Context, input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},
		}

		_, err := NewCloudWatchLogsProvider(client, "pineapple-pizza")
		if err != nil {
			t.Fatalf("Error returned: %v", err)
		}
	})

	t.Run("Init Error Exists", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			DescribeLogGroupsFn: func(ctx context.Context, input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return nil, awserr.New("Invalid", "Invalid", nil)
			},

			CreateLogGroupFn: func(ctx context.Context, input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},
		}

		_, err := NewCloudWatchLogsProvider(client, "pineapple-pizza")
		if err == nil {
			t.Fatalf("Error returned: %v", err)
		}
	})

	t.Run("Init Error Unable to Create", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			DescribeLogGroupsFn: func(ctx context.Context, input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []*cloudwatchlogs.LogGroup{},
				}, nil
			},

			CreateLogGroupFn: func(ctx context.Context, input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, awserr.New("Invalid", "Invalid", nil)
			},
		}

		_, err := NewCloudWatchLogsProvider(client, "pineapple-pizza")
		if err == nil {
			t.Fatalf("Error returned: %v", err)
		}
	})

	t.Run("Write", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			DescribeLogGroupsFn: func(ctx context.Context, input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []*cloudwatchlogs.LogGroup{
						&cloudwatchlogs.LogGroup{LogGroupName: input.LogGroupNamePrefix},
					},
				}, nil
			},

			CreateLogGroupFn: func(ctx context.Context, input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},

			PutLogEventsFn: func(ctx context.Context, input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
				return nil, nil
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
		client := CallbackCloudWatchLogs{
			DescribeLogGroupsFn: func(ctx context.Context, input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []*cloudwatchlogs.LogGroup{
						&cloudwatchlogs.LogGroup{LogGroupName: input.LogGroupNamePrefix},
					},
				}, nil
			},

			CreateLogGroupFn: func(ctx context.Context, input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},

			PutLogEventsFn: func(ctx context.Context, input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
				return nil, awserr.New("Invalid", "Invalid", nil)
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

func TestCloudWatchLogGroupExists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			DescribeLogGroupsFn: func(ctx context.Context, input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []*cloudwatchlogs.LogGroup{
						&cloudwatchlogs.LogGroup{LogGroupName: input.LogGroupNamePrefix},
					},
				}, nil
			},
		}

		if _, err := CloudWatchLogGroupExists(client, "pineapple-pizza"); err != nil {
			t.Fatalf("Error returned: %v", err)
		}
	})

	t.Run("Error", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			DescribeLogGroupsFn: func(ctx context.Context, input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
				return nil, awserr.New("Invalid", "Invalid", nil)
			},
		}

		if _, err := CloudWatchLogGroupExists(client, "pineapple-pizza"); err == nil {
			t.Fatalf("Error not returned")
		}
	})
}

func TestCreateCloudWatchLogGroup(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			CreateLogGroupFn: func(ctx context.Context, input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},
		}

		if err := CreateNewCloudWatchLogGroup(client, "pineapple-pizza"); err != nil {
			t.Fatalf("Error returned: %v", err)
		}
	})

	t.Run("Error", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			CreateLogGroupFn: func(ctx context.Context, input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, awserr.New("Invalid", "Invalid", nil)
			},
		}

		if err := CreateNewCloudWatchLogGroup(client, "pineapple-pizza"); err == nil {
			t.Fatalf("Error not returned")
		}
	})
}

type CallbackCloudWatchLogs struct {
	cloudwatchlogsiface.CloudWatchLogsAPI

	DescribeLogGroupsFn func(ctx context.Context, input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	CreateLogGroupFn    func(ctx context.Context, input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error)
	PutLogEventsFn      func(ctx context.Context, input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error)
}

func (cwl CallbackCloudWatchLogs) DescribeLogGroupsWithContext(ctx context.Context, input *cloudwatchlogs.DescribeLogGroupsInput, opts ...request.Option) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	return cwl.DescribeLogGroupsFn(ctx, input)
}

func (cwl CallbackCloudWatchLogs) CreateLogGroupWithContext(ctx context.Context, input *cloudwatchlogs.CreateLogGroupInput, opts ...request.Option) (*cloudwatchlogs.CreateLogGroupOutput, error) {
	return cwl.CreateLogGroupFn(ctx, input)
}

func (cwl CallbackCloudWatchLogs) PutLogEventsWithContext(ctx context.Context, input *cloudwatchlogs.PutLogEventsInput, opts ...request.Option) (*cloudwatchlogs.PutLogEventsOutput, error) {
	return cwl.PutLogEventsFn(ctx, input)
}
