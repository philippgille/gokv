package util

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

// ToJSON marshals the given value into JSON.
func ToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// FromJSON unmarshals the given JSON and populates the value that the given pointer points to accordingly.
func FromJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ToGob marshals the given value into a gob.
func ToGob(v interface{}) ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// FromGob unmarshals the given gob and populates the value that the given pointer points to accordingly.
func FromGob(data []byte, v interface{}) error {
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)
	return decoder.Decode(v)
}
