package util

import "encoding/json"

// ToJSON marshals the given value into JSON.
func ToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// FromJSON unmarshals the given JSON and populates the value that the given pointer points to accordingly.
func FromJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
