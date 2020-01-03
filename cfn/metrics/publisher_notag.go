// +build !metrics

package metrics

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

// A Publisher represents an object that publishes metrics to AWS Cloudwatch.
type Publisher struct {
	client    cloudwatchiface.CloudWatchAPI // AWS CloudWatch Service Client
	namespace string
	logger    *log.Logger // custom resouces's namespace
}

// New creates a new Publisher.
func New(client cloudwatchiface.CloudWatchAPI, account string) *Publisher {
	return &Publisher{
		client:            newNoopClient(),
		logger:            logging.New("metrics"),
		providerAccountID: account,
	}
}

// PublishExceptionMetric publishes an exception metric.
func (p *Publisher) PublishExceptionMetric(date time.Time, action string, e error) error {

	if len(p.namespace) == 0 {
		message := fmt.Sprintf("Name Space was not set")
		err := errors.New(message)
		return cfnerr.New(ServiceInternalError, "Publisher error", err)
	}

	dimensions := map[string]string{
		DimensionKeyAcionType:     string(action),
		DimensionKeyExceptionType: e.Error(),
		DimensionKeyResouceType:   p.namespace,
	}

	_, err := p.publishMetric(MetricNameHanderException, dimensions, cloudwatch.StandardUnitCount, 1.0, date)

	if err != nil {
		return cfnerr.New(ServiceInternalError, "Publisher error", err)
	}

	return nil
}

// PublishInvocationMetric publishes an invocation metric.
func (p *Publisher) PublishInvocationMetric(date time.Time, action string) error {

	if len(p.namespace) == 0 {
		message := fmt.Sprintf("Name Space was not set")
		err := errors.New(message)
		return cfnerr.New(ServiceInternalError, "Publisher error", err)
	}

	dimensions := map[string]string{
		DimensionKeyAcionType:   string(action),
		DimensionKeyResouceType: p.namespace,
	}

	_, err := p.publishMetric(MetricNameHanderInvocationCount, dimensions, cloudwatch.StandardUnitCount, 1.0, date)

	if err != nil {
		return cfnerr.New(ServiceInternalError, "Publisher error", err)
	}

	return nil
}

// PublishDurationMetric publishes an duration metric.
//
// A duration metric is the timing of something.
func (p *Publisher) PublishDurationMetric(date time.Time, action string, secs float64) error {
	if len(p.namespace) == 0 {
		message := fmt.Sprintf("Name Space was not set")
		err := errors.New(message)
		return cfnerr.New(ServiceInternalError, "Publisher error", err)
	}
	dimensions := map[string]string{
		DimensionKeyAcionType:   string(action),
		DimensionKeyResouceType: p.namespace,
	}

	_, err := p.publishMetric(MetricNameHanderDuration, dimensions, cloudwatch.StandardUnitMilliseconds, secs, date)

	if err != nil {
		return cfnerr.New(ServiceInternalError, "Publisher error", err)
	}

	return nil
}

func (p *Publisher) publishMetric(metricName string, data map[string]string, unit string, value float64, date time.Time) (*cloudwatch.PutMetricDataOutput, error) {

	var d []*cloudwatch.Dimension

	for k, v := range data {
		dim := &cloudwatch.Dimension{

			Name:  aws.String(k),
			Value: aws.String(v),
		}
		d = append(d, dim)
	}

	md := []*cloudwatch.MetricDatum{
		&cloudwatch.MetricDatum{
			MetricName: aws.String(metricName),
			Unit:       aws.String(unit),
			Value:      aws.Float64(value),
			Dimensions: d,
			Timestamp:  &date},
	}

	pi := cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(fmt.Sprintf("%s/%s/%s", MetricNameSpaceRoot, p.providerAccountID, p.namespace)),
		MetricData: md,
	}

	out, err := p.client.PutMetricData(&pi)

	if err != nil {

		return nil, cfnerr.New(ServiceInternalError, "Publisher error", err)
	}

	return out, nil
}
