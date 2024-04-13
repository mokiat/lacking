package asset

import (
	"fmt"
	"io"
)

type Scene struct {
	ModelDefinitions []Model
	ModelInstances   []ModelInstance
}

func (s *Scene) EncodeTo(out io.Writer) error {
	return encodeResource(out, header{
		Version: 1,
		Flags:   headerFlagZlib,
	}, s)
}

func (s *Scene) DecodeFrom(in io.Reader) error {
	return decodeResource(in, s)
}

func (s *Scene) encodeVersionTo(out io.Writer, version uint16) error {
	switch version {
	case 1:
		return s.encodeV1(out)
	default:
		panic(fmt.Errorf("unknown version %d", version))
	}
}

func (s *Scene) decodeVersionFrom(in io.Reader, version uint16) error {
	switch version {
	case 1:
		return s.decodeV1(in)
	default:
		panic(fmt.Errorf("unknown version %d", version))
	}
}
