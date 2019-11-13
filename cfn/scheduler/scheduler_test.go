package scheduler

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/logging"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
)

const (
	TargetRegx  string = `reinvoke-target-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`
	HandlerRegx string = `reinvoke-handler-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`
	Succeed     string = "\u2713"
	Failed      string = "\u2717"
	Arn         string = "arn:aws:lambda:us-east-2:123456789:function:myproject"
)

//MockedEvents mocks the call to AWS CloudWatch Events
type MockedEvents struct {
	cloudwatcheventsiface.CloudWatchEventsAPI
	RuleName   string
	TargetName string
}

func NewMockEvents() *MockedEvents {
	return &MockedEvents{}
}

func (m *MockedEvents) PutRule(in *cloudwatchevents.PutRuleInput) (*cloudwatchevents.PutRuleOutput, error) {
	m.RuleName = *in.Name
	return nil, nil
}

func (m *MockedEvents) PutTargets(in *cloudwatchevents.PutTargetsInput) (*cloudwatchevents.PutTargetsOutput, error) {
	m.TargetName = *in.Targets[0].Id
	return nil, nil

}

func (m *MockedEvents) DeleteRule(in *cloudwatchevents.DeleteRuleInput) (*cloudwatchevents.DeleteRuleOutput, error) {
	m.RuleName = *in.Name
	return nil, nil
}

func (m *MockedEvents) RemoveTargets(in *cloudwatchevents.RemoveTargetsInput) (*cloudwatchevents.RemoveTargetsOutput, error) {
	m.TargetName = *in.Ids[0]
	return nil, nil
}

func TestGenerateOneTimeCronExpression(t *testing.T) {
	type args struct {
		minutesFromNow int64
		t              time.Time
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"TestCreateOneTimeCronExpression", args{0, time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)}, "cron(34 20 17 11 ? 2009)"},
		{"TestCreateOneTimeCronExpression", args{0, time.Date(2001, 5, 25, 11, 04, 14, 651387237, time.UTC)}, "cron(04 11 25 05 ? 2001)"},
		{"TestCreateOneTimeCronExpression", args{0, time.Date(2006, 7, 17, 7, 18, 23, 651387237, time.UTC)}, "cron(18 07 17 07 ? 2006)"},
		{"TestCreateOneTimeCronExpression", args{0, time.Date(1999, 2, 07, 21, 28, 45, 651387237, time.UTC)}, "cron(28 21 07 02 ? 1999)"},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\tTest: %d\tWhen checking %q for match status %v", i, tt.name, tt.want)
			{
				got := GenerateOneTimeCronExpression(tt.args.minutesFromNow, tt.args.t)

				if got == tt.want {
					t.Logf("\t%s\tOneTimeCronExpression match should be (%v).", Succeed, tt.want)
				} else {
					t.Errorf("\t%s\tOneTimeCronExpression match should be (%v). : %v", Failed, tt.want, got)
				}
			}
		})
	}
}

func TestCloudWatchSchedulerRescheduleAfterMinutes(t *testing.T) {

	var cb = `{ string: "Foo"}`
	ct1, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*time.Duration(1000)))
	defer cancel()
	ct2, cancel2 := context.WithDeadline(context.Background(), time.Now().Add(time.Second*time.Duration(16)))
	defer cancel2()

	type fields struct {
		Client cloudwatcheventsiface.CloudWatchEventsAPI
	}
	type args struct {
		ctx             context.Context
		secFromNow      int64
		callbackContext string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantErr      bool
		computeLocal bool
	}{
		{"TestCloudWatchScheduler56SecsComputeLocal", fields{NewMockEvents()}, args{lambdacontext.NewContext(ct1, &lambdacontext.LambdaContext{InvokedFunctionArn: Arn}), 15, cb}, false, true},
		{"TestCloudWatchScheduler56SecsComputeNotLocal", fields{NewMockEvents()}, args{lambdacontext.NewContext(ct2, &lambdacontext.LambdaContext{InvokedFunctionArn: Arn}), 15, cb}, false, false},
		{"TestCloudWatchSchedulerLessThen0", fields{NewMockEvents()}, args{lambdacontext.NewContext(ct1, &lambdacontext.LambdaContext{InvokedFunctionArn: Arn}), -87, cb}, true, false},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\tTest: %d\tWhen checking %q for error status %v", i, tt.name, tt.wantErr)
			{
				c := &Scheduler{
					client: tt.fields.Client,
					logger: logging.New("scheduler: "),
				}

				ids, _ := GenerateCloudWatchIDS()
				cp, err := c.Reschedule(tt.args.ctx, tt.args.secFromNow, tt.args.callbackContext, ids)
				if err != nil &&
					tt.wantErr {

					t.Logf("\t%s\tShould be able to make the RescheduleAfterMinutes call.", Succeed)
					return
				}

				if cp.ComputeLocal == tt.computeLocal {
					t.Logf("\t%s\tCompute Local should be (%v).", Succeed, tt.computeLocal)
				} else {
					t.Errorf("\t%s\tCompute Local should be (%v). : Value:%v", Failed, tt.computeLocal, cp.ComputeLocal)
					return
				}
			}
		})
	}

}

