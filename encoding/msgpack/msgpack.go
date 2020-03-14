package msgpack

import "github.com/vmihailenco/msgpack/v4"

// Codec encodes/decodes Go values to/from Msgpack.
// You can use MsgPack instead of creating an instance of this struct.
type Codec struct{}

// Marshal encodes a Go value to Msgpack.
func (c Codec) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Unmarshal decodes a Msgpack value into a Go value.
func (c Codec) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

var MsgPack = Codec{}
