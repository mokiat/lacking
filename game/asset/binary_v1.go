package asset

import (
	"io"

	"github.com/mokiat/lacking/util/blob"
)

func (b *Binary) encodeV1(out io.Writer) error {
	writer := blob.NewTypedWriter(out)
	if err := writer.WriteByteBlock(b.Data); err != nil {
		return err
	}
	return nil
}

func (b *Binary) decodeV1(in io.Reader) error {
	reader := blob.NewTypedReader(in)
	if data, err := reader.ReadBytesBlock(); err != nil {
		return err
	} else {
		b.Data = data
	}
	return nil
}
