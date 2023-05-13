package asset

import (
	"io"

	"github.com/mokiat/gblob"
)

func (b *Binary) encodeV1(out io.Writer) error {
	return gblob.NewLittleEndianPackedEncoder(out).Encode(b)
}

func (b *Binary) decodeV1(in io.Reader) error {
	return gblob.NewLittleEndianPackedDecoder(in).Decode(b)
}
