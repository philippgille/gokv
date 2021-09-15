//nolint:paralleltest
package encoding_test

import (
	"testing"

	"github.com/philippgille/gokv/encoding"
	"github.com/stretchr/testify/require"
)

func TestTOMLImplements(t *testing.T) {
	require.Implements(t, (*encoding.Codec)(nil), new(encoding.TOMLcodec))
}

func TestTOMLMarshal(t *testing.T) {
	require := require.New(t)

	tables := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			"string string map",
			map[string]string{"foo": "bar"},
			"foo = \"bar\"\n",
		},
		{
			"struct",
			struct {
				Foo string
				Bar int
			}{
				"foo",
				7,
			},
			"Foo = \"foo\"\nBar = 7\n",
		},
		{
			"sub struct",
			struct {
				Foo struct{ Bar string }
			}{
				struct{ Bar string }{"bar"},
			},
			"[Foo]\n  Bar = \"bar\"\n",
		},
	}

	for _, table := range tables {
		table := table
		t.Run(table.name, func(t *testing.T) {
			r, err := encoding.TOML.Marshal(table.input)
			require.NoError(err)

			require.Equal(table.expected, string(r))
		})
	}
}

func TestTOMLUnmarshal(t *testing.T) {
	require := require.New(t)

	tables := []struct {
		name     string
		input    []byte
		expected interface{}
	}{
		{
			"to struct",
			[]byte("Foo = \"foo\"\nBar = 7\n"),
			struct {
				Foo string
				Bar int
			}{
				"foo",
				7,
			},
		},
	}

	for _, table := range tables {
		table := table
		t.Run(table.name, func(t *testing.T) {
			var s struct {
				Foo string
				Bar int
			}

			err := encoding.TOML.Unmarshal(table.input, &s)
			require.NoError(err)

			require.Equal(table.expected, s)
		})
	}
}
