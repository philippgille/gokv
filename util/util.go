package util

import (
	"errors"
	"time"
)

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

const MinExpiration = 10 * time.Second

// CheckExp returns false if expiration time is too small
func CheckExp(exp time.Duration) bool {
	if exp < MinExpiration {
		return true
	}
	return false
}
