package asset

import (
	"io"

	"github.com/mokiat/gblob"
)

func (t *CubeTexture) encodeV1(out io.Writer) error {
	return gblob.NewLittleEndianPackedEncoder(out).Encode(t)
}

func (t *CubeTexture) decodeV1(in io.Reader) error {
	return gblob.NewLittleEndianPackedDecoder(in).Decode(t)
}
