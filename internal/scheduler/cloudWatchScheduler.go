package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws-cloudformation-rpdk-go-plugin/internal/platform/proxy"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
)

//CloudWatchScheduler is used to schedule Cloudwatch Events.
type CloudWatchScheduler struct {
	Client cloudwatcheventsiface.CloudWatchEventsAPI
}

//New creates a CloudWatchScheduler and returns a pointer to the struct.
func New(cl cloudwatcheventsiface.CloudWatchEventsAPI) *CloudWatchScheduler {
	return &CloudWatchScheduler{
		Client: cl,
	}
}

//RescheduleAfterMinutes schedules a re-invocation of the executing handler no less than 1 minute from now.
func (c *CloudWatchScheduler) RescheduleAfterMinutes(arn string, minFromNow int, callbackContext *proxy.RequestContext, t time.Time) error {

	if minFromNow < 1 {
		e := "minFromNow must be positve."
		return errors.New(e)
	}

	if arn == "" {
		e := "Arn is required."
		return errors.New(e)
	}

	// generate a cron expression; minutes must be a positive integer
	cr := GenerateOneTimeCronExpression(minFromNow, t)

	rd, err := NewUUID()
	if err != nil {
		return err
	}

	rn := fmt.Sprintf("reinvoke-handler-%s", rd)
	tID := fmt.Sprintf("reinvoke-target-%s", rd)

	// record the CloudWatchEvents objects for cleanup on the callback
	callbackContext.CloudWatchEventsRuleName = rn
	callbackContext.CloudWatchEventsTargetID = tID

	cj, err := json.Marshal(callbackContext)

	if err != nil {
		return err
	}
	log.Printf("Scheduling re-invoke at %s (%s)\n", cr, rd)

	pr, err := c.Client.PutRule(&cloudwatchevents.PutRuleInput{

		Name:               aws.String(rn),
		ScheduleExpression: aws.String(cr),
		State:              aws.String(cloudwatchevents.RuleStateEnabled),
	})

	log.Printf("Scheduling result: %v", pr)

	if err != nil {
		return err
	}

	tr, err := c.Client.PutTargets(&cloudwatchevents.PutTargetsInput{
		Rule: aws.String(rn),
		Targets: []*cloudwatchevents.Target{
			&cloudwatchevents.Target{
				Arn:   aws.String(arn),
				Id:    aws.String(tID),
				Input: aws.String(string(cj)),
			},
		},
	})

	if err != nil {
		return err
	}

	log.Printf("PutTargets result : %v ", tr)

	return nil
}

//CleanupCloudWatchEvents is used to clean up Cloudwatch Events.
//After a re-invocation, the CWE rule which generated the reinvocation should be scrubbed.
func (c *CloudWatchScheduler) CleanupCloudWatchEvents(cloudWatchEventsRuleName string, cloudWatchEventsTargetID string) error {

	if cloudWatchEventsRuleName == "" {
		e := "cloudWatchEventsRuleName is required."
		return errors.New(e)
	}

	if cloudWatchEventsTargetID == "" {
		e := "cloudWatchEventsTargetID is required."
		return errors.New(e)
	}

	t, err := c.Client.RemoveTargets(&cloudwatchevents.RemoveTargetsInput{
		Ids: []*string{
			aws.String(cloudWatchEventsTargetID),
		},
		Rule: aws.String(cloudWatchEventsRuleName),
	})

	if err != nil {
		log.Printf("Error cleaning CloudWatchEvents Target (targetId=%s)", cloudWatchEventsTargetID)
		return err
	}

	log.Printf("CloudWatchEvents Target (targetId=%s) removed", cloudWatchEventsTargetID)
	log.Printf("CloudWatchEvents remove Target reponse: %s", t)

	r, err := c.Client.DeleteRule(&cloudwatchevents.DeleteRuleInput{
		Name: aws.String(cloudWatchEventsRuleName),
	})
	if err != nil {
		log.Printf("Error cleaning CloudWatchEvents (ruleName=%s)", cloudWatchEventsRuleName)
		return err
	}
	log.Printf("CloudWatchEvents (ruleName=%s) removed", cloudWatchEventsRuleName)
	log.Printf("CloudWatchEvents remove Rule reponse reponse: %s", r)

	return nil
}
