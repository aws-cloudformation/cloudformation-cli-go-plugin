package logging

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

// FailurePublisher leverages metricFailurePublisher to publish log failures
type failurePublisher struct {
	metricPublisher metricFailurePublisher
}

type metricFailurePublisher interface {
	PublishExceptionMetric(date time.Time, action string, e error)
}

func (p *failurePublisher) Publish(operation string, err error) {
	p.metricPublisher.PublishExceptionMetric(time.Now(), "ProviderLogDelivery", errFromLogFailure(operation, err))
}

func errFromLogFailure(operation string, err error) error {
	awsErr, ok := err.(awserr.Error)
	if !ok {
		// avoid %w for compatiblility
		return fmt.Errorf("%s: %s", operation, err)
	}

	return fmt.Errorf("%s: %s", operation, awsErr.Code())
}
