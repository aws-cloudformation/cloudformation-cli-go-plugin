package encoding

import (
	"fmt"
	"reflect"
	"strings"
)

var zeroValue reflect.Value

var stringType = reflect.TypeOf("")
var interfaceType = reflect.TypeOf((*interface{})(nil)).Elem()

func stringifyType(t reflect.Type) reflect.Type {
	switch t.Kind() {
	case reflect.Map:
		return reflect.MapOf(stringType, interfaceType)
	case reflect.Slice:
		return reflect.SliceOf(interfaceType)
	case reflect.Struct:
		return stringifyStructType(t)
	case reflect.Ptr:
		return stringifyType(t.Elem())
	default:
		return stringType
	}
}

func stringifyStructType(t reflect.Type) reflect.Type {
	fields := make([]reflect.StructField, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		fields[i] = reflect.StructField{
			Name: f.Name,
			Type: stringifyType(f.Type),
			Tag:  f.Tag,
		}
	}

	return reflect.StructOf(fields)
}

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

		out := reflect.New(stringifyStructType(t)).Elem()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			v := val.FieldByName(f.Name)

			if tag, ok := f.Tag.Lookup("json"); ok {
				if strings.Contains(tag, ",omitempty") {
					if v.IsZero() {
						continue
					}
				}
			}

			s, err := Stringify(v.Interface())
			switch {
			case err != nil:
				return nil, err
			case s != nil:
				out.Field(i).Set(reflect.ValueOf(s))
			}
		}

		return out.Interface(), nil
	case reflect.Ptr:
		if val.IsNil() {
			return nil, nil
		}

		return Stringify(val.Elem().Interface())
	}

	return nil, fmt.Errorf("Unsupported type: '%v'", val.Kind())
}
