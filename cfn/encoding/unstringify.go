package encoding

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func convertStruct(i interface{}, t reflect.Type, pointer bool) (reflect.Value, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return zeroValue, fmt.Errorf("Cannot convert %T to struct", i)
	}

	out := reflect.New(t)

	Unstringify(m, out.Interface())

	if !pointer {
		out = out.Elem()
	}

	return out, nil
}

func convertString(i interface{}, pointer bool) (reflect.Value, error) {
	s, ok := i.(string)

	if !ok {
		return zeroValue, fmt.Errorf("Cannot convert %T to string", i)
	}

	if pointer {
		return reflect.ValueOf(&s), nil
	}

	return reflect.ValueOf(s), nil
}

func convertBool(i interface{}, pointer bool) (reflect.Value, error) {
	var b bool
	var err error

	switch v := i.(type) {
	case bool:
		b = v

	case string:
		b, err = strconv.ParseBool(v)
		if err != nil {
			return zeroValue, err
		}

	default:
		return zeroValue, fmt.Errorf("Cannot convert %T to bool", i)
	}

	if pointer {
		return reflect.ValueOf(&b), nil
	}

	return reflect.ValueOf(b), nil
}

func convertInt(i interface{}, pointer bool) (reflect.Value, error) {
	var n int

	switch v := i.(type) {
	case int:
		n = v

	case float64:
		n = int(v)

	case string:
		n64, err := strconv.ParseInt(v, 0, 32)
		if err != nil {
			return zeroValue, err
		}

		n = int(n64)

	default:
		return zeroValue, fmt.Errorf("Cannot convert %T to bool", i)
	}

	if pointer {
		return reflect.ValueOf(&n), nil
	}

	return reflect.ValueOf(n), nil
}

func convertFloat64(i interface{}, pointer bool) (reflect.Value, error) {
	var f float64
	var err error

	switch v := i.(type) {
	case float64:
		f = v

	case int:
		f = float64(v)

	case string:
		f, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return zeroValue, err
		}

	default:
		return zeroValue, fmt.Errorf("Cannot convert %T to bool", i)
	}

	if pointer {
		return reflect.ValueOf(&f), nil
	}

	return reflect.ValueOf(f), nil
}

func convertType(t reflect.Type, i interface{}) (reflect.Value, error) {
	pointer := false
	if t.Kind() == reflect.Ptr {
		pointer = true
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		return convertStruct(i, t, pointer)

	case reflect.String:
		return convertString(i, pointer)

	case reflect.Bool:
		return convertBool(i, pointer)

	case reflect.Int:
		return convertInt(i, pointer)

	case reflect.Float64:
		return convertFloat64(i, pointer)

	default:
		return zeroValue, fmt.Errorf("Cannot convert %T into %v", i, t)
	}
}

// Unstringify takes a stringified representation of a value
// and populates it into the supplied interface
func Unstringify(data map[string]interface{}, v interface{}) error {
	t := reflect.TypeOf(v).Elem()

	val := reflect.ValueOf(v).Elem()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		jsonName := f.Name
		jsonTag := strings.Split(f.Tag.Get("json"), ",")
		if len(jsonTag) > 0 && jsonTag[0] != "" {
			jsonName = jsonTag[0]
		}

		if value, ok := data[jsonName]; ok {
			newValue, err := convertType(f.Type, value)
			if err != nil {
				return err
			}

			val.FieldByName(f.Name).Set(newValue)
		}
	}

	return nil
}
