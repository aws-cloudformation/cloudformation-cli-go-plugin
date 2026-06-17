package logging

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"

	"github.com/segmentio/ksuid"
)

// CloudWatchLogsClient is the subset of the CloudWatch Logs API used by this package.
type CloudWatchLogsClient interface {
	DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	CreateLogGroup(ctx context.Context, params *cloudwatchlogs.CreateLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogGroupOutput, error)
	CreateLogStream(ctx context.Context, params *cloudwatchlogs.CreateLogStreamInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error)
	PutLogEvents(ctx context.Context, params *cloudwatchlogs.PutLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutLogEventsOutput, error)
}

// NewCloudWatchLogsProvider creates an io.Writer that writes to a specific log group.
//
// Each time NewCloudWatchLogsProvider is used, a new log stream is created
// inside the log group.
func NewCloudWatchLogsProvider(client CloudWatchLogsClient, logGroupName string) (io.Writer, error) {
	logger := New("internal: ")

	// If running in SAM CLI, return stderr.
	if len(os.Getenv("AWS_SAM_LOCAL")) > 0 && len(os.Getenv("AWS_FORCE_INTEGRATIONS")) == 0 {
		return stdErr, nil
	}

	ctx := context.Background()

	ok, err := cloudWatchLogGroupExists(ctx, client, logGroupName)
	if err != nil {
		return nil, err
	}

	if !ok {
		logger.Printf("Need to create loggroup: %v", logGroupName)
		if err := createNewCloudWatchLogGroup(ctx, client, logGroupName); err != nil {
			return nil, err
		}
	}

	logStreamName := ksuid.New().String()
	if err := createNewLogStream(ctx, client, logGroupName, logStreamName); err != nil {
		return nil, err
	}

	provider := &cloudWatchLogsProvider{
		client:        client,
		logGroupName:  logGroupName,
		logStreamName: logStreamName,
		logger:        logger,
	}

	if _, err := provider.Write([]byte("Initialization of log stream")); err != nil {
		return nil, err
	}

	return provider, nil
}

type cloudWatchLogsProvider struct {
	client        CloudWatchLogsClient
	logGroupName  string
	logStreamName string
	sequence      *string
	logger        *log.Logger
}

func (p *cloudWatchLogsProvider) Write(b []byte) (int, error) {
	p.logger.Printf("Need to write: %v", string(b))

	input := &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(p.logGroupName),
		LogStreamName: aws.String(p.logStreamName),
		LogEvents: []cwltypes.InputLogEvent{
			{
				Message:   aws.String(string(b)),
				Timestamp: aws.Int64(time.Now().UnixMilli()),
			},
		},
	}

	if p.sequence != nil {
		input.SequenceToken = p.sequence
	}

	resp, err := p.client.PutLogEvents(context.Background(), input)
	if err != nil {
		return 0, err
	}

	p.sequence = resp.NextSequenceToken

	return len(b), nil
}

func cloudWatchLogGroupExists(ctx context.Context, client CloudWatchLogsClient, logGroupName string) (bool, error) {
	resp, err := client.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		Limit:              aws.Int32(1),
		LogGroupNamePrefix: aws.String(logGroupName),
	})
	if err != nil {
		return false, err
	}

	if len(resp.LogGroups) == 0 || aws.ToString(resp.LogGroups[0].LogGroupName) != logGroupName {
		return false, nil
	}

	return true, nil
}

func createNewCloudWatchLogGroup(ctx context.Context, client CloudWatchLogsClient, logGroupName string) error {
	_, err := client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	})
	return err
}

func createNewLogStream(ctx context.Context, client CloudWatchLogsClient, logGroupName, logStreamName string) error {
	_, err := client.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	})
	return err
}
