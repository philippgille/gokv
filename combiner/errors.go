package combiner

import (
	"strings"
)

// MultiError contains a list of errors.
type MultiError struct {
	Errors []error
}

func (e MultiError) Error() string {
	if e.Errors == nil {
		return "<nil>"
	}

	s := make([]string, len(e.Errors))

	for i := range e.Errors {
		s[i] = e.Errors[i].Error()
	}

	return strings.Join(s, " | ")
}

func (e *MultiError) addError(err error) {
	e.Errors = append(e.Errors, err)
}

func newMultiError(err ...error) MultiError {
	return MultiError{err}
}
