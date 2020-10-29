package metrics

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

const (
	// MetricNameSpaceRoot is the Metric name space root.
	MetricNameSpaceRoot = "AWS/CloudFormation"
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
	//DimensionKeyResourceType  is the ResourceType in the dimension.
	DimensionKeyResourceType = "ResourceType"
	//ServiceInternalError ...
	ServiceInternalError string = "ServiceInternal"
)

// A Publisher represents an object that publishes metrics to AWS Cloudwatch.
type Publisher struct {
	client       cloudwatchiface.CloudWatchAPI // AWS CloudWatch Service Client
	namespace    string                        // custom resouces's namespace
	logger       *log.Logger
	resourceType string // type of resource
}

// New creates a new Publisher.
func New(client cloudwatchiface.CloudWatchAPI, resType string) *Publisher {
	if len(os.Getenv("AWS_SAM_LOCAL")) > 0 {
		client = newNoopClient()
	}
	rn := ResourceTypeName(resType)
	return &Publisher{
		client:       client,
		logger:       logging.New("metrics"),
		namespace:    fmt.Sprintf("%s/%s", MetricNameSpaceRoot, rn),
		resourceType: rn,
	}
}

// PublishExceptionMetric publishes an exception metric.
func (p *Publisher) PublishExceptionMetric(date time.Time, action string, e error) {
	v := strings.ReplaceAll(e.Error(), "\n", " ")
	dimensions := map[string]string{
		DimensionKeyAcionType:     string(action),
		DimensionKeyExceptionType: v,
		DimensionKeyResourceType:  p.resourceType,
	}
	p.publishMetric(MetricNameHanderException, dimensions, cloudwatch.StandardUnitCount, 1.0, date)
}

// PublishInvocationMetric publishes an invocation metric.
func (p *Publisher) PublishInvocationMetric(date time.Time, action string) {
	dimensions := map[string]string{
		DimensionKeyAcionType:    string(action),
		DimensionKeyResourceType: p.resourceType,
	}
	p.publishMetric(MetricNameHanderInvocationCount, dimensions, cloudwatch.StandardUnitCount, 1.0, date)
}

// PublishDurationMetric publishes an duration metric.
//
// A duration metric is the timing of something.
func (p *Publisher) PublishDurationMetric(date time.Time, action string, secs float64) {
	dimensions := map[string]string{
		DimensionKeyAcionType:    string(action),
		DimensionKeyResourceType: p.resourceType,
	}
	p.publishMetric(MetricNameHanderDuration, dimensions, cloudwatch.StandardUnitMilliseconds, secs, date)
}

func (p *Publisher) publishMetric(metricName string, data map[string]string, unit string, value float64, date time.Time) {

	var d []*cloudwatch.Dimension

	for k, v := range data {
		dim := &cloudwatch.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v),
		}
		d = append(d, dim)
	}
	md := []*cloudwatch.MetricDatum{
		{
			MetricName: aws.String(metricName),
			Unit:       aws.String(unit),
			Value:      aws.Float64(value),
			Dimensions: d,
			Timestamp:  aws.Time(date.UTC()),
		},
	}
	pi := cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(p.namespace),
		MetricData: md,
	}
	_, err := p.client.PutMetricData(&pi)
	if err != nil {
		p.logger.Printf("An error occurred while publishing metrics: %s", err)

	}
}

// ResourceTypeName returns a type name by removing (::) and replaing with (/)
//
// Example
//
// 	r := metrics.ResourceTypeName("AWS::Service::Resource")
//
// 	// Will return "AWS/Service/Resource"
func ResourceTypeName(t string) string {
	return strings.ReplaceAll(t, "::", "/")

}
