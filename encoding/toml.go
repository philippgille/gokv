package encoding

import (
	"bytes"

	"github.com/BurntSushi/toml"
)

// TOMLcodec encodes/decodes Go values to/from TOML.
// You can use encoding.TOML instead of creating an instance of this struct.
type TOMLcodec struct{}

// Marshal encodes a Go value to TOML.
func (c TOMLcodec) Marshal(v interface{}) ([]byte, error) {
	// return toml.Marshal(v)
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(v); err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

// Unmarshal decodes a TOML value into a Go value.
func (c TOMLcodec) Unmarshal(data []byte, v interface{}) error {
	return toml.Unmarshal(data, v)
}
