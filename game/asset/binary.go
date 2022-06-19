package asset

import (
	"fmt"
	"io"
)

// Binary represents a blob of bytes
type Binary struct {
	Data []byte
}

func (b *Binary) EncodeTo(out io.Writer) error {
	return encodeResource(out, header{
		Version: 1,
		Flags:   headerFlagZlib,
	}, b)
}

func (b *Binary) DecodeFrom(in io.Reader) error {
	return decodeResource(in, b)
}

func (b *Binary) encodeVersionTo(out io.Writer, version uint16) error {
	switch version {
	case 1:
		return b.encodeV1(out)
	default:
		panic(fmt.Errorf("unknown version %d", version))
	}
}

func (b *Binary) decodeVersionFrom(in io.Reader, version uint16) error {
	switch version {
	case 1:
		return b.decodeV1(in)
	default:
		panic(fmt.Errorf("unknown version %d", version))
	}
}
