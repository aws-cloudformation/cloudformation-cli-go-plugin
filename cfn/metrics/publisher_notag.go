// +build !metrics

package metrics

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
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
	//ServiceInternalError ...
	ServiceInternalError string = "ServiceInternal"
)

// A Publisher represents an object that publishes metrics to AWS Cloudwatch.
type Publisher struct {
	client    cloudwatchiface.CloudWatchAPI // AWS CloudWatch Service Client
	namespace string                        // custom resouces's namespace
}

// New creates a new Publisher.
func New(client cloudwatchiface.CloudWatchAPI) *Publisher {
	if len(os.Getenv("AWS_SAM_LOCAL")) > 0 {
		client = newNoopClient()
	}

	return &Publisher{
		client: client,
	}
}

//PublishExceptionMetric publishes an exception metric.
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

//PublishInvocationMetric publishes an invocation metric.
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
		Namespace:  aws.String(p.namespace),
		MetricData: md,
	}

	out, err := p.client.PutMetricData(&pi)

	if err != nil {

		return nil, cfnerr.New(ServiceInternalError, "Publisher error", err)
	}

	return out, nil
}

// SetResourceTypeName returns a type name by removing (::) and replaing with (/)
//
// Example
//
// 	pub := metrics.New(cw)
//
// 	// Will return "AWS/Service/Resource"
// 	pub.SetResourceTypeName("AWS::Service::Resource")
func (p *Publisher) SetResourceTypeName(t string) {
	p.namespace = strings.ReplaceAll(t, "::", "/")

}
