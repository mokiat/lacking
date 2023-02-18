package asset

import (
	"io"

	"github.com/mokiat/gblob"
)

func (s *Scene) encodeV1(out io.Writer) error {
	return gblob.NewLittleEndianPackedEncoder(out).Encode(s)
}

func (s *Scene) decodeV1(in io.Reader) error {
	return gblob.NewLittleEndianPackedDecoder(in).Decode(s)
}
