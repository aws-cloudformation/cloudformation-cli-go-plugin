package scheduler

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/logging"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchevents"
)

const (
	TargetRegx  string = `reinvoke-target-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`
	HandlerRegx string = `reinvoke-handler-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`
	Succeed     string = "\u2713"
	Failed      string = "\u2717"
	Arn         string = "arn:aws:lambda:us-east-2:123456789:function:myproject"
)

type MockedEvents struct {
	RuleName   string
	TargetName string
}

func NewMockEvents() *MockedEvents {
	return &MockedEvents{}
}

func (m *MockedEvents) PutRule(ctx context.Context, in *cloudwatchevents.PutRuleInput, optFns ...func(*cloudwatchevents.Options)) (*cloudwatchevents.PutRuleOutput, error) {
	m.RuleName = *in.Name
	return nil, nil
}

func (m *MockedEvents) PutTargets(ctx context.Context, in *cloudwatchevents.PutTargetsInput, optFns ...func(*cloudwatchevents.Options)) (*cloudwatchevents.PutTargetsOutput, error) {
	m.TargetName = *in.Targets[0].Id
	return nil, nil
}

func (m *MockedEvents) DeleteRule(ctx context.Context, in *cloudwatchevents.DeleteRuleInput, optFns ...func(*cloudwatchevents.Options)) (*cloudwatchevents.DeleteRuleOutput, error) {
	m.RuleName = *in.Name
	return nil, nil
}

func (m *MockedEvents) RemoveTargets(ctx context.Context, in *cloudwatchevents.RemoveTargetsInput, optFns ...func(*cloudwatchevents.Options)) (*cloudwatchevents.RemoveTargetsOutput, error) {
	m.TargetName = in.Ids[0]
	return nil, nil
}

func TestGenerateOneTimeCronExpression(t *testing.T) {
	tests := []struct {
		name string
		secs int64
		t    time.Time
		want string
	}{
		{"t1", 0, time.Date(2009, 11, 17, 20, 34, 58, 0, time.UTC), "cron(34 20 17 11 ? 2009)"},
		{"t2", 0, time.Date(2001, 5, 25, 11, 04, 14, 0, time.UTC), "cron(04 11 25 05 ? 2001)"},
		{"t3", 0, time.Date(2006, 7, 17, 7, 18, 23, 0, time.UTC), "cron(18 07 17 07 ? 2006)"},
		{"t4", 0, time.Date(1999, 2, 07, 21, 28, 45, 0, time.UTC), "cron(28 21 07 02 ? 1999)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateOneTimeCronExpression(tt.secs, tt.t)
			if got != tt.want {
				t.Errorf("GenerateOneTimeCronExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCloudWatchSchedulerReschedule(t *testing.T) {
	cb := `{ string: "Foo"}`
	ct1, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*1000))
	defer cancel()
	ct2, cancel2 := context.WithDeadline(context.Background(), time.Now().Add(time.Second*16))
	defer cancel2()

	tests := []struct {
		name         string
		ctx          context.Context
		secFromNow   int64
		wantErr      bool
		computeLocal bool
	}{
		{"local compute", lambdacontext.NewContext(ct1, &lambdacontext.LambdaContext{InvokedFunctionArn: Arn}), 15, false, true},
		{"remote compute", lambdacontext.NewContext(ct2, &lambdacontext.LambdaContext{InvokedFunctionArn: Arn}), 15, false, false},
		{"negative secs", lambdacontext.NewContext(ct1, &lambdacontext.LambdaContext{InvokedFunctionArn: Arn}), -87, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Scheduler{
				client: NewMockEvents(),
				logger: logging.New("scheduler: "),
			}
			ids, _ := GenerateCloudWatchIDS()
			result, err := c.Reschedule(tt.ctx, tt.secFromNow, cb, ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reschedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && result.ComputeLocal != tt.computeLocal {
				t.Errorf("ComputeLocal = %v, want %v", result.ComputeLocal, tt.computeLocal)
			}
		})
	}
}

func TestCloudWatchSchedulerCleanupEvents(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		target   string
		wantErr  bool
	}{
		{"valid", "reinvoke-handler-c51d7ba5-8eed-4226-99a6-6743f1169688", "reinvoke-target-c51d7ba5-8eed-4226-99a6-6743f1169688", false},
		{"blank rule", "", "reinvoke-target-c51d7ba5-8eed-4226-99a6-6743f1169688", true},
		{"blank target", "reinvoke-handler-c51d7ba5-8eed-4226-99a6-6743f1169688", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Scheduler{
				client: NewMockEvents(),
				logger: logging.New("scheduler: "),
			}
			if err := c.CleanupEvents(tt.rule, tt.target); (err != nil) != tt.wantErr {
				t.Errorf("CleanupEvents() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateCloudWatchIDS(t *testing.T) {
	ids, err := GenerateCloudWatchIDS()
	if err != nil {
		t.Fatalf("GenerateCloudWatchIDS() error = %v", err)
	}

	matchedHandler, _ := regexp.MatchString(HandlerRegx, ids.Handler)
	if !matchedHandler {
		t.Errorf("Handler %q does not match pattern %s", ids.Handler, HandlerRegx)
	}

	matchedTarget, _ := regexp.MatchString(TargetRegx, ids.Target)
	if !matchedTarget {
		t.Errorf("Target %q does not match pattern %s", ids.Target, TargetRegx)
	}
}
