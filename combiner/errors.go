package combiner

import (
	"errors"
	"strings"
)

type Error struct {
	Errs []error
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
	if e.Errs == nil {
		return "<nil>"
	}

	s := make([]string, len(e.Errs))

	for i := range e.Errs {
		s[i] = e.Errs[i].Error()
	}

	return strings.Join(s, " | ")
}
