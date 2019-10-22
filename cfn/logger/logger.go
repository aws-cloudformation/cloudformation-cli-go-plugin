package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"

	"github.com/segmentio/ksuid"
)

const (
	// LoggingError ...
	LoggingError string = "LoggingError"
)

const (
	// LogBufferMaxSize ...
	LogBufferMaxSize = 512000

	// LogFlushDuration ...
	LogFlushDuration = time.Millisecond * 50
)

// LogOutputProvider is an interface to write logs to a io.Writer interface.
// A logger could be anything in the backend.
type LogOutputProvider interface {
	io.Writer
}

// NewCloudWatchLogOutputProvider ...
func NewCloudWatchLogOutputProvider(client cloudwatchlogsiface.CloudWatchLogsAPI, logGroupName string) *CloudWatchLogOutputProvider {
	logStreamName := ksuid.New()

	provider := &CloudWatchLogOutputProvider{
		client:        client,
		logGroupName:  logGroupName,
		logStreamName: logStreamName.String(),

		logLines: []*cloudwatchlogs.InputLogEvent{},

		logLinesSize:     0,
		logBufferMaxSize: LogBufferMaxSize,
		logFlushDuration: LogFlushDuration,
	}

	return provider
}

// CloudWatchLogOutputProvider will write logs to
// CloudWatch Logs into an AWS account.
type CloudWatchLogOutputProvider struct {
	client        cloudwatchlogsiface.CloudWatchLogsAPI
	logGroupName  string
	logStreamName string

	logLines []*cloudwatchlogs.InputLogEvent

	logLinesSize     int64
	logBufferMaxSize int64
	logFlushDuration time.Duration

	ticker *time.Ticker
}

func (p *CloudWatchLogOutputProvider) Write(b []byte) (int, error) {
	msgBytes := len(string(b))

	logLine := &cloudwatchlogs.InputLogEvent{
		Message:   aws.String(string(b)),
		Timestamp: aws.Int64(time.Now().Unix()),
	}

	p.logLines = append(p.logLines, logLine)
	p.logLinesSize = p.logLinesSize + int64(msgBytes)

	if p.requiresFlush() {
		if err := p.flush(); err != nil {
			return 0, err
		}
	}

	return msgBytes, nil
}

func (p *CloudWatchLogOutputProvider) requiresFlush() bool {
	return p.logLinesSize > 0 && p.logLinesSize > p.logBufferMaxSize
}

// EnableAutoFlush ...
func (p *CloudWatchLogOutputProvider) EnableAutoFlush() {
	ticker := time.NewTicker(p.logFlushDuration)
	p.ticker = ticker

	go func() {
		for {
			select {
			case <-ticker.C:
				p.flush()
			}
		}
	}()
}

// StopAutoFlush ...
func (p *CloudWatchLogOutputProvider) StopAutoFlush() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
}

func (p *CloudWatchLogOutputProvider) flush() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1000)
	defer cancel()

	_, err := p.client.PutLogEventsWithContext(ctx, &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(p.logGroupName),
		LogStreamName: aws.String(p.logStreamName),
		LogEvents:     p.logLines,
	})

	if err != nil {
		return cfnerr.New(LoggingError, "Unable to ship logs to CWL", err)
	}

	p.logLinesSize = 0
	p.logLines = []*cloudwatchlogs.InputLogEvent{}

	return nil
}

// NewLocalLogOutputProvider ...
func NewLocalLogOutputProvider() *LocalLogOutputProvider {
	return &LocalLogOutputProvider{
		writer: os.Stdout,
	}
}

// LocalLogOutputProvider will write logs to local
// stdout.
//
// This log provider is mostly for local execution and testing.
type LocalLogOutputProvider struct {
	writer io.Writer
}

// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
func (p *LocalLogOutputProvider) Write(b []byte) (int, error) {
	return p.writer.Write(b)
}
