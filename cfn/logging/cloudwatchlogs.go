// +build logging

package logging

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"

	"github.com/segmentio/ksuid"
)

// NewCloudWatchLogsProvider creates a io.Writer that writes
// to a specifc log group.
//
// Each time NewCloudWatchLogsProvider is used, a new log stream is created
// inside the log group. The log stream will have a unique, random identifer
//
//	sess := session.Must(aws.NewConfig())
//	svc := cloudwatchlogs.New(sess)
//
//	provider, err := NewCloudWatchLogsProvider(svc, "pineapple-pizza")
//	if err != nil {
//		panic(err)
//	}
//
//	// set log output to the provider, all log messages will then be
//	// pushed through the Write func and sent to CloudWatch Logs
//	log.SetOutput(provider)
//	log.Printf("Eric loves pineapple pizza!")
func NewCloudWatchLogsProvider(client cloudwatchlogsiface.CloudWatchLogsAPI, logGroupName string) (io.Writer, error) {
	// If we're running in SAM CLI, we can return the stdout
	if len(os.Getenv("AWS_SAM_LOCAL")) > 0 {
		return stdErr, nil
	}

	ok, err := CloudWatchLogGroupExists(client, logGroupName)
	if err != nil {
		return nil, err
	}

	if !ok {
		if err := CreateNewCloudWatchLogGroup(client, logGroupName); err != nil {
			return nil, err
		}
	}

	logStreamName := ksuid.New()

	provider := &cloudWatchLogsProvider{
		client: client,

		logGroupName:  logGroupName,
		logStreamName: logStreamName.String(),
	}

	return provider, nil
}

type cloudWatchLogsProvider struct {
	client cloudwatchlogsiface.CloudWatchLogsAPI

	logGroupName  string
	logStreamName string
}

func (p *cloudWatchLogsProvider) Write(b []byte) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	if _, err := p.client.PutLogEventsWithContext(ctx, &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(p.logGroupName),
		LogStreamName: aws.String(p.logStreamName),

		LogEvents: []*cloudwatchlogs.InputLogEvent{
			&cloudwatchlogs.InputLogEvent{
				Message:   aws.String(string(b)),
				Timestamp: aws.Int64(time.Now().Unix()),
			},
		},
	}); err != nil {
		return 0, err
	}

	return len(b), nil
}

// CloudWatchLogGroupExists checks if a log group exists
//
// Using the client provided, it will check the CloudWatch Logs
// service to verify the log group
//
//	sess := session.Must(aws.NewConfig())
//	svc := cloudwatchlogs.New(sess)
//
//	// checks if the pineapple-pizza log group exists
//	ok, err := LogGroupExists(svc, "pineapple-pizza")
//	if err != nil {
//		panic(err)
//	}
//	if ok {
//		// do something
//	}
func CloudWatchLogGroupExists(client cloudwatchlogsiface.CloudWatchLogsAPI, logGroupName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	resp, err := client.DescribeLogGroupsWithContext(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		Limit:              aws.Int64(1),
		LogGroupNamePrefix: aws.String(logGroupName),
	})

	if err != nil {
		return false, err
	}

	if len(resp.LogGroups) == 0 || *resp.LogGroups[0].LogGroupName != logGroupName {
		return false, nil
	}

	return true, nil
}

// CreateNewCloudWatchLogGroup creates a log group in CloudWatch Logs.
//
// Using a passed in client to create the call to the service, it
// will create a log group of the specified name
//
//	sess := session.Must(aws.NewConfig())
//	svc := cloudwatchlogs.New(sess)
//
//	if err := CreateNewCloudWatchLogGroup(svc, "pineapple-pizza"); err != nil {
//		panic("Unable to create log group", err)
//	}
func CreateNewCloudWatchLogGroup(client cloudwatchlogsiface.CloudWatchLogsAPI, logGroupName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	if _, err := client.CreateLogGroupWithContext(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	}); err != nil {
		return err
	}

	return nil
}
