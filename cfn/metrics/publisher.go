package metrics

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

const (
	MetricNameSpaceRoot             = "AWS/CloudFormation"
	MetricNameHanderException       = "HandlerException"
	MetricNameHanderDuration        = "HandlerInvocationDuration"
	MetricNameHanderInvocationCount = "HandlerInvocationCount"
	DimensionKeyAcionType           = "Action"
	DimensionKeyExceptionType       = "ExceptionType"
	DimensionKeyResourceType        = "ResourceType"
	ServiceInternalError            = "ServiceInternal"
)

// CloudWatchClient is the subset of the CloudWatch API used by this package.
type CloudWatchClient interface {
	PutMetricData(ctx context.Context, params *cloudwatch.PutMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error)
}

// Publisher publishes metrics to AWS CloudWatch.
type Publisher struct {
	client       CloudWatchClient
	namespace    string
	logger       *log.Logger
	resourceType string
}

// New creates a new Publisher.
func New(client CloudWatchClient, resType string) *Publisher {
	if len(os.Getenv("AWS_SAM_LOCAL")) > 0 {
		client = &noopCloudWatchClient{}
	}
	rn := ResourceTypeName(resType)
	return &Publisher{
		client:       client,
		logger:       logging.New("metrics"),
		namespace:    fmt.Sprintf("%s/%s", MetricNameSpaceRoot, rn),
		resourceType: rn,
	}
}

func (p *Publisher) PublishExceptionMetric(date time.Time, action string, e error) {
	v := strings.ReplaceAll(e.Error(), "\n", " ")
	dimensions := map[string]string{
		DimensionKeyAcionType:     action,
		DimensionKeyExceptionType: v,
		DimensionKeyResourceType:  p.resourceType,
	}
	p.publishMetric(MetricNameHanderException, dimensions, cwtypes.StandardUnitCount, 1.0, date)
}

func (p *Publisher) PublishInvocationMetric(date time.Time, action string) {
	dimensions := map[string]string{
		DimensionKeyAcionType:    action,
		DimensionKeyResourceType: p.resourceType,
	}
	p.publishMetric(MetricNameHanderInvocationCount, dimensions, cwtypes.StandardUnitCount, 1.0, date)
}

func (p *Publisher) PublishDurationMetric(date time.Time, action string, secs float64) {
	dimensions := map[string]string{
		DimensionKeyAcionType:    action,
		DimensionKeyResourceType: p.resourceType,
	}
	p.publishMetric(MetricNameHanderDuration, dimensions, cwtypes.StandardUnitMilliseconds, secs, date)
}

func (p *Publisher) publishMetric(metricName string, data map[string]string, unit cwtypes.StandardUnit, value float64, date time.Time) {
	var d []cwtypes.Dimension
	for k, v := range data {
		d = append(d, cwtypes.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	md := []cwtypes.MetricDatum{
		{
			MetricName: aws.String(metricName),
			Unit:       unit,
			Value:      aws.Float64(value),
			Dimensions: d,
			Timestamp:  &date,
		},
	}

	_, err := p.client.PutMetricData(context.Background(), &cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(p.namespace),
		MetricData: md,
	})
	if err != nil {
		p.logger.Printf("An error occurred while publishing metrics: %s", err)
	}
}

// ResourceTypeName returns a type name by replacing (::) with (/).
func ResourceTypeName(t string) string {
	return strings.ReplaceAll(t, "::", "/")
}
