package metric_test

import (
	"errors"
	"testing"
	"time"

	"github.com/aws-cloudformation-rpdk-go-plugin/internal/metric"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

const succeed = "\u2713"
const failed = "\u2717"

// Define a mock struct to be used in your unit tests of myFunc.
type mockCloudWatchClient struct {
	cloudwatchiface.CloudWatchAPI
	MetricName string
	Unit       string
	Value      float64
	Dim        map[string]string
}

func New() *mockCloudWatchClient {
	return &mockCloudWatchClient{}
}

func (m *mockCloudWatchClient) PutMetricData(in *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {

	//copy dimension in to a map for searching

	d := make(map[string]string)

	for _, v := range in.MetricData[0].Dimensions {
		d[*v.Name] = *v.Value
	}

	m.MetricName = *in.MetricData[0].MetricName
	m.Unit = *in.MetricData[0].Unit
	m.Value = *in.MetricData[0].Value
	m.Dim = d

	return nil, nil
}

// Define a mock struct to be used in your unit tests of myFunc.
type mockCloudWatchClientError struct {
	cloudwatchiface.CloudWatchAPI
}

func NewMockCloudWatchClientError() *mockCloudWatchClientError {
	return &mockCloudWatchClientError{}
}

func (m *mockCloudWatchClientError) PutMetricData(in *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {

	return nil, errors.New("Error")
}
func TestPublisher_PublishExceptionMetric(t *testing.T) {
	type fields struct {
		Client    cloudwatchiface.CloudWatchAPI
		namespace string
	}
	type args struct {
		date   time.Time
		action string
		e      error
	}
	tests := []struct {
		name                          string
		fields                        fields
		args                          args
		MetricName                    string
		wantErr                       bool
		wantAction                    string
		wantDimensionKeyExceptionType string
		wantDimensionKeyResouceType   string
		wantMetricName                string
		wantUnit                      string
		wantValue                     float64
	}{
		{"testPublisherPublishExceptionMetric", fields{New(), "foo::bar::test"}, args{time.Now(), "CREATE", errors.New("failed to create resource")}, "HandlerException", false, "CREATE", "failed to create resource", "foo::bar::test", "HandlerException", cloudwatch.StandardUnitCount, 1.0},
		{"testPublisherPublishExceptionMetricWantError", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "CREATE", errors.New("failed to create resource")}, "HandlerException", true, "CREATE", "failed to create resource", "foo::bar::test", "HandlerException", cloudwatch.StandardUnitCount, 1.0},
		{"testPublisherPublishExceptionMetric", fields{New(), "foo::bar::test"}, args{time.Now(), "UPDATE", errors.New("failed to create resource")}, "HandlerException", false, "UPDATE", "failed to create resource", "foo::bar::test", "HandlerException", cloudwatch.StandardUnitCount, 1.0},
		{"testPublisherPublishExceptionMetricWantError", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "UPDATE", errors.New("failed to create resource")}, "HandlerException", true, "UPDATE", "failed to create resource", "foo::bar::test", "HandlerException", cloudwatch.StandardUnitCount, 1.0},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p := &metric.Publisher{
				Client:    tt.fields.Client,
				Namespace: tt.fields.namespace,
			}
			t.Logf("\tTest: %d\tWhen checking %q for success", i, tt.name)
			{
				if err := p.PublishExceptionMetric(tt.args.date, tt.args.action, tt.args.e); (err != nil) != tt.wantErr {
					t.Errorf("\t%s\tShould be able to make the PublishExceptionMetric call : %v", failed, err)
					return
				}
				t.Logf("\t%s\tShould be able to make the PublishExceptionMetric call.", succeed)
				if !tt.wantErr {
					e := tt.fields.Client.(*mockCloudWatchClient)

					if e.Dim[metric.DimensionKeyAcionType] == tt.wantAction {
						t.Logf("\t%s\tAction should be (%v).", succeed, tt.wantAction)
					} else {
						t.Errorf("\t%s\tAction should be (%v). : %v", failed, tt.wantAction, e.Dim[metric.DimensionKeyAcionType])
					}

					if e.Dim[metric.DimensionKeyExceptionType] == tt.wantDimensionKeyExceptionType {
						t.Logf("\t%s\tDimensionKeyExceptionType should be (%v).", succeed, tt.wantDimensionKeyExceptionType)
					} else {
						t.Errorf("\t%s\tDimensionKeyExceptionType should be (%v). : %v", failed, tt.wantDimensionKeyExceptionType, e.Dim[metric.DimensionKeyExceptionType])
					}

					if e.Dim[metric.DimensionKeyResouceType] == tt.wantDimensionKeyResouceType {
						t.Logf("\t%s\t DimensionKeyResouceType should be (%v).", succeed, tt.wantDimensionKeyResouceType)
					} else {
						t.Errorf("\t%s\tDimensionKeyResouceType should be (%v). : %v", failed, tt.wantDimensionKeyResouceType, e.Dim[metric.DimensionKeyResouceType])
					}

					if e.MetricName == tt.wantMetricName {
						t.Logf("\t%s\t MetricName should be (%v).", succeed, tt.wantMetricName)
					} else {
						t.Errorf("\t%s\tMetricName should be (%v). : %v", failed, tt.wantMetricName, e.MetricName)
					}

					if e.Unit == tt.wantUnit {
						t.Logf("\t%s\t Unit should be (%v).", succeed, tt.wantUnit)
					} else {
						t.Errorf("\t%s\tUnit should be (%v). : %v", failed, tt.wantUnit, e.Unit)
					}

					if e.Value == tt.wantValue {
						t.Logf("\t%s\t Unit should be (%v).", succeed, tt.wantValue)
					} else {
						t.Errorf("\t%s\tUnit should be (%v). : %v", failed, tt.wantValue, e.Value)
					}
				}
			}

		})
	}
}

