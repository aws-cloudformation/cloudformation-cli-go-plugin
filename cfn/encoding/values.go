package encoding

import (
	"reflect"
)

var zeroValue reflect.Value

var interfaceType = reflect.TypeOf((*interface{})(nil)).Elem()
var stringType = reflect.TypeOf("")
