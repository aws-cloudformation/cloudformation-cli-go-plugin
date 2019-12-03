package cfn

import (
	"fmt"
	"log"
	"time"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/callback"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/metrics"
)

//reportErr is an unexported struct that handles reporting of errors.
type reportErr struct {
	callbackAdapter  *callback.CloudFormationCallbackAdapter
	metricsPublisher *metrics.Publisher
	publishStatus    bool
}

//NewReportErr is a factory func that returns a pointer to a struct
func newReportErr(callbackAdapter *callback.CloudFormationCallbackAdapter, metricsPublisher *metrics.Publisher) *reportErr {
	return &reportErr{
		callbackAdapter:  callbackAdapter,
		metricsPublisher: metricsPublisher,
		publishStatus:    false,
	}
}

//Report publishes errors and reports error status to Cloudformation.
func (r *reportErr) report(event *event, message string, err error, errCode string) (response, error) {
	m := fmt.Sprintf("Unable to complete request; %s error", message)

	if isMutatingAction(event.Action) && r.publishStatus {
		if reportErr := r.callbackAdapter.ReportFailureStatus(event.RequestData.ResourceProperties, cfnerr.InternalFailure, err); reportErr != nil {
			log.Printf("Callback report error; Error: %s", reportErr.Error())
		}
	}
	r.metricsPublisher.PublishExceptionMetric(time.Now(), string(event.Action), err)

	return newFailedResponse(cfnerr.New(serviceInternalError, m, err), event.BearerToken), err
}

func (r *reportErr) setPublishSatus(report bool) {
	r.publishStatus = report
}
