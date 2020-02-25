package gokv

import "time"

// SetterWithExp is an abstraction for setter with expiration feature
type SetterWithExp interface {
	// SetExp works like Set, but supports key expiration
	SetExp(k string, v interface{}, exp time.Duration) error
}
