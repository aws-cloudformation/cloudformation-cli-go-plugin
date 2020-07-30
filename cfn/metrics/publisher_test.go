package metrics

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

const succeed = "\u2713"
const failed = "\u2717"

// Define a mock struct to be used in your unit tests of myFunc.
type MockCloudWatchClient struct {
	cloudwatchiface.CloudWatchAPI
	MetricName string
	Unit       string
	Value      float64
	Dim        map[string]string
}

func NewMockCloudWatchClient() *MockCloudWatchClient {
	return &MockCloudWatchClient{}
}

func (m *MockCloudWatchClient) PutMetricData(in *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {

	// copy dimension in to a map for searching

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
type MockCloudWatchClientError struct {
	cloudwatchiface.CloudWatchAPI
}

func NewMockCloudWatchClientError() *MockCloudWatchClientError {
	return &MockCloudWatchClientError{}
}

func (m *MockCloudWatchClientError) PutMetricData(in *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {

	return nil, errors.New("Error")
}
func TestPublisher_PublishExceptionMetric(t *testing.T) {
	type fields struct {
		Client  cloudwatchiface.CloudWatchAPI
		resName string
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
		wantDimensionKeyResourceType  string
		wantMetricName                string
		wantUnit                      string
		wantValue                     float64
	}{
		{"testPublisherPublishExceptionMetric", fields{NewMockCloudWatchClient(), "foo::bar::test"}, args{time.Now(), "CREATE", errors.New("failed to create\nresource")}, "HandlerException", false, "CREATE", "failed to create resource", "foo/bar/test", "HandlerException", cloudwatch.StandardUnitCount, 1.0},
		{"testPublisherPublishExceptionMetricWantError", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "CREATE", errors.New("failed to create resource")}, "HandlerException", true, "CREATE", "failed to create resource", "foo/bar/test", "HandlerException", cloudwatch.StandardUnitCount, 1.0},
		{"testPublisherPublishExceptionMetric", fields{NewMockCloudWatchClient(), "foo::bar::test"}, args{time.Now(), "UPDATE", errors.New("failed to create resource")}, "HandlerException", false, "UPDATE", "failed to create resource", "foo/bar/test", "HandlerException", cloudwatch.StandardUnitCount, 1.0},
		{"testPublisherPublishExceptionMetricWantError", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "UPDATE", errors.New("failed to create resource")}, "HandlerException", true, "UPDATE", "failed to create resource", "foo/bar/test", "HandlerException", cloudwatch.StandardUnitCount, 1.0},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.fields.Client, tt.fields.resName)
			t.Logf("\tTest: %d\tWhen checking %q for success", i, tt.name)
			{
				p.PublishExceptionMetric(tt.args.date, tt.args.action, tt.args.e)
				t.Logf("\t%s\tShould be able to make the PublishExceptionMetric call.", succeed)
				if !tt.wantErr {
					e := tt.fields.Client.(*MockCloudWatchClient)

					if e.Dim[DimensionKeyAcionType] == tt.wantAction {
						t.Logf("\t%s\tAction should be (%v).", succeed, tt.wantAction)
					} else {
						t.Errorf("\t%s\tAction should be (%v). : %v", failed, tt.wantAction, e.Dim[DimensionKeyAcionType])
					}

					if e.Dim[DimensionKeyExceptionType] == tt.wantDimensionKeyExceptionType {
						t.Logf("\t%s\tDimensionKeyExceptionType should be (%v).", succeed, tt.wantDimensionKeyExceptionType)
					} else {
						t.Errorf("\t%s\tDimensionKeyExceptionType should be (%v). : %v", failed, tt.wantDimensionKeyExceptionType, e.Dim[DimensionKeyExceptionType])
					}

					if e.Dim[DimensionKeyResourceType] == tt.wantDimensionKeyResourceType {
						t.Logf("\t%s\t DimensionKeyResourceType should be (%v).", succeed, tt.wantDimensionKeyResourceType)
					} else {
						t.Errorf("\t%s\tDimensionKeyResourceType should be (%v). : %v", failed, tt.wantDimensionKeyResourceType, e.Dim[DimensionKeyResourceType])
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
		Client  cloudwatchiface.CloudWatchAPI
		resName string
	}
	type args struct {
		date   time.Time
		action string
	}
	tests := []struct {
		name                         string
		fields                       fields
		args                         args
		MetricName                   string
		wantErr                      bool
		wantAction                   string
		wantDimensionKeyResourceType string
		wantMetricName               string
		wantUnit                     string
		wantValue                    float64
	}{
		{"testPublishInvocationMetric", fields{NewMockCloudWatchClient(), "foo::bar::test"}, args{time.Now(), "CREATE"}, "HandlerInvocationCount", false, "CREATE", "foo/bar/test", "HandlerInvocationCount", cloudwatch.StandardUnitCount, 1.0},
		{"testPublishInvocationMetricWantError", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "CREATE"}, "HandlerException", true, "CREATE", "foo/bar/test", "HandlerException", cloudwatch.StandardUnitCount, 1.0},
		{"testPublishInvocationMetric", fields{NewMockCloudWatchClient(), "foo::bar::test"}, args{time.Now(), "UPDATE"}, "HandlerInvocationCount", false, "UPDATE", "foo/bar/test", "HandlerInvocationCount", cloudwatch.StandardUnitCount, 1.0},
		{"testPublishInvocationMetricError", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "UPDATE"}, "HandlerException", true, "UPDATE", "foo/bar/test", "HandlerInvocationCount", cloudwatch.StandardUnitCount, 1.0},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.fields.Client, tt.fields.resName)
			t.Logf("\tTest: %d\tWhen checking %q for success", i, tt.name)
			{
				p.PublishInvocationMetric(tt.args.date, tt.args.action)
				t.Logf("\t%s\tShould be able to make the PublishInvocationMetric call.", succeed)
				if !tt.wantErr {
					e := tt.fields.Client.(*MockCloudWatchClient)

					if e.Dim[DimensionKeyAcionType] == tt.wantAction {
						t.Logf("\t%s\tAction should be (%v).", succeed, tt.wantAction)
					} else {
						t.Errorf("\t%s\tAction should be (%v). : %v", failed, tt.wantAction, e.Dim[DimensionKeyAcionType])
					}

					if e.Dim[DimensionKeyResourceType] == tt.wantDimensionKeyResourceType {
						t.Logf("\t%s\t DimensionKeyResourceType should be (%v).", succeed, tt.wantDimensionKeyResourceType)
					} else {
						t.Errorf("\t%s\tDimensionKeyResourceType should be (%v). : %v", failed, tt.wantDimensionKeyResourceType, e.Dim[DimensionKeyResourceType])
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
		Client  cloudwatchiface.CloudWatchAPI
		resName string
	}
	type args struct {
		date   time.Time
		action string
		sec    float64
	}
	tests := []struct {
		name                         string
		fields                       fields
		args                         args
		MetricName                   string
		wantErr                      bool
		wantAction                   string
		wantDimensionKeyResourceType string
		wantMetricName               string
		wantUnit                     string
		wantValue                    float64
	}{
		{"testPublishInvocationMetric", fields{NewMockCloudWatchClient(), "foo::bar::test"}, args{time.Now(), "CREATE", 15.0}, "HandlerInvocationDuration", false, "CREATE", "foo/bar/test", "HandlerInvocationDuration", cloudwatch.StandardUnitMilliseconds, 15},
		{"testPublishInvocationMetricWantError", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "CREATE", 15.0}, "HandlerInvocationDuration", true, "CREATE", "foo/bar/test", "HandlerInvocationDuration", cloudwatch.StandardUnitMilliseconds, 15},
		{"testPublishInvocationMetric", fields{NewMockCloudWatchClient(), "foo::bar::test"}, args{time.Now(), "UPDATE", 15.0}, "HandlerInvocationDuration", false, "UPDATE", "foo/bar/test", "HandlerInvocationDuration", cloudwatch.StandardUnitMilliseconds, 15},
		{"testPublishInvocationMetricError", fields{NewMockCloudWatchClientError(), "foo::bar::test"}, args{time.Now(), "UPDATE", 15.0}, "HandlerInvocationDuration", true, "UPDATE", "foo/bar/test", "HandlerInvocationDuration", cloudwatch.StandardUnitMilliseconds, 15},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.fields.Client, tt.fields.resName)
			t.Logf("\tTest: %d\tWhen checking %q for success", i, tt.name)
			{
				p.PublishDurationMetric(tt.args.date, tt.args.action, tt.args.sec)
				t.Logf("\t%s\tShould be able to make the PublishDurationMetric call.", succeed)
				if !tt.wantErr {
					e := tt.fields.Client.(*MockCloudWatchClient)

					if e.Dim[DimensionKeyAcionType] == tt.wantAction {
						t.Logf("\t%s\tAction should be (%v).", succeed, tt.wantAction)
					} else {
						t.Errorf("\t%s\tAction should be (%v). : %v", failed, tt.wantAction, e.Dim[DimensionKeyAcionType])
					}

					if e.Dim[DimensionKeyResourceType] == tt.wantDimensionKeyResourceType {
						t.Logf("\t%s\t DimensionKeyResourceType should be (%v).", succeed, tt.wantDimensionKeyResourceType)
					} else {
						t.Errorf("\t%s\tDimensionKeyResourceType should be (%v). : %v", failed, tt.wantDimensionKeyResourceType, e.Dim[DimensionKeyResourceType])
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

func TestPublisher_ResourceTypeName(t *testing.T) {
	type args struct {
		t string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test foo", args{"foo::bar::test"}, "foo/bar/test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ResourceTypeName(tt.args.t)

			if r != tt.want {
				t.Errorf("Should be %v : got %v", tt.want, r)
				return
			}

		})
	}
}
