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
		out := make(map[string]interface{})
		for _, key := range val.MapKeys() {
			v, err := Stringify(val.MapIndex(key).Interface())
			switch {
			case err != nil:
				return nil, err
			case v != nil:
				out[key.String()] = v
			}
		}
		return out, nil
	case reflect.Slice:
		out := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			v, err = Stringify(val.Index(i).Interface())
			switch {
			case err != nil:
				return nil, err
			case v != nil:
				out[i] = v
			}
		}
		return out, nil
	case reflect.Struct:
		t := val.Type()
		out := make(map[string]interface{})
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			v, err := Stringify(val.FieldByName(f.Name).Interface())
			switch {
			case err != nil:
				return nil, err
			case v != nil:
				out[f.Name] = v
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
