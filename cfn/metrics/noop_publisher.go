package metrics

import (
	"log"
	"strings"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/logging"
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
	datum := []string{}
	for _, v := range input.MetricData {
		datum = append(datum, v.GoString())
	}

	n.logger.Printf("Namespace: %s, Datums: %s", *input.Namespace, strings.Join(datum, " :: "))

	// out implementation doesn't care about the response
	return nil, nil
}
