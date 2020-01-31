package encoding

import (
	"encoding/json"
)

// Marshal converts a value into stringified-JSON
func Marshal(v interface{}) ([]byte, error) {
	stringified, err := Stringify(v)
	if err != nil {
		return nil, err
	}

	return json.Marshal(stringified)
}
