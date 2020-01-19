package encoding

import (
	"encoding/json"
)

// Marshal converts a value into stringified-JSON
func Marshal(v interface{}) ([]byte, error) {
	one, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var two map[string]interface{}
	err = json.Unmarshal(one, &two)
	if err != nil {
		return nil, err
	}

	three, err := Stringify(two)
	if err != nil {
		return nil, err
	}

	four, err := json.Marshal(three)
	if err != nil {
		return nil, err
	}

	return four, nil
}
