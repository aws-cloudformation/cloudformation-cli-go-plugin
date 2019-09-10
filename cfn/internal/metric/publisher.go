package metric

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

const (
	// MetricNameSpaceRoot is the Metric name space root.
	MetricNameSpaceRoot = "AWS_TMP/CloudFormation"
	//MetricNameHanderException  is a metric type.
	MetricNameHanderException = "HandlerException"
	//MetricNameHanderDuration is a metric type.
	MetricNameHanderDuration = "HandlerInvocationDuration"
	//MetricNameHanderInvocationCount is a metric type.
	MetricNameHanderInvocationCount = "HandlerInvocationCount"
	//DimensionKeyAcionType  is the Action key in the dimension.
	DimensionKeyAcionType = "Action"
	//DimensionKeyExceptionType  is the ExceptionType in the dimension.
	DimensionKeyExceptionType = "ExceptionType"
	//DimensionKeyResouceType  is the ResourceType in the dimension.
	DimensionKeyResouceType = "ResourceType"
)

type serviceProvider interface {
	Get() (cloudwatchiface.CloudWatchAPI, error)
}

// A Publisher represents an object that publishes metrics to AWS Cloudwatch.
type Publisher struct {
	cProvider serviceProvider
	client    cloudwatchiface.CloudWatchAPI // AWS CloudWatch Service Client
	namespace string                        // custom resouces's namespace
}

// New creates a new Publisher.
func New(cloudWatchProvider serviceProvider) *Publisher {
	return &Publisher{
		cProvider: cloudWatchProvider,
	}
}

//RefreshClient returns a new service session.
func (p *Publisher) RefreshClient() error {

	pr, err := p.cProvider.Get()

	if err != nil {
		return err
	}
	p.client = pr

	return nil
}

//PublishExceptionMetric publishes an exception metric.
func (p *Publisher) PublishExceptionMetric(date time.Time, action Action, e error) error {
	dimensions := map[string]string{
		DimensionKeyAcionType:     action.String(),
		DimensionKeyExceptionType: e.Error(),
		DimensionKeyResouceType:   p.SetResourceTypeName(p.namespace),
	}

	_, err := p.publishMetric(MetricNameHanderException, dimensions, cloudwatch.StandardUnitCount, 1.0, date)

	if err != nil {
		return err
	}

	return nil
}

//PublishInvocationMetric publishes an invocation metric.
func (p *Publisher) PublishInvocationMetric(date time.Time, action string) error {
	dimensions := map[string]string{
		DimensionKeyAcionType:   action,
		DimensionKeyResouceType: p.namespace,
	}

	_, err := p.publishMetric(MetricNameHanderInvocationCount, dimensions, cloudwatch.StandardUnitCount, 1.0, date)

	if err != nil {
		return err
	}

	return nil
}

//PublishDurationMetric publishes an duration metric.
func (p *Publisher) PublishDurationMetric(date time.Time, action string, secs float64) error {
	dimensions := map[string]string{
		DimensionKeyAcionType:   action,
		DimensionKeyResouceType: p.namespace,
	}

	_, err := p.publishMetric(MetricNameHanderDuration, dimensions, cloudwatch.StandardUnitMilliseconds, secs, date)

	if err != nil {
		return err
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
		Namespace:  aws.String(p.namespace),
		MetricData: md,
	}

	out, err := p.client.PutMetricData(&pi)

	if err != nil {

		return nil, err
	}

	return out, nil
}

//SetResourceTypeName returns a type name by removing (::) and replaing with (/)
func (p *Publisher) SetResourceTypeName(t string) string {
	strings.ReplaceAll(t, "::", "/")
	return t
}
