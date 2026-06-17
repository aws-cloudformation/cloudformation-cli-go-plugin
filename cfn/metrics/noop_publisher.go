package metrics

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

type noopCloudWatchClient struct{}

func (n *noopCloudWatchClient) PutMetricData(ctx context.Context, params *cloudwatch.PutMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error) {
	return nil, nil
}