func TestPublisher_PublishInvocationMetric(t *testing.T) {
	type fields struct {
		Client    cloudwatchiface.CloudWatchAPI
		namespace string
	}
	type args struct {
		date   time.Time
		action string
	}
	tests := []struct {
		name                        string
		fields                      fields
		args                        args
		MetricName                  string
		wantErr                     bool
		wantAction                  string
		wantDimensionKeyResouceType string
		wantMetricName              string
		wantUnit                    string
		wantValue                   float64
	}{
		{"testPublishInvocationMetric", fields{New(), "foo::bar::test"}, args{time.Now(), "CREATE"}, "HandlerInvocationCount", false, "CREATE", "foo::bar::test", "HandlerInvocationCount", cloudwatch.StandardUnitCount, 1.0},
		{"testPublishInvocationMetricWantError", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "CREATE"}, "HandlerException", true, "CREATE", "foo::bar::test", "HandlerException", cloudwatch.StandardUnitCount, 1.0},
		{"testPublishInvocationMetric", fields{New(), "foo::bar::test"}, args{time.Now(), "UPDATE"}, "HandlerInvocationCount", false, "UPDATE", "foo::bar::test", "HandlerInvocationCount", cloudwatch.StandardUnitCount, 1.0},
		{"testPublishInvocationMetricError", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "UPDATE"}, "HandlerException", true, "UPDATE", "foo::bar::test", "HandlerInvocationCount", cloudwatch.StandardUnitCount, 1.0},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p := &metric.Publisher{
				Client:    tt.fields.Client,
				Namespace: tt.fields.namespace,
			}
			t.Logf("\tTest: %d\tWhen checking %q for success", i, tt.name)
			{
				if err := p.PublishInvocationMetric(tt.args.date, tt.args.action); (err != nil) != tt.wantErr {
					t.Errorf("\t%s\tShould be able to make the PublishInvocationMetric call : %v", failed, err)
					return
				}
				t.Logf("\t%s\tShould be able to make the PublishInvocationMetric call.", succeed)
				if !tt.wantErr {
					e := tt.fields.Client.(*mockCloudWatchClient)

					if e.Dim[metric.DimensionKeyAcionType] == tt.wantAction {
						t.Logf("\t%s\tAction should be (%v).", succeed, tt.wantAction)
					} else {
						t.Errorf("\t%s\tAction should be (%v). : %v", failed, tt.wantAction, e.Dim[metric.DimensionKeyAcionType])
					}

					if e.Dim[metric.DimensionKeyResouceType] == tt.wantDimensionKeyResouceType {
						t.Logf("\t%s\t DimensionKeyResouceType should be (%v).", succeed, tt.wantDimensionKeyResouceType)
					} else {
						t.Errorf("\t%s\tDimensionKeyResouceType should be (%v). : %v", failed, tt.wantDimensionKeyResouceType, e.Dim[metric.DimensionKeyResouceType])
					}

					if e.MetricName == tt.wantMetricName {
						t.Logf("\t%s\t MetricName should be (%v).", succeed, tt.wantMetricName)
					} else {
						t.Errorf("\t%s\tMetricName should be (%v). : %v", failed, tt.wantMetricName, e.MetricName)
					}

					if e.Unit == tt.wantUnit {
						t.Logf("\t%s\t Unit should be (%v).", succeed, tt.wantUnit)
					} else {
						t.Errorf("\t%s\tUnit should be (%v). : %v", failed, tt.wantUnit, e.Unit)
					}

					if e.Value == tt.wantValue {
						t.Logf("\t%s\t Unit should be (%v).", succeed, tt.wantValue)
					} else {
						t.Errorf("\t%s\tUnit should be (%v). : %v", failed, tt.wantValue, e.Value)
					}
				}
			}

		})
	}

}

