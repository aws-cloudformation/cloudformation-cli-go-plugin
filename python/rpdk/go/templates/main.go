package main

import (
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/proxy"
	{{ path }}
)

func main() {

	r := resource.{{ model_name }}{}

	proxy.Start(r)

}
