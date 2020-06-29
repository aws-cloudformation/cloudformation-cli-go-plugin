package cfn

import (
	"fmt"
	"time"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/metrics"
)

// reportErr is an unexported struct that handles reporting of errors.
type reportErr struct {
	metricsPublisher *metrics.Publisher
}

// NewReportErr is a factory func that returns a pointer to a struct
func newReportErr(metricsPublisher *metrics.Publisher) *reportErr {
	return &reportErr{
		metricsPublisher: metricsPublisher,
	}
}

// Report publishes errors and reports error status to Cloudformation.
func (r *reportErr) report(event *event, message string, err error, errCode string) (response, error) {
	m := fmt.Sprintf("Unable to complete request; %s error", message)
	r.metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), err)
	return newFailedResponse(cfnerr.New(serviceInternalError, m, err), event.BearerToken), err
}
