package logging

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
func NewCloudWatchLogsProvider(
	client cloudwatchlogsiface.CloudWatchLogsAPI,
	metricPublisher metricFailurePublisher,
	logGroupName string,
) (io.Writer, error) {
	logger := New("internal: ")
	fp := &failurePublisher{
		metricPublisher: metricPublisher,
	}

	// If we're running in SAM CLI, we can return the stdout
	if len(os.Getenv("AWS_SAM_LOCAL")) > 0 && len(os.Getenv("AWS_FORCE_INTEGRATIONS")) == 0 {
		return stdErr, nil
	}

	if err := CreateNewCloudWatchLogGroup(client, logGroupName); err != nil {
		fp.Publish("CreateLogGroup", err)
		return nil, err
	}

	logStreamName := ksuid.New().String()
	// need to create logstream
	if err := CreateNewLogStream(client, logGroupName, logStreamName); err != nil {
		fp.Publish("CreateLogStream", err)
		return nil, err
	}

	provider := &cloudWatchLogsProvider{
		client:           client,
		failurePublisher: fp,

		logGroupName:  logGroupName,
		logStreamName: logStreamName,

		logger: logger,
	}

	if _, err := provider.Write([]byte("Initialization of log stream")); err != nil {
		return nil, err
	}

	return provider, nil
}

type cloudWatchLogsProvider struct {
	client           cloudwatchlogsiface.CloudWatchLogsAPI
	failurePublisher *failurePublisher

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
		switch v := err.(type) {
		case *cloudwatchlogs.DataAlreadyAcceptedException:
			p.sequence = *v.ExpectedSequenceToken
		case *cloudwatchlogs.InvalidSequenceTokenException:
			p.sequence = *v.ExpectedSequenceToken
		}

		p.logger.Printf("An error occurred while putting log events [%s] to resource owner account, with error: %s", string(b), err)
		p.failurePublisher.Publish("PutLogEvents", err)
		return 0, err
	}

	p.sequence = *resp.NextSequenceToken

	return len(b), nil
}

// CreateNewCloudWatchLogGroup creates a log group in CloudWatch Logs.
//
// Using a passed in client to create the call to the service, it
// will create a log group of the specified name. If the log group
// already exists, no erorr is returned.
//
//	sess := session.Must(aws.NewConfig())
//	svc := cloudwatchlogs.New(sess)
//
//	if err := CreateNewCloudWatchLogGroup(svc, "pineapple-pizza"); err != nil {
//		panic("Unable to create log group", err)
//	}
func CreateNewCloudWatchLogGroup(client cloudwatchlogsiface.CloudWatchLogsAPI, logGroupName string) error {
	_, err := client.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	})
	if err == nil {
		return nil
	}

	awsErr, ok := err.(awserr.Error)
	if ok && awsErr.Code() == cloudwatchlogs.ErrCodeResourceAlreadyExistsException {
		return nil
	}

	return err
}

// CreateNewLogStream creates a log stream inside of a LogGroup.
// If the log stream already exists, no error is returned.
func CreateNewLogStream(client cloudwatchlogsiface.CloudWatchLogsAPI, logGroupName string, logStreamName string) error {
	_, err := client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	})

	if err == nil {
		return nil
	}

	awsErr, ok := err.(awserr.Error)
	if ok && awsErr.Code() == cloudwatchlogs.ErrCodeResourceAlreadyExistsException {
		return nil
	}

	return err
}
