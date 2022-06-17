package asset

import (
	"fmt"
	"io"
)

type Model struct {
	Nodes           []Node
	Materials       []Material
	MeshDefinitions []MeshDefinition
	MeshInstances   []MeshInstance
	BodyDefinitions []BodyDefinition
	BodyInstances   []BodyInstance
	// TODO: Model Instances (ref model resources)
	// TODO: Speakers
	// TODO: Lights
	// TODO: Cameras
	// TODO: Constraints
}

func (m *Model) EncodeTo(out io.Writer) error {
	return encodeResource(out, header{
		Version: 1,
		Flags:   headerFlagZlib,
	}, m)
}

func (m *Model) DecodeFrom(in io.Reader) error {
	return decodeResource(in, m)
}

func (m *Model) encodeVersionTo(out io.Writer, version uint16) error {
	switch version {
	case 1:
		return m.encodeV1(out)
	default:
		panic(fmt.Errorf("unknown version %d", version))
	}
}

func (m *Model) decodeVersionFrom(in io.Reader, version uint16) error {
	switch version {
	case 1:
		return m.decodeV1(in)
	default:
		panic(fmt.Errorf("unknown version %d", version))
	}
}