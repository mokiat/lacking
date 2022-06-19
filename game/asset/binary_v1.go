package asset

import (
	"io"

	"github.com/mokiat/lacking/data/storage"
)

func (b *Binary) encodeV1(out io.Writer) error {
	writer := storage.NewTypedWriter(out)
	if err := writer.WriteByteBlock(b.Data); err != nil {
		return err
	}
	return nil
}

func (b *Binary) decodeV1(in io.Reader) error {
	reader := storage.NewTypedReader(in)
	if data, err := reader.ReadBytesBlock(); err != nil {
		return err
	} else {
		b.Data = data
	}
	return nil
}
