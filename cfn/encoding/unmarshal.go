package encoding

import (
	"encoding/json"
)

// Unmarshal converts stringified-JSON into the passed-in type
func Unmarshal(data []byte, v interface{}) error {
	var dataMap map[string]interface{}
	err := json.Unmarshal(data, &dataMap)
	if err != nil {
		return err
	}

	Unstringify(dataMap, v)
	if err != nil {
		return err
	}

	return nil
}
