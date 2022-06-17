package asset

import (
	"encoding/gob"
	"io"
)

func (m *Model) encodeV1(out io.Writer) error {
	return gob.NewEncoder(out).Encode(m)
}

func (m *Model) decodeV1(in io.Reader) error {
	return gob.NewDecoder(in).Decode(m)
}
