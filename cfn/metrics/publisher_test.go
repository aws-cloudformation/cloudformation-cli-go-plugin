package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

const succeed = "\u2713"
const failed = "\u2717"

type MockCloudWatchClient struct {
	MetricName string
	Unit       cwtypes.StandardUnit
	Value      float64
	Dim        map[string]string
}

func NewMockCloudWatchClient() *MockCloudWatchClient {
	return &MockCloudWatchClient{}
}

func (m *MockCloudWatchClient) PutMetricData(ctx context.Context, in *cloudwatch.PutMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error) {
	d := make(map[string]string)
	for _, v := range in.MetricData[0].Dimensions {
		d[*v.Name] = *v.Value
	}
	m.MetricName = *in.MetricData[0].MetricName
	m.Unit = in.MetricData[0].Unit
	m.Value = *in.MetricData[0].Value
	m.Dim = d
	return nil, nil
}

type MockCloudWatchClientError struct{}

func NewMockCloudWatchClientError() *MockCloudWatchClientError {
	return &MockCloudWatchClientError{}
}

func (m *MockCloudWatchClientError) PutMetricData(ctx context.Context, in *cloudwatch.PutMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error) {
	return nil, errors.New("Error")
}

func TestPublisher_PublishExceptionMetric(t *testing.T) {
	type fields struct {
		Client  CloudWatchClient
		resName string
	}
	type args struct {
		date   time.Time
		action string
		e      error
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantAction string
		wantExc    string
		wantRes    string
		wantMetric string
		wantUnit   cwtypes.StandardUnit
		wantValue  float64
	}{
		{"exception", fields{NewMockCloudWatchClient(), "foo::bar::test"}, args{time.Now(), "CREATE", errors.New("failed to create\nresource")}, false, "CREATE", "failed to create resource", "foo/bar/test", "HandlerException", cwtypes.StandardUnitCount, 1.0},
		{"exception error", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "CREATE", errors.New("failed")}, true, "", "", "", "", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.fields.Client, tt.fields.resName)
			p.PublishExceptionMetric(tt.args.date, tt.args.action, tt.args.e)
			if !tt.wantErr {
				e := tt.fields.Client.(*MockCloudWatchClient)
				if e.Dim[DimensionKeyAcionType] != tt.wantAction {
					t.Errorf("Action: got %v, want %v", e.Dim[DimensionKeyAcionType], tt.wantAction)
				}
				if e.Dim[DimensionKeyExceptionType] != tt.wantExc {
					t.Errorf("ExceptionType: got %v, want %v", e.Dim[DimensionKeyExceptionType], tt.wantExc)
				}
				if e.Dim[DimensionKeyResourceType] != tt.wantRes {
					t.Errorf("ResourceType: got %v, want %v", e.Dim[DimensionKeyResourceType], tt.wantRes)
				}
				if e.MetricName != tt.wantMetric {
					t.Errorf("MetricName: got %v, want %v", e.MetricName, tt.wantMetric)
				}
				if e.Unit != tt.wantUnit {
					t.Errorf("Unit: got %v, want %v", e.Unit, tt.wantUnit)
				}
				if e.Value != tt.wantValue {
					t.Errorf("Value: got %v, want %v", e.Value, tt.wantValue)
				}
			}
		})
	}
}

func TestPublisher_PublishInvocationMetric(t *testing.T) {
	p := New(NewMockCloudWatchClient(), "foo::bar::test")
	p.PublishInvocationMetric(time.Now(), "CREATE")
	e := p.client.(*MockCloudWatchClient)
	if e.MetricName != MetricNameHanderInvocationCount {
		t.Errorf("MetricName: got %v, want %v", e.MetricName, MetricNameHanderInvocationCount)
	}
	if e.Unit != cwtypes.StandardUnitCount {
		t.Errorf("Unit: got %v, want %v", e.Unit, cwtypes.StandardUnitCount)
	}
}

func TestPublisher_PublishDurationMetric(t *testing.T) {
	p := New(NewMockCloudWatchClient(), "foo::bar::test")
	p.PublishDurationMetric(time.Now(), "CREATE", 15.0)
	e := p.client.(*MockCloudWatchClient)
	if e.MetricName != MetricNameHanderDuration {
		t.Errorf("MetricName: got %v, want %v", e.MetricName, MetricNameHanderDuration)
	}
	if e.Unit != cwtypes.StandardUnitMilliseconds {
		t.Errorf("Unit: got %v, want %v", e.Unit, cwtypes.StandardUnitMilliseconds)
	}
	if e.Value != 15.0 {
		t.Errorf("Value: got %v, want %v", e.Value, 15.0)
	}
}

func TestPublisher_ResourceTypeName(t *testing.T) {
	r := ResourceTypeName("foo::bar::test")
	if r != "foo/bar/test" {
		t.Errorf("got %v, want foo/bar/test", r)
	}
}
