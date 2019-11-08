package main

import (
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/examples/s3-object/cmd/resource"
)

/*
This file is autogenerated, do not edit;
changes will be undone by the next 'generate' command.
*/

// main is the entry point of the applicaton.
func main() {
	cfn.Start(&resource.Handler{})
}
