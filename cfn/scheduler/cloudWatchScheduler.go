package scheduler

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin-thulsimo/cfn/cfnerr"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
)

//CloudWatchScheduler is used to schedule Cloudwatch Events.
type CloudWatchScheduler struct {
	client cloudwatcheventsiface.CloudWatchEventsAPI
}

//New creates a CloudWatchScheduler and returns a pointer to the struct.
func New(sess cloudwatcheventsiface.CloudWatchEventsAPI) *CloudWatchScheduler {
	return &CloudWatchScheduler{
		client: sess,
	}
}

//RescheduleAfterMinutes schedules a re-invocation of the executing handler no less than 1 minute from now.
func (c *CloudWatchScheduler) RescheduleAfterMinutes(arn string, minFromNow int, callbackRequest string, t time.Time, uID string, rn string, tID string) error {

	if minFromNow < 1 {
		minFromNow = 1
	}
	if len(arn) == 0 {
		e := "Arn is required."
		return errors.New(e)
	}

	// generate a cron expression; minutes must be a positive integer
	cr := GenerateOneTimeCronExpression(minFromNow, t)
	log.Printf("Scheduling re-invoke at %s (%s)\n", cr, uID)
	pr, err := c.client.PutRule(&cloudwatchevents.PutRuleInput{

		Name:               aws.String(rn),
		ScheduleExpression: aws.String(cr),
		State:              aws.String(cloudwatchevents.RuleStateEnabled),
	})
	log.Printf("Scheduling result: %v", pr)
	if err != nil {
		return err
	}
	_, perr := c.client.PutTargets(&cloudwatchevents.PutTargetsInput{
		Rule: aws.String(rn),
		Targets: []*cloudwatchevents.Target{
			&cloudwatchevents.Target{
				Arn:   aws.String(arn),
				Id:    aws.String(tID),
				Input: aws.String(string(callbackRequest)),
			},
		},
	})
	if perr != nil {
		return err
	}

	return nil
}

//CleanupCloudWatchEvents is used to clean up Cloudwatch Events.
//After a re-invocation, the CWE rule which generated the reinvocation should be scrubbed.
func (c *CloudWatchScheduler) CleanupCloudWatchEvents(cloudWatchEventsRuleName string, cloudWatchEventsTargetID string) error {

	if len(cloudWatchEventsRuleName) == 0 {
		return cfnerr.New(ServiceInternalError, "Unable to complete request", errors.New("cloudWatchEventsRuleName is required"))
	}
	if len(cloudWatchEventsTargetID) == 0 {
		return cfnerr.New(ServiceInternalError, "Unable to complete request", errors.New("cloudWatchEventsTargetID is required"))
	}
	t, err := c.client.RemoveTargets(&cloudwatchevents.RemoveTargetsInput{
		Ids: []*string{
			aws.String(cloudWatchEventsTargetID),
		},
		Rule: aws.String(cloudWatchEventsRuleName),
	})
	if err != nil {
		es := fmt.Sprintf("Error cleaning CloudWatchEvents Target (targetId=%s)", cloudWatchEventsTargetID)
		log.Println(es)
		return cfnerr.New(ServiceInternalError, es, err)
	}
	log.Printf("CloudWatchEvents Target (targetId=%s) removed", cloudWatchEventsTargetID)
	log.Printf("CloudWatchEvents remove Target reponse: %s", t)
	r, err := c.client.DeleteRule(&cloudwatchevents.DeleteRuleInput{
		Name: aws.String(cloudWatchEventsRuleName),
	})
	if err != nil {
		es := fmt.Sprintf("Error cleaning CloudWatchEvents (ruleName=%s)", cloudWatchEventsRuleName)
		log.Println(es)
		return cfnerr.New(ServiceInternalError, es, err)
	}
	log.Printf("CloudWatchEvents (ruleName=%s) removed", cloudWatchEventsRuleName)
	log.Printf("CloudWatchEvents remove Rule reponse reponse: %s", r)

	return nil
}
