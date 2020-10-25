package asset

import "io"

type Model struct {
	Meshes []Mesh
	Nodes  []Node
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
