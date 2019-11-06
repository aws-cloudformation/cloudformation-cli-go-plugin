package encoding

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type String struct {
	value *string
}

func NewString(s string) String {
	return String{&s}
}

func (s String) Value() *string {
	return s.value
}

func (s String) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.value)
}

func (s *String) UnmarshalJSON(data []byte) error {
	var ss *string
	err := json.Unmarshal(data, &ss)
	if err != nil {
		return err
	}

	s.value = ss
	return nil
}

type Bool struct {
	value *bool
}

func NewBool(b bool) Bool {
	return Bool{&b}
}

func (b Bool) Value() *bool {
	return b.value
}

func (b Bool) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprint(b.value))
}

func (b *Bool) UnmarshalJSON(data []byte) error {
	var s *string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	if s == nil {
		b.value = nil
	} else {
		val, err := strconv.ParseBool(*s)
		if err != nil {
			return err
		}
		b.value = &val
	}

	return nil
}

type Int struct {
	value *int64
}

func NewInt(i int64) Int {
	return Int{&i}
}

func (i Int) Value() *int64 {
	return i.value
}

func (i Int) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprint(i.value))
}

func (i *Int) UnmarshalJSON(data []byte) error {
	var s *string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	if s == nil {
		i.value = nil
	} else {
		val, err := strconv.ParseInt(*s, 0, 64)
		if err != nil {
			return err
		}
		i.value = &val
	}

	return nil
}

type Float struct {
	value *float64
}

func NewFloat(f float64) Float {
	return Float{&f}
}

func (f Float) Value() *float64 {
	return f.value
}

func (f Float) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprint(f.value))
}

func (f *Float) UnmarshalJSON(data []byte) error {
	var s *string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	if s == nil {
		f.value = nil
	} else {
		val, err := strconv.ParseFloat(*s, 64)
		if err != nil {
			return err
		}
		f.value = &val
	}

	return nil
}
