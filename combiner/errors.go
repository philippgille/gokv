package combiner

import (
	"errors"
	"strings"
)

type Error struct {
	errs []error
}

var (
	ErrNotEnoughStores = newError("at least 2 stores needed")
)

func newError(err string) Error {
	return newErrors(errors.New("combiner: " + err))
}

func newErrors(err ...error) Error {
	return Error{err}
}

func (e Error) Error() string {
	if e.errs == nil {
		return "<nil>"
	}

	s := make([]string, len(e.errs))

	for i := range e.errs {
		s[i] = e.errs[i].Error()
	}

	return strings.Join(s, " | ")
}
