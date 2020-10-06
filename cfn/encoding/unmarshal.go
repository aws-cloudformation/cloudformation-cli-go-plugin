package encoding

import (
	"encoding/json"
)

// Unmarshal converts stringified-JSON into the passed-in type
func Unmarshal(data []byte, v interface{}) error {
	var dataMap map[string]interface{}
	var err error
	err = json.Unmarshal(data, &dataMap)
	if err != nil {
		return err
	}

	err = Unstringify(dataMap, v)
	if err != nil {
		return err
	}

	return nil
}
