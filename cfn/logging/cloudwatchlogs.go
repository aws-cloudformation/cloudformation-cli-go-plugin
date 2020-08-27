package logging

import (
	"io"
	"log"
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
	logger := New("internal: ")

	// If we're running in SAM CLI, we can return the stdout
	if len(os.Getenv("AWS_SAM_LOCAL")) > 0 && len(os.Getenv("AWS_FORCE_INTEGRATIONS")) == 0 {
		return stdErr, nil
	}

	ok, err := CloudWatchLogGroupExists(client, logGroupName)
	if err != nil {
		return nil, err
	}

	if !ok {
		logger.Printf("Need to create loggroup: %v", logGroupName)
		if err := CreateNewCloudWatchLogGroup(client, logGroupName); err != nil {
			return nil, err
		}
	}

	logStreamName := ksuid.New()
	// need to create logstream
	if err := CreateNewLogStream(client, logGroupName, logStreamName.String()); err != nil {
		return nil, err
	}

	provider := &cloudWatchLogsProvider{
		client: client,

		logGroupName:  logGroupName,
		logStreamName: logStreamName.String(),

		logger: logger,
	}

	if _, err := provider.Write([]byte("Initialization of log stream")); err != nil {
		return nil, err
	}

	return provider, nil
}

type cloudWatchLogsProvider struct {
	client cloudwatchlogsiface.CloudWatchLogsAPI

	logGroupName  string
	logStreamName string

	sequence string

	logger *log.Logger
}

func (p *cloudWatchLogsProvider) Write(b []byte) (int, error) {
	p.logger.Printf("Need to write: %v", string(b))

	input := &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(p.logGroupName),
		LogStreamName: aws.String(p.logStreamName),

		LogEvents: []*cloudwatchlogs.InputLogEvent{
			&cloudwatchlogs.InputLogEvent{
				Message:   aws.String(string(b)),
				Timestamp: aws.Int64(time.Now().UnixNano() / int64(time.Millisecond)),
			},
		},
	}

	if len(p.sequence) != 0 {
		input.SetSequenceToken(p.sequence)
	}

	resp, err := p.client.PutLogEvents(input)

	if err != nil {
		return 0, err
	}

	p.sequence = *resp.NextSequenceToken

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
	resp, err := client.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{
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
	if _, err := client.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	}); err != nil {
		return err
	}

	return nil
}

// CreateNewLogStream creates a log stream inside of a LogGroup
func CreateNewLogStream(client cloudwatchlogsiface.CloudWatchLogsAPI, logGroupName string, logStreamName string) error {
	_, err := client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	})

	return err
}
