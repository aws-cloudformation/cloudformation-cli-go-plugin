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
	"github.com/google/uuid"
)

const (
	HandlerPrepend string = "reinvoke-handler-%s"
	TargentPrepend string = "reinvoke-target-%s"
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

//Reschedule when a handler requests a sub-minute callback delay, and if the lambda
//invocation has enough runtime (with 20% buffer), we can reschedule from a thread wait
//otherwise we re-invoke through CloudWatchEvents which have a granularity of
//minutes. re-invoke through CloudWatchEvents no less than 1 minute from now.
//Returns if re-invoke through CloudWatchEvents.
func (c *CloudWatchScheduler) Reschedule(arn string, secsFromNow int, callbackRequest string, deadline time.Time) (bool, error) {

	if len(arn) == 0 {
		err := errors.New("Arn is required")
		return false, cfnerr.New(ServiceInternalError, "Arn is required", err)
	}

	if secsFromNow <= 0 {
		err := errors.New("Scheduled seconds must be greater than 0")
		return false, cfnerr.New(ServiceInternalError, "Scheduled seconds must be greater than 0", err)
	}

	uID, err := uuid.NewUUID()

	if err != nil {
		return false, cfnerr.New(ServiceInternalError, "uuid error", err)
	}

	rn := fmt.Sprintf(HandlerPrepend, uID)
	tID := fmt.Sprintf(TargentPrepend, uID)

	ds := time.Until(deadline).Seconds()

	if secsFromNow < 60 && ds > float64(secsFromNow)*1.2 {

		log.Printf("Scheduling re-invoke locally after %v seconds, with Context %s", secsFromNow, string(callbackRequest))

		time.Sleep(time.Duration(secsFromNow) * time.Second)

		return true, nil
	}

	//re-invoke through CloudWatchEvents no less than 1 minute from now.
	if secsFromNow < 60 {
		secsFromNow = 60
	}

	cr := GenerateOneTimeCronExpression(secsFromNow, time.Now())
	log.Printf("Scheduling re-invoke at %s (%s)\n", cr, uID)
	_, rerr := c.client.PutRule(&cloudwatchevents.PutRuleInput{

		Name:               aws.String(rn),
		ScheduleExpression: aws.String(cr),
		State:              aws.String(cloudwatchevents.RuleStateEnabled),
	})

	if rerr != nil {
		return false, cfnerr.New(ServiceInternalError, "Schedule error", rerr)
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
		return false, cfnerr.New(ServiceInternalError, "Schedule error", err)
	}

	return false, nil
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
	_, err := c.client.RemoveTargets(&cloudwatchevents.RemoveTargetsInput{
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

	_, rerr := c.client.DeleteRule(&cloudwatchevents.DeleteRuleInput{
		Name: aws.String(cloudWatchEventsRuleName),
	})
	if rerr != nil {
		es := fmt.Sprintf("Error cleaning CloudWatchEvents (ruleName=%s)", cloudWatchEventsRuleName)
		log.Println(es)
		return cfnerr.New(ServiceInternalError, es, rerr)
	}
	log.Printf("CloudWatchEvents Rule (ruleName=%s) removed", cloudWatchEventsRuleName)

	return nil
}
