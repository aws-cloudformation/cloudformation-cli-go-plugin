package encoding

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type String string

func (s String) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *String) UnmarshalJSON(data []byte) error {
	var ss string
	err := json.Unmarshal(data, &ss)
	*s = String(ss)

	return err
}

type Bool bool

func (b Bool) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprint(b))
}

func (b *Bool) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	val, err := strconv.ParseBool(s)
	*b = Bool(val)

	return err
}

type Float float64

func (f Float) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprint(f))
}

func (f *Float) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	num, err := strconv.ParseFloat(s, 64)
	*f = Float(num)

	return err
}

type Int int64

func (i Int) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprint(i))
}

func (i *Int) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	num, err := strconv.ParseInt(s, 0, 64)
	*i = Int(num)

	return err
}
