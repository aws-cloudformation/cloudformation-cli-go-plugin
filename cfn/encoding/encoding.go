package encoding

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type String string

func NewString(s string) *String {
	out := String(s)
	return &out
}

func (s String) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *String) UnmarshalJSON(data []byte) error {
	var ss string
	err := json.Unmarshal(data, &ss)

	if err == nil {
		*s = String(ss)
	}
	return err
}

type Bool bool

func NewBool(b bool) *Bool {
	out := Bool(b)
	return &out
}

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
	if err == nil {
		*b = Bool(val)
	}
	return err
}

type Float float64

func NewFloat(f float64) *Float {
	out := Float(f)
	return &out
}

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
	if err == nil {
		*f = Float(num)
	}
	return err
}

type Int int64

func NewInt(i int64) *Int {
	out := Int(i)
	return &out
}

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
	if err == nil {
		*i = Int(num)
	}
	return err
}
