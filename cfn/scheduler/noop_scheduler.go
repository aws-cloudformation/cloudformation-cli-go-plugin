package scheduler

import (
	"log"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/logging"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
)

type noopCloudWatchClient struct {
	cloudwatcheventsiface.CloudWatchEventsAPI
	logger *log.Logger
}

func newNoopCloudWatchClient() *noopCloudWatchClient {
	return &noopCloudWatchClient{
		logger: logging.New("scheduler"),
	}
}

func (m *noopCloudWatchClient) PutRule(in *cloudwatchevents.PutRuleInput) (*cloudwatchevents.PutRuleOutput, error) {
	m.logger.Printf("Rule name: %v", *in.Name)
	// out implementation doesn't care about the response
	return nil, nil
}

func (m *noopCloudWatchClient) PutTargets(in *cloudwatchevents.PutTargetsInput) (*cloudwatchevents.PutTargetsOutput, error) {
	m.logger.Printf("Target ID: %v", *in.Targets[0].Id)
	// out implementation doesn't care about the response
	return nil, nil

}

func (m *noopCloudWatchClient) DeleteRule(in *cloudwatchevents.DeleteRuleInput) (*cloudwatchevents.DeleteRuleOutput, error) {
	m.logger.Printf("Rule name: %v", *in.Name)
	// out implementation doesn't care about the response
	return nil, nil
}

func (m *noopCloudWatchClient) RemoveTargets(in *cloudwatchevents.RemoveTargetsInput) (*cloudwatchevents.RemoveTargetsOutput, error) {
	m.logger.Printf("Target ID: %v", *in.Ids[0])
	// out implementation doesn't care about the response
	return nil, nil
}
