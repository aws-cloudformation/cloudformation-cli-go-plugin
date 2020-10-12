package metrics

import (
	"log"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/logging"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

func newNoopClient() *noopCloudWatchClient {
	return &noopCloudWatchClient{
		logger: logging.New("metrics"),
	}
}

type noopCloudWatchClient struct {
	logger *log.Logger
	cloudwatchiface.CloudWatchAPI
}

func (n *noopCloudWatchClient) PutMetricData(input *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {
	// out implementation doesn't care about the response
	return nil, nil
}
