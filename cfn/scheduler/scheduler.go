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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchevents"
	cwetypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchevents/types"
	"github.com/google/uuid"
)

const (
	HandlerPrepend string = "reinvoke-handler-%s"
	TargentPrepend string = "reinvoke-target-%s"
)

const (
	ServiceInternalError string = "ServiceInternal"
)

// CloudWatchEventsClient is the subset of the CloudWatch Events API used by this package.
type CloudWatchEventsClient interface {
	PutRule(ctx context.Context, params *cloudwatchevents.PutRuleInput, optFns ...func(*cloudwatchevents.Options)) (*cloudwatchevents.PutRuleOutput, error)
	PutTargets(ctx context.Context, params *cloudwatchevents.PutTargetsInput, optFns ...func(*cloudwatchevents.Options)) (*cloudwatchevents.PutTargetsOutput, error)
	DeleteRule(ctx context.Context, params *cloudwatchevents.DeleteRuleInput, optFns ...func(*cloudwatchevents.Options)) (*cloudwatchevents.DeleteRuleOutput, error)
	RemoveTargets(ctx context.Context, params *cloudwatchevents.RemoveTargetsInput, optFns ...func(*cloudwatchevents.Options)) (*cloudwatchevents.RemoveTargetsOutput, error)
}

// Result holds the confirmation of the rescheduled invocation.
type Result struct {
	ComputeLocal bool
	IDS          ScheduleIDS
}

// ScheduleIDS is of the invocation
type ScheduleIDS struct {
	Target  string
	Handler string
}

// Scheduler reschedules handler invocations via CloudWatch Events.
type Scheduler struct {
	logger *log.Logger
	client CloudWatchEventsClient
}

// New creates a Scheduler.
func New(client CloudWatchEventsClient) *Scheduler {
	return &Scheduler{
		logger: logging.New("scheduler"),
		client: client,
	}
}

// Reschedule handles re-invocation logic.
func (s *Scheduler) Reschedule(lambdaCtx context.Context, secsFromNow int64, callbackRequest string, invocationIDS *ScheduleIDS) (*Result, error) {
	lc, hasValue := lambdacontext.FromContext(lambdaCtx)
	if !hasValue {
		return nil, cfnerr.New(ServiceInternalError, "Lambda Context has no value", errors.New("Lambda Context has no value"))
	}

	deadline, _ := lambdaCtx.Deadline()
	secondsUntilDeadline := time.Until(deadline).Seconds()

	if secsFromNow <= 0 {
		return nil, cfnerr.New(ServiceInternalError, "Scheduled seconds must be greater than 0", errors.New("Scheduled seconds must be greater than 0"))
	}

	if secsFromNow < 60 && secondsUntilDeadline > float64(secsFromNow)*1.2 {
		s.logger.Printf("Scheduling re-invoke locally after %v seconds, with Context %s", secsFromNow, callbackRequest)
		time.Sleep(time.Duration(secsFromNow) * time.Second)
		return &Result{ComputeLocal: true, IDS: *invocationIDS}, nil
	}

	if secsFromNow < 60 {
		secsFromNow = 60
	}

	ctx := context.Background()
	cr := GenerateOneTimeCronExpression(secsFromNow, time.Now())
	s.logger.Printf("Scheduling re-invoke at %s\n", cr)

	_, rerr := s.client.PutRule(ctx, &cloudwatchevents.PutRuleInput{
		Name:               aws.String(invocationIDS.Handler),
		ScheduleExpression: aws.String(cr),
		State:              cwetypes.RuleStateEnabled,
	})
	if rerr != nil {
		return nil, cfnerr.New(ServiceInternalError, "Schedule error", rerr)
	}

	_, perr := s.client.PutTargets(ctx, &cloudwatchevents.PutTargetsInput{
		Rule: aws.String(invocationIDS.Handler),
		Targets: []cwetypes.Target{
			{
				Arn:   aws.String(lc.InvokedFunctionArn),
				Id:    aws.String(invocationIDS.Target),
				Input: aws.String(callbackRequest),
			},
		},
	})
	if perr != nil {
		return nil, cfnerr.New(ServiceInternalError, "Schedule error", perr)
	}

	return &Result{ComputeLocal: false, IDS: *invocationIDS}, nil
}

// CleanupEvents removes CloudWatch Events rule and target after re-invocation.
func (s *Scheduler) CleanupEvents(ruleName string, targetID string) error {
	if len(ruleName) == 0 {
		return cfnerr.New(ServiceInternalError, "Unable to complete request", errors.New("ruleName is required"))
	}
	if len(targetID) == 0 {
		return cfnerr.New(ServiceInternalError, "Unable to complete request", errors.New("targetID is required"))
	}

	ctx := context.Background()

	_, err := s.client.RemoveTargets(ctx, &cloudwatchevents.RemoveTargetsInput{
		Ids:  []string{targetID},
		Rule: aws.String(ruleName),
	})
	if err != nil {
		es := fmt.Sprintf("Error cleaning CloudWatchEvents Target (targetId=%s)", targetID)
		s.logger.Println(es)
		return cfnerr.New(ServiceInternalError, es, err)
	}
	s.logger.Printf("CloudWatchEvents Target (targetId=%s) removed", targetID)

	_, rerr := s.client.DeleteRule(ctx, &cloudwatchevents.DeleteRuleInput{
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

// GenerateOneTimeCronExpression generates a cron expression for a single invocation.
func GenerateOneTimeCronExpression(secFromNow int64, t time.Time) string {
	a := t.Add(time.Second * time.Duration(secFromNow))
	return fmt.Sprintf("cron(%02d %02d %02d %02d ? %d)", a.Minute(), a.Hour(), a.Day(), a.Month(), a.Year())
}

// GenerateCloudWatchIDS creates the targetID and handlerID for invocation.
func GenerateCloudWatchIDS() (*ScheduleIDS, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return nil, cfnerr.New(ServiceInternalError, "uuid error", err)
	}

	return &ScheduleIDS{
		Target:  fmt.Sprintf(TargentPrepend, u),
		Handler: fmt.Sprintf(HandlerPrepend, u),
	}, nil
}
