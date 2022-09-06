package asset

import (
	"encoding/gob"
	"io"
)

func (s *Scene) encodeV1(out io.Writer) error {
	return gob.NewEncoder(out).Encode(s)
}

func (s *Scene) decodeV1(in io.Reader) error {
	return gob.NewDecoder(in).Decode(s)
}
