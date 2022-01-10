package asset

import "io"

type Model struct {
	Meshes []Mesh
	Nodes  []Node
}

func (m *Model) EncodeTo(out io.Writer) error {
	return Encode(out, m)
}

func (m *Model) DecodeFrom(in io.Reader) error {
	return Decode(in, m)
}

type Node struct {
	ParentIndex int16
	Name        string
	Matrix      [16]float32
	MeshIndex   int16
}

func EncodeModel(out io.Writer, model *Model) error {
	return Encode(out, model)
}

func DecodeModel(in io.Reader, model *Model) error {
	return Decode(in, model)
}
