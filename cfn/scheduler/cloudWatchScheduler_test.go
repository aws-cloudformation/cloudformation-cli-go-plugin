package scheduler

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
)

const succeed = "\u2713"
const failed = "\u2717"

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
		minutesFromNow int
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
					t.Logf("\t%s\tOneTimeCronExpression match should be (%v).", succeed, tt.want)
				} else {
					t.Errorf("\t%s\tOneTimeCronExpression match should be (%v). : %v", failed, tt.want, got)
				}
			}
		})
	}
}

func TestCloudWatchSchedulerRescheduleAfterMinutes(t *testing.T) {

	var cb = `{ string: "Foo"}`

	type fields struct {
		Client cloudwatcheventsiface.CloudWatchEventsAPI
	}
	type args struct {
		ctx             context.Context
		secFromNow      int
		callbackContext string
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		wantErr            bool
		WantRegxRuleName   string
		WantRegxTargetName string
		WantRuleMatch      bool
		WantTargetMatch    bool
		computeLocal       bool
	}{
		{"TestCloudWatchScheduler56SecsComputeLocal", fields{NewMockEvents()}, args{"arn:aws:lambda:us-east-2:123456789:function:myproject", 15, cb, time.Now().Add(time.Second * time.Duration(1000))}, false, `reinvoke-handler-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, `reinvoke-target-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, true, true, true},
		{"TestCloudWatchScheduler56SecsComputeNotLocal", fields{NewMockEvents()}, args{"arn:aws:lambda:us-east-2:123456789:function:myproject", 15, cb, time.Now().Add(time.Second * time.Duration(16))}, false, `reinvoke-handler-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, `reinvoke-target-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, true, true, false},
		{"TestCloudWatchSchedulerARNMustHaveValue", fields{NewMockEvents()}, args{"", 15, cb, time.Now().Add(time.Second * time.Duration(16))}, true, `reinvoke-handler-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, `reinvoke-target-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, true, true, false},
		{"TestCloudWatchSchedulerLessThen0", fields{NewMockEvents()}, args{"arn:aws:lambda:us-east-2:123456789:function:myproject", -87, cb, time.Now().Add(time.Second * time.Duration(1000))}, true, `reinvoke-handler-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, `reinvoke-target-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, true, true, false},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.fields.Client.(*MockedEvents)
			t.Logf("\tTest: %d\tWhen checking %q for error status %v", i, tt.name, tt.wantErr)
			{
				c := &CloudWatchScheduler{
					client: tt.fields.Client,
				}
				cp, err := c.Reschedule(tt.args.arn, tt.args.secFromNow, cb, tt.args.deadline)
				if err != nil && !tt.wantErr {

					t.Errorf("\t%s\tShould be able to make the RescheduleAfterMinutes call : %v", failed, err)
					return
				}
				t.Logf("\t%s\tShould be able to make the RescheduleAfterMinutes call.", succeed)

				if cp == tt.computeLocal {
					t.Logf("\t%s\tCompute Local should be (%v).", succeed, tt.computeLocal)
				} else {
					t.Errorf("\t%s\tCompute Local should be (%v). : Value:%v", failed, tt.computeLocal, cp)
					return
				}

				if tt.wantErr == false && !tt.computeLocal {

					matchedRule, err := regexp.Match(tt.WantRegxRuleName, []byte(e.RuleName))

					if (err != nil) != tt.wantErr {
						t.Errorf("\t%s\tShould be able to make the Match call : %v", failed, err)
						return
					}
					t.Logf("\t%s\tShould be able to make the Matchcall.", succeed)

					if matchedRule == tt.WantRuleMatch {
						t.Logf("\t%s\tRule match should be (%v).", succeed, tt.WantRuleMatch)
					} else {
						t.Errorf("\t%s\tRule match should be (%v). : %v  Value:%s", failed, tt.WantRuleMatch, matchedRule, e.RuleName)
					}

					matchedTarget, err := regexp.Match(tt.WantRegxTargetName, []byte(e.TargetName))

					if (err != nil) != tt.wantErr {
						t.Errorf("\t%s\tShould be able to make the Match call : %v", failed, err)
						return
					}
					t.Logf("\t%s\tShould be able to make the Matchcall.", succeed)

					if matchedTarget == tt.WantTargetMatch {
						t.Logf("\t%s\tTarget match should be (%v).", succeed, tt.WantTargetMatch)
					} else {
						t.Errorf("\t%s\tTarget match should be (%v). : %v  Value: %s", failed, tt.WantRegxTargetName, matchedTarget, e.RuleName)
					}

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
		name               string
		fields             fields
		args               args
		wantErr            bool
		WantRegxRuleName   string
		WantRegxTargetName string
		WantRuleMatch      bool
		WantTargetMatch    bool
	}{
		{"TestCloudWatchRemove", fields{NewMockEvents()}, args{"reinvoke-handler-c51d7ba5-8eed-4226-99a6-6743f1169688", "reinvoke-target-c51d7ba5-8eed-4226-99a6-6743f1169688"}, false, `reinvoke-handler-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, `reinvoke-target-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, true, true},
		{"TestCloudWatchRemoveBlankCloudWatchEventsRuleName", fields{NewMockEvents()}, args{"", "reinvoke-target-c51d7ba5-8eed-4226-99a6-6743f1169688"}, true, `reinvoke-handler-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, `reinvoke-target-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, true, true},
		{"TestCloudWatchRemoveBlankcloudWatchEventsTargetID", fields{NewMockEvents()}, args{"reinvoke-handler-c51d7ba5-8eed-4226-99a6-6743f1169688", ""}, true, `reinvoke-handler-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, `reinvoke-target-([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}`, true, true},
	}
	for i, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			e := tt.fields.Client.(*MockedEvents)
			t.Logf("\tTest: %d\tWhen checking %q for error status %v", i, tt.name, tt.wantErr)
			{
				c := &CloudWatchScheduler{
					client: tt.fields.Client,
				}
				if err := c.CleanupCloudWatchEvents(tt.args.cloudWatchEventsRuleName, tt.args.cloudWatchEventsTargetID); (err != nil) != tt.wantErr {
					t.Errorf("\t%s\tShould be able to make the cloudWatchEventsRuleName call : %v", failed, err)
					return
				}
				t.Logf("\t%s\tShould be able to make the cloudWatchEventsRuleName call.", succeed)
				if tt.wantErr == false {

					matchedRule, err := regexp.Match(tt.WantRegxRuleName, []byte(e.RuleName))

					if (err != nil) != tt.wantErr {
						t.Errorf("\t%s\tShould be able to make the Match call : %v", failed, err)
						return
					}
					t.Logf("\t%s\tShould be able to make the Matchcall.", succeed)

					if matchedRule == tt.WantRuleMatch {
						t.Logf("\t%s\tRule match should be (%v).", succeed, tt.WantRuleMatch)
					} else {
						t.Errorf("\t%s\tRule match should be (%v). : %v  Value:%s", failed, tt.WantRuleMatch, matchedRule, e.RuleName)
					}

					matchedTarget, err := regexp.Match(tt.WantRegxTargetName, []byte(e.TargetName))

					if (err != nil) != tt.wantErr {
						t.Errorf("\t%s\tShould be able to make the Match call : %v", failed, err)
						return
					}
					t.Logf("\t%s\tShould be able to make the Matchcall.", succeed)

					if matchedTarget == tt.WantTargetMatch {
						t.Logf("\t%s\tTarget match should be (%v).", succeed, tt.WantTargetMatch)
					} else {
						t.Errorf("\t%s\tTarget match should be (%v). : %v  Value: %s", failed, tt.WantRegxTargetName, matchedTarget, e.RuleName)
					}

				}

			}
		})
	}

}
