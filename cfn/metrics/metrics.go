package metrics

import "strings"

const (
	// MetricNameSpaceRoot is the Metric name space root.
	MetricNameSpaceRoot = "AWS/CloudFormation"
	//MetricNameHanderException  is a metric type.
	MetricNameHanderException = "HandlerException"
	//MetricNameHanderDuration is a metric type.
	MetricNameHanderDuration = "HandlerInvocationDuration"
	//MetricNameHanderInvocationCount is a metric type.
	MetricNameHanderInvocationCount = "HandlerInvocationCount"
	//DimensionKeyAcionType  is the Action key in the dimension.
	DimensionKeyAcionType = "Action"
	//DimensionKeyExceptionType  is the ExceptionType in the dimension.
	DimensionKeyExceptionType = "ExceptionType"
	//DimensionKeyResouceType  is the ResourceType in the dimension.
	DimensionKeyResouceType = "ResourceType"
	//ServiceInternalError ...
	ServiceInternalError string = "ServiceInternal"
)

// ResourceTypeName returns a type name by removing (::) and replaing with (/)
//
// Example
//
// 	r := metrics.ResourceTypeName("AWS::Service::Resource")
//
// 	// Will return "AWS/Service/Resource"
func ResourceTypeName(t string) string {
	return strings.ReplaceAll(t, "::", "/")

}
