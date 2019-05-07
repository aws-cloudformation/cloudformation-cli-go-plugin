package resource

import (
	"encoding/json"
	"fmt"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy"
)

//CreateRequest handles the Create event from the Cloudformation service.
//The DesiredResourceState and PreviousResourceState are Unmarshal into a resource struct and
//made avalible as an interface{}. The interfaces needs Type assertions be for you can access
//the stored values.
func (r *{{ model_name|uppercase_first_letter }}) CreateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	//***Add code here: Make your API call, modify the model, etc..
	//Example: printing the DesiredResourceState resouce value
	rJSON, _ := json.MarshalIndent(r.DesiredResourceState, "", "\t")
	fmt.Println(string(rJSON))

	//Example: printing the PreviousResourceState resouce value
	pJSON, _ := json.MarshalIndent(r.PreviousResourceState, "", "\t")
	fmt.Println(string(pJSON))
	//***

	r.DesiredResourceState.SecondCopyOfMemo = "test"

	//return the status
	return &proxy.ProgressEvent{

		OperationStatus:      proxy.Complete,
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
		ResourceModel:        r.DesiredResourceState,
	}, nil
}

//DeleteRequest handles the Delete event from the Cloudformation service.
//The DesiredResourceState and PreviousResourceState are Unmarshal into a resource struct and
//made avalible as an interface{}. The interfaces needs Type assertions be for you can access
//the stored values.
func (r *{{ model_name|uppercase_first_letter }}) DeleteRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	//***Add code here: Make your API call, modify the model, etc..

	//Example of printing the DesiredResourceState resouce value
	rJSON, _ := json.MarshalIndent(r.DesiredResourceState, "", "\t")
	fmt.Println(string(rJSON))

	//Example: printing the PreviousResourceState resouce value
	//pJSON, _ := json.MarshalIndent(r.PreviousResourceState, "", "\t")
	//fmt.Println(string(pJSON))
	//***

	//return the status
	return &proxy.ProgressEvent{
		OperationStatus:      proxy.Complete,
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

//ListRequest handles the List event from the Cloudformation service.
//The DesiredResourceState and PreviousResourceState are Unmarshal into a resource struct and
//made avalible as an interface{}. The interfaces needs Type assertions be for you can access
//the stored values.
func (r *{{ model_name|uppercase_first_letter }}) ListRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	//***Add code here: Make your API call, modify the model, etc..

	//Example of printing the DesiredResourceState resouce value
	rJSON, _ := json.MarshalIndent(r.DesiredResourceState, "", "\t")
	fmt.Println(string(rJSON))

	//Example: printing the PreviousResourceState resouce value
	//pJSON, _ := json.MarshalIndent(r.PreviousResourceState, "", "\t")
	//fmt.Println(string(pJSON))
	//***

	//return the status
	return &proxy.ProgressEvent{
		OperationStatus:      proxy.Complete,
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

//ReadRequest handles the Read event from the Cloudformation service.
//The DesiredResourceState and PreviousResourceState are Unmarshal into a resource struct and
//made avalible as an interface{}. The interfaces needs Type assertions be for you can access
//the stored values.
func (r *{{ model_name|uppercase_first_letter }}) ReadRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	//***Add code here: Make your API call, modify the model, etc..

	//Example of printing the DesiredResourceState resouce value
	rJSON, _ := json.MarshalIndent(r.DesiredResourceState, "", "\t")
	fmt.Println(string(rJSON))

	//Example of printing the PreviousResourceState resouce value
	//pJSON, _ := json.MarshalIndent(r.PreviousResourceState, "", "\t")
	//fmt.Println(string(pJSON))
	//***

	//return the status
	return &proxy.ProgressEvent{
		OperationStatus:      proxy.Complete,
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
	}, nil
}

//UpdateRequest handles the Update event from the Cloudformation service.
//The DesiredResourceState and PreviousResourceState are Unmarshal into a resource struct and
//made avalible as an interface{}. The interfaces needs Type assertions be for you can access
//the stored values.
func (r *{{ model_name|uppercase_first_letter }}) UpdateRequest(request *proxy.ResourceHandlerRequest, callbackContext proxy.RequestContext) (*proxy.ProgressEvent, error) {

	//***Add code here: Make your API call, modify the model, etc..
	//Example of printing the DesiredResourceState resouce value
	rJSON, _ := json.MarshalIndent(r.DesiredResourceState, "", "\t")
	fmt.Println(string(rJSON))

	//Example: printing the PreviousResourceState resouce value
	pJSON, _ := json.MarshalIndent(r.PreviousResourceState, "", "\t")
	fmt.Println(string(pJSON))
	//***

	r.DesiredResourceState.SecondCopyOfMemo = "test"

	//return the status
	return &proxy.ProgressEvent{
		OperationStatus:      proxy.Complete,
		CallbackContext:      callbackContext.CallbackContext,
		CallbackDelayMinutes: 0,
		ResourceModel:        r.DesiredResourceState,
	}, nil
}
