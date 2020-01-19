package encoding

import (
	"fmt"
	"reflect"
)

var zeroValue reflect.Value

// Stringify converts any supported type into a stringified value
func Stringify(v interface{}) (interface{}, error) {
	var err error

	if v == nil {
		return nil, nil
	}

	val := reflect.ValueOf(v)

	switch val.Kind() {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Float64:
		return fmt.Sprint(v), nil
	case reflect.Map:
		vMap := v.(map[string]interface{})
		out := make(map[string]interface{})
		for key, value := range vMap {
			out[key], err = Stringify(value)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	case reflect.Slice:
		vSlice := v.([]interface{})
		out := make([]interface{}, len(vSlice))
		for i, value := range vSlice {
			out[i], err = Stringify(value)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	case reflect.Struct:
		t := val.Type()
		out := make(map[string]interface{})

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			out[f.Name], err = Stringify(val.FieldByName(f.Name).Interface())
			if err != nil {
				return nil, err
			}
		}

		return out, nil
	case reflect.Ptr:
		if val.IsNil() {
			return nil, nil
		}

		return Stringify(val.Elem().Interface())
	}

	return nil, fmt.Errorf("Unsupported type: '%v'", val.Kind())
}
