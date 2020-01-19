package encoding

import (
	"reflect"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
)

// Strings
func toString(s string) string {
	return s
}

func toStringValue(s string) reflect.Value {
	return reflect.ValueOf(toString(s))
}

func toStringPtrValue(s string) reflect.Value {
	return reflect.ValueOf(aws.String(toString(s)))
}

// Bools
func toBool(s string) bool {
	val, _ := strconv.ParseBool(s)
	return val
}

func toBoolValue(s string) reflect.Value {
	return reflect.ValueOf(toBool(s))
}

func toBoolPtrValue(s string) reflect.Value {
	return reflect.ValueOf(aws.Bool(toBool(s)))
}

// Ints
func toInt(s string) int {
	val, _ := strconv.ParseInt(s, 0, 32)
	return int(val)
}

func toIntValue(s string) reflect.Value {
	return reflect.ValueOf(toInt(s))
}

func toIntPtrValue(s string) reflect.Value {
	return reflect.ValueOf(aws.Int(toInt(s)))
}

// Floats
func toFloat64(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

func toFloat64Value(s string) reflect.Value {
	return reflect.ValueOf(toFloat64(s))
}

func toFloat64PtrValue(s string) reflect.Value {
	return reflect.ValueOf(aws.Float64(toFloat64(s)))
}