func TestCloudWatchSchedulerCleanupCloudWatchEvents(t *testing.T) {
	type fields struct {
		Client cloudwatcheventsiface.CloudWatchEventsAPI
	}
	type args struct {
		cloudWatchEventsRuleName string
		cloudWatchEventsTargetID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"TestCloudWatchRemove", fields{NewMockEvents()}, args{"reinvoke-handler-c51d7ba5-8eed-4226-99a6-6743f1169688", "reinvoke-target-c51d7ba5-8eed-4226-99a6-6743f1169688"}, false},
		{"TestCloudWatchRemoveBlankCloudWatchEventsRuleName", fields{NewMockEvents()}, args{"", "reinvoke-target-c51d7ba5-8eed-4226-99a6-6743f1169688"}, true},
		{"TestCloudWatchRemoveBlankcloudWatchEventsTargetID", fields{NewMockEvents()}, args{"reinvoke-handler-c51d7ba5-8eed-4226-99a6-6743f1169688", ""}, true},
	}
	for i, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Logf("\tTest: %d\tWhen checking %q for error status %v", i, tt.name, tt.wantErr)
			{
				c := &Scheduler{
					client: tt.fields.Client,
					logger: logging.New("scheduler: "),
				}
				if err := c.CleanupEvents(tt.args.cloudWatchEventsRuleName, tt.args.cloudWatchEventsTargetID); (err != nil) != tt.wantErr {
					t.Errorf("\t%s\tShould be able to make the cloudWatchEventsRuleName call : %v", Failed, err)
					return
				}
				t.Logf("\t%s\tShould be able to make the cloudWatchEventsRuleName call.", Succeed)

			}
		})
	}

}

func TestGenerateCloudWatchIDS(t *testing.T) {
	tests := []struct {
		name               string
		WantRegxRuleName   string
		WantRegxTargetName string
		WantRuleMatch      bool
		WantTargetMatch    bool
		wantErr            bool
	}{
		{"Test Rule and Target pattern", HandlerRegx, TargetRegx, true, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateCloudWatchIDS()
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCloudWatchIDS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == false {

				matchedRule, err := regexp.Match(tt.WantRegxRuleName, []byte(got.Handler))

				if (err != nil) != tt.wantErr {
					t.Errorf("\t%s\tShould be able to make the Match call : %v", Failed, err)
					return
				}
				t.Logf("\t%s\tShould be able to make the Matchcall.", Succeed)

				if matchedRule == tt.WantRuleMatch {
					t.Logf("\t%s\tRule match should be (%v).", Succeed, tt.WantRuleMatch)
				} else {
					t.Errorf("\t%s\tRule match should be (%v). : %v", Failed, tt.WantRuleMatch, matchedRule)
				}

				matchedTarget, err := regexp.Match(tt.WantRegxTargetName, []byte(got.Target))

				if (err != nil) != tt.wantErr {
					t.Errorf("\t%s\tShould be able to make the Match call : %v", Failed, err)
					return
				}
				t.Logf("\t%s\tShould be able to make the Matchcall.", Succeed)

				if matchedTarget == tt.WantTargetMatch {
					t.Logf("\t%s\tTarget match should be (%v).", Succeed, tt.WantTargetMatch)
				} else {
					t.Errorf("\t%s\tTarget match should be (%v). : %v", Failed, tt.WantRegxTargetName, matchedTarget)
				}

			}
		})
	}
}
