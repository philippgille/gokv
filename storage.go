package gokv

import (
	"encoding/json"
)

func toJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func fromJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
