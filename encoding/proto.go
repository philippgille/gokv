package encoding

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

// PBcodec encodes/decodes Go values to/from protocol buffers.
type PBcodec struct{}

// Marshal encodes a proto message struct into the binary wire format.
// Passed value can't be any Go value, but must be an object of a proto message struct.
func (c PBcodec) Marshal(v interface{}) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, errors.New("error casting interface to proto")
	}
	return proto.Marshal(msg)
}

// Unmarshal parses a wire-format message in proto message struct.
// Passed value can't be any Go value, but must be an object of a proto message struct.
func (c PBcodec) Unmarshal(data []byte, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return errors.New("error casting interface to proto")
	}
	return proto.Unmarshal(data, msg)
}