func TestPublisher_PublishDurationMetric(t *testing.T) {
	type fields struct {
		Client    cloudwatchiface.CloudWatchAPI
		namespace string
	}
	type args struct {
		date   time.Time
		action string
		sec    int64
	}
	tests := []struct {
		name                        string
		fields                      fields
		args                        args
		MetricName                  string
		wantErr                     bool
		wantAction                  string
		wantDimensionKeyResouceType string
		wantMetricName              string
		wantUnit                    string
		wantValue                   float64
	}{
		{"testPublishInvocationMetric", fields{New(), "foo::bar::test"}, args{time.Now(), "CREATE", 15.0}, "HandlerInvocationDuration", false, "CREATE", "foo::bar::test", "HandlerInvocationDuration", cloudwatch.StandardUnitMilliseconds, 15},
		{"testPublishInvocationMetricWantError", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "CREATE", 15.0}, "HandlerInvocationDuration", true, "CREATE", "foo::bar::test", "HandlerInvocationDuration", cloudwatch.StandardUnitMilliseconds, 15},
		{"testPublishInvocationMetric", fields{New(), "foo::bar::test"}, args{time.Now(), "UPDATE", 15.0}, "HandlerInvocationDuration", false, "UPDATE", "foo::bar::test", "HandlerInvocationDuration", cloudwatch.StandardUnitMilliseconds, 15},
		{"testPublishInvocationMetricError", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "UPDATE", 15.0}, "HandlerInvocationDuration", true, "UPDATE", "foo::bar::test", "HandlerInvocationDuration", cloudwatch.StandardUnitMilliseconds, 15},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p := &metric.Publisher{
				Client:    tt.fields.Client,
				Namespace: tt.fields.namespace,
			}
			t.Logf("\tTest: %d\tWhen checking %q for success", i, tt.name)
			{
				if err := p.PublishDurationMetric(tt.args.date, tt.args.action, tt.args.sec); (err != nil) != tt.wantErr {
					t.Errorf("\t%s\tShould be able to make the PublishDurationMetric call : %v", failed, err)
					return
				}
				t.Logf("\t%s\tShould be able to make the PublishDurationMetric call.", succeed)
				if !tt.wantErr {
					e := tt.fields.Client.(*mockCloudWatchClient)

					if e.Dim[metric.DimensionKeyAcionType] == tt.wantAction {
						t.Logf("\t%s\tAction should be (%v).", succeed, tt.wantAction)
					} else {
						t.Errorf("\t%s\tAction should be (%v). : %v", failed, tt.wantAction, e.Dim[metric.DimensionKeyAcionType])
					}

					if e.Dim[metric.DimensionKeyResouceType] == tt.wantDimensionKeyResouceType {
						t.Logf("\t%s\t DimensionKeyResouceType should be (%v).", succeed, tt.wantDimensionKeyResouceType)
					} else {
						t.Errorf("\t%s\tDimensionKeyResouceType should be (%v). : %v", failed, tt.wantDimensionKeyResouceType, e.Dim[metric.DimensionKeyResouceType])
					}

					if e.MetricName == tt.wantMetricName {
						t.Logf("\t%s\t MetricName should be (%v).", succeed, tt.wantMetricName)
					} else {
						t.Errorf("\t%s\tMetricName should be (%v). : %v", failed, tt.wantMetricName, e.MetricName)
					}

					if e.Unit == tt.wantUnit {
						t.Logf("\t%s\t Unit should be (%v).", succeed, tt.wantUnit)
					} else {
						t.Errorf("\t%s\tUnit should be (%v). : %v", failed, tt.wantUnit, e.Unit)
					}

					if e.Value == tt.wantValue {
						t.Logf("\t%s\t Unit should be (%v).", succeed, tt.wantValue)
					} else {
						t.Errorf("\t%s\tUnit should be (%v). : %v", failed, tt.wantValue, e.Value)
					}
				}
			}

		})
	}

}
