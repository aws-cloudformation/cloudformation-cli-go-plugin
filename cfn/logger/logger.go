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

		ctx:      context.Background(),
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

	ctx      context.Context
	logLines []*cloudwatchlogs.InputLogEvent

	logLinesSize     int64
	logBufferMaxSize int64
	logFlushDuration time.Duration
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
		p.flush()
	}

	return msgBytes, nil
}

func (p *CloudWatchLogOutputProvider) requiresFlush() bool {
	return p.logLinesSize > 0 && p.logLinesSize > p.logBufferMaxSize
}

func (p *CloudWatchLogOutputProvider) setupFlusher() {
	go func() {
		for range time.Tick(p.logFlushDuration) {
			p.flush()
		}
	}()
}

func (p *CloudWatchLogOutputProvider) flush() error {
	ctx, cancel := context.WithTimeout(p.ctx, time.Millisecond*1000)
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
