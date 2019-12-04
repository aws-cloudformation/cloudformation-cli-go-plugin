// +build scheduler

/*
Package scheduler handles rescheduling resource provider handlers
when required by in_progress events.
*/
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/logging"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
	"github.com/google/uuid"
)

const (
	HandlerPrepend string = "reinvoke-handler-%s"
	TargentPrepend string = "reinvoke-target-%s"
)

const (
	// ServiceInternalError is used when there's a downstream error
	// in the code.
	ServiceInternalError string = "ServiceInternal"
)

// Result holds the confirmation of the rescheduled invocation.
type Result struct {
	// Denotes if the computation was done locally.
	ComputeLocal bool
	IDS          ScheduleIDS
}

// ScheduleIDS is of the invocation
type ScheduleIDS struct {
	// The Cloudwatch target ID.
	Target string
	// The Cloudwatch handler ID.
	Handler string
}

// Scheduler is the implementation of the rescheduler of an invoke
//
// Invokes will be rescheduled if a handler takes longer than 60
// seconds. The invoke is rescheduled through CloudWatch Events
// via a CRON expression
type Scheduler struct {
	logger *log.Logger
	client cloudwatcheventsiface.CloudWatchEventsAPI
}

// New creates a CloudWatchScheduler and returns a pointer to the struct.
func New(client cloudwatcheventsiface.CloudWatchEventsAPI) *Scheduler {
	return &Scheduler{
		logger: logging.New("scheduler"),
		client: client,
	}
}

// Reschedule when a handler requests a sub-minute callback delay, and if the lambda
// invocation has enough runtime (with 20% buffer), we can reschedule from a thread wait
// otherwise we re-invoke through CloudWatchEvents which have a granularity of
// minutes. re-invoke through CloudWatchEvents no less than 1 minute from now.
func (s *Scheduler) Reschedule(lambdaCtx context.Context, secsFromNow int64, callbackRequest string, invocationIDS *ScheduleIDS) (*Result, error) {

	lc, hasValue := lambdacontext.FromContext(lambdaCtx)

	if !hasValue {
		return nil, cfnerr.New(ServiceInternalError, "Lambda Context has no value", errors.New("Lambda Context has no value"))
	}

	deadline, _ := lambdaCtx.Deadline()
	secondsUnitDeadline := time.Until(deadline).Seconds()

	if secsFromNow <= 0 {
		err := errors.New("Scheduled seconds must be greater than 0")
		return nil, cfnerr.New(ServiceInternalError, "Scheduled seconds must be greater than 0", err)
	}

	if secsFromNow < 60 && secondsUnitDeadline > float64(secsFromNow)*1.2 {

		s.logger.Printf("Scheduling re-invoke locally after %v seconds, with Context %s", secsFromNow, string(callbackRequest))

		time.Sleep(time.Duration(secsFromNow) * time.Second)

		return &Result{ComputeLocal: true, IDS: *invocationIDS}, nil
	}

	// re-invoke through CloudWatchEvents no less than 1 minute from now.
	if secsFromNow < 60 {
		secsFromNow = 60
	}

	cr := GenerateOneTimeCronExpression(secsFromNow, time.Now())
	s.logger.Printf("Scheduling re-invoke at %s \n", cr)
	_, rerr := s.client.PutRule(&cloudwatchevents.PutRuleInput{

		Name:               aws.String(invocationIDS.Handler),
		ScheduleExpression: aws.String(cr),
		State:              aws.String(cloudwatchevents.RuleStateEnabled),
	})

	if rerr != nil {
		return nil, cfnerr.New(ServiceInternalError, "Schedule error", rerr)
	}
	_, perr := s.client.PutTargets(&cloudwatchevents.PutTargetsInput{
		Rule: aws.String(invocationIDS.Handler),
		Targets: []*cloudwatchevents.Target{
			&cloudwatchevents.Target{
				Arn:   aws.String(lc.InvokedFunctionArn),
				Id:    aws.String(invocationIDS.Target),
				Input: aws.String(string(callbackRequest)),
			},
		},
	})
	if perr != nil {
		return nil, cfnerr.New(ServiceInternalError, "Schedule error", perr)
	}

	return &Result{ComputeLocal: false, IDS: *invocationIDS}, nil
}

// CleanupEvents is used to clean up Cloudwatch Events.
// After a re-invocation, the CWE rule which generated the reinvocation should be scrubbed.
func (s *Scheduler) CleanupEvents(ruleName string, targetID string) error {

	if len(ruleName) == 0 {
		return cfnerr.New(ServiceInternalError, "Unable to complete request", errors.New("ruleName is required"))
	}
	if len(targetID) == 0 {
		return cfnerr.New(ServiceInternalError, "Unable to complete request", errors.New("targetID is required"))
	}
	_, err := s.client.RemoveTargets(&cloudwatchevents.RemoveTargetsInput{
		Ids: []*string{
			aws.String(targetID),
		},
		Rule: aws.String(ruleName),
	})
	if err != nil {
		es := fmt.Sprintf("Error cleaning CloudWatchEvents Target (targetId=%s)", targetID)
		s.logger.Println(es)
		return cfnerr.New(ServiceInternalError, es, err)
	}
	s.logger.Printf("CloudWatchEvents Target (targetId=%s) removed", targetID)

	_, rerr := s.client.DeleteRule(&cloudwatchevents.DeleteRuleInput{
		Name: aws.String(ruleName),
	})
	if rerr != nil {
		es := fmt.Sprintf("Error cleaning CloudWatchEvents (ruleName=%s)", ruleName)
		s.logger.Println(es)
		return cfnerr.New(ServiceInternalError, es, rerr)
	}
	s.logger.Printf("CloudWatchEvents Rule (ruleName=%s) removed", ruleName)

	return nil
}

// GenerateOneTimeCronExpression a cron(..) expression for a single instance
// at Now+minutesFromNow
//
// Example
//
// 	// Will generate a cron string of: "1 0 0 0 0"
// 	scheduler.GenerateOneTimeCronExpression(60, time.Now())
//
func GenerateOneTimeCronExpression(secFromNow int64, t time.Time) string {
	a := t.Add(time.Second * time.Duration(secFromNow))
	return fmt.Sprintf("cron(%02d %02d %02d %02d ? %d)", a.Minute(), a.Hour(), a.Day(), a.Month(), a.Year())
}

// GenerateCloudWatchIDS creates the targetID and handlerID for invocation
func GenerateCloudWatchIDS() (*ScheduleIDS, error) {
	uuid, err := uuid.NewUUID()

	if err != nil {
		return nil, cfnerr.New(ServiceInternalError, "uuid error", err)
	}

	handlerID := fmt.Sprintf(HandlerPrepend, uuid)
	targetID := fmt.Sprintf(TargentPrepend, uuid)

	return &ScheduleIDS{
		Target:  targetID,
		Handler: handlerID,
	}, nil
}
