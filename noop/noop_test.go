package noop_test

import (
	"testing"

	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/noop"
)

func TestNop(t *testing.T) {
	t.Parallel()

	var s gokv.Store = noop.NewStore()

	if err := s.Set("foo", 1); err != nil {
		t.Error(err)
	}

	var v int
	found, err := s.Get("foo", &v)
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("A value was found, but no value was expected")
	}

	if err := s.Delete("foo"); err != nil {
		t.Error(err)
	}

	if err := s.Close(); err != nil {
		t.Error(err)
	}
}

func TestInputValidation(t *testing.T) {
	t.Parallel()

	var s gokv.Store = noop.NewStore()

	{
		err := s.Set("", 1)
		assertEqualError(t, err, errInvalidKey)
	}

	{
		err := s.Set("foo", nil)
		assertEqualError(t, err, errInvalidValue)
	}

	{
		var v int
		found, err := s.Get("", &v)
		assertEqualError(t, err, errInvalidKey)
		if found {
			t.Error("A value was found, but no value was expected")
		}
	}

	{
		found, err := s.Get("foo", nil)
		assertEqualError(t, err, errInvalidValue)
		if found {
			t.Error("A value was found, but no value was expected")
		}
	}

	{
		err := s.Delete("")
		assertEqualError(t, err, errInvalidKey)
	}
}

func assertEqualError(t *testing.T, err error, expectedErrMsg string) {
	t.Helper()

	if err == nil {
		t.Error("expect error, got nil")
	} else if err.Error() != expectedErrMsg {
		t.Error(err)
	}
}

// TODO: We updated error capitalization in the utils package after v0.7.0, but
// as long as utils doesn't have a v0.8.0 release, we can't import it yet and have
// to match on the old error messages.
var (
	errInvalidKey   = "The passed key is an empty string, which is invalid"
	errInvalidValue = "The passed value is nil, which is not allowed"
)
