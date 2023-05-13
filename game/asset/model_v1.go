package asset

import (
	"io"

	"github.com/mokiat/gblob"
)

func (m *Model) encodeV1(out io.Writer) error {
	return gblob.NewLittleEndianPackedEncoder(out).Encode(m)
}

func (m *Model) decodeV1(in io.Reader) error {
	return gblob.NewLittleEndianPackedDecoder(in).Decode(m)
}
