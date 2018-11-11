package util

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
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

// CheckKeyAndValue returns an error if k == "" or if v == nil
func CheckKeyAndValue(k string, v interface{}) error {
	if err := CheckKey(k); err != nil {
		return err
	}
	return CheckVal(v)
}

// CheckKey returns an error if k == ""
func CheckKey(k string) error {
	if k == "" {
		return errors.New("The passed key is an empty string, which is invalid")
	}
	return nil
}

// CheckVal returns an error if v == nil
func CheckVal(v interface{}) error {
	if v == nil {
		return errors.New("The passed value is nil, which is not allowed")
	}
	return nil
}
