package logging

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
)

func TestCloudWatchLogProvider(t *testing.T) {
	t.Run("Init", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			CreateLogGroupFn: func(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},

			CreateLogStreamFn: func(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, nil
			},

			PutLogEventsFn: func(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
				return &cloudwatchlogs.PutLogEventsOutput{
					NextSequenceToken: aws.String("zomg"),
				}, nil
			},
		}
		fp := &fakeExceptionPublisher{}
		_, err := NewCloudWatchLogsProvider(client, fp, "pineapple-pizza")
		if err != nil {
			t.Fatalf("Error returned: %v", err)
		}
		fp.equalPublishes(t, nil)
	})

	t.Run("Init Log Group Exists", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			CreateLogGroupFn: func(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, awserr.New(cloudwatchlogs.ErrCodeResourceAlreadyExistsException, "", errors.New(""))
			},

			CreateLogStreamFn: func(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, nil
			},

			PutLogEventsFn: func(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
				return &cloudwatchlogs.PutLogEventsOutput{
					NextSequenceToken: aws.String("zomg"),
				}, nil
			},
		}
		fp := &fakeExceptionPublisher{}
		_, err := NewCloudWatchLogsProvider(client, fp, "pineapple-pizza")
		if err != nil {
			t.Fatalf("Error returned: %v", err)
		}
		fp.equalPublishes(t, nil)
	})

	t.Run("Init Log Stream Exists", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			CreateLogGroupFn: func(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},

			CreateLogStreamFn: func(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, awserr.New(cloudwatchlogs.ErrCodeResourceAlreadyExistsException, "", errors.New(""))
			},

			PutLogEventsFn: func(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
				return &cloudwatchlogs.PutLogEventsOutput{
					NextSequenceToken: aws.String("zomg"),
				}, nil
			},
		}
		fp := &fakeExceptionPublisher{}
		_, err := NewCloudWatchLogsProvider(client, fp, "pineapple-pizza")
		if err != nil {
			t.Fatalf("Error returned: %v", err)
		}
		fp.equalPublishes(t, nil)
	})

	t.Run("Init Error Unable to Create Log Group", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			CreateLogGroupFn: func(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, awserr.New("Invalid", "Invalid", nil)
			},

			CreateLogStreamFn: func(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, nil
			},

			PutLogEventsFn: func(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
				return &cloudwatchlogs.PutLogEventsOutput{
					NextSequenceToken: aws.String("zomg"),
				}, nil
			},
		}
		fp := &fakeExceptionPublisher{}
		_, err := NewCloudWatchLogsProvider(client, fp, "pineapple-pizza")
		if err == nil {
			t.Fatalf("Error returned: %v", err)
		}
		fp.equalPublishes(t, map[string]int{
			"CreateLogGroup": 1,
		})
	})

	t.Run("Init Error Unable to Create Log Stream", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			CreateLogGroupFn: func(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},

			CreateLogStreamFn: func(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, awserr.New("Invalid", "Invalid", nil)
			},

			PutLogEventsFn: func(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
				return &cloudwatchlogs.PutLogEventsOutput{
					NextSequenceToken: aws.String("zomg"),
				}, nil
			},
		}
		fp := &fakeExceptionPublisher{}
		_, err := NewCloudWatchLogsProvider(client, fp, "pineapple-pizza")
		if err == nil {
			t.Fatalf("Error returned: %v", err)
		}
		fp.equalPublishes(t, map[string]int{
			"CreateLogStream": 1,
		})
	})

	t.Run("Write", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			CreateLogGroupFn: func(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},

			CreateLogStreamFn: func(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, nil
			},

			PutLogEventsFn: func(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
				return &cloudwatchlogs.PutLogEventsOutput{
					NextSequenceToken: aws.String("zomg"),
				}, nil
			},
		}
		fp := &fakeExceptionPublisher{}
		p, err := NewCloudWatchLogsProvider(client, fp, "pineapple-pizza")
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
		fp.equalPublishes(t, nil)
	})

	t.Run("Write Error", func(t *testing.T) {
		writeCount := 0
		client := CallbackCloudWatchLogs{
			CreateLogGroupFn: func(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},

			CreateLogStreamFn: func(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, nil
			},

			PutLogEventsFn: func(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
				if writeCount == 0 {
					writeCount++
					return &cloudwatchlogs.PutLogEventsOutput{
						NextSequenceToken: aws.String("zomg"),
					}, nil
				}

				return nil, awserr.New("Invalid", "Invalid", nil)
			},
		}
		fp := &fakeExceptionPublisher{}
		p, err := NewCloudWatchLogsProvider(client, fp, "pineapple-pizza")
		if err != nil {
			t.Fatalf("Error returned: %v", err)
		}
		fp.equalPublishes(t, nil)

		line := "Eric loves pineapple pizza"
		c, err := p.Write([]byte(line))
		if err == nil && c != 0 {
			t.Fatalf("Error not returned")
		}
		fp.equalPublishes(t, map[string]int{
			"PutLogEvents": 1,
		})
	})
}

func TestCreateCloudWatchLogGroup(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			CreateLogGroupFn: func(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
				return nil, nil
			},
		}

		if err := CreateNewCloudWatchLogGroup(client, "pineapple-pizza"); err != nil {
			t.Fatalf("Error returned: %v", err)
		}
	})

	t.Run("Error", func(t *testing.T) {
		client := CallbackCloudWatchLogs{
			CreateLogGroupFn: func(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
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

	CreateLogGroupFn  func(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error)
	CreateLogStreamFn func(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error)
	PutLogEventsFn    func(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error)
}

func (cwl CallbackCloudWatchLogs) CreateLogGroup(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
	return cwl.CreateLogGroupFn(input)
}

func (cwl CallbackCloudWatchLogs) CreateLogStream(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error) {
	return cwl.CreateLogStreamFn(input)
}

func (cwl CallbackCloudWatchLogs) PutLogEvents(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
	return cwl.PutLogEventsFn(input)
}

type fakeExceptionPublisher struct {
	publishesByOperation map[string]int
}

func (p *fakeExceptionPublisher) PublishExceptionMetric(date time.Time, action string, e error) {
	if p.publishesByOperation == nil {
		p.publishesByOperation = make(map[string]int)
	}

	operation := strings.Split(e.Error(), ":")[0]

	p.publishesByOperation[operation]++
}

func (p fakeExceptionPublisher) equalPublishes(t *testing.T, expected map[string]int) {
	if expected == nil && len(p.publishesByOperation) > 0 {
		t.Fatalf("expected no exceptions to be published: %v", p.publishesByOperation)
	}

	if len(expected) != len(p.publishesByOperation) {
		t.Fatalf("expected %v, got: %v", expected, p.publishesByOperation)
	}

	for operation, count := range expected {
		if p.publishesByOperation[operation] != count {
			t.Fatalf("expected %q to have count %d, got %d. actual map: %v", operation, count, p.publishesByOperation[operation], p.publishesByOperation)
		}
	}
}
