package asset

import "io"

type Mesh struct {
	VertexData     []byte
	VertexStride   uint8
	CoordOffset    uint8
	NormalOffset   uint8
	TexCoordOffset uint8
	IndexData      []byte
	SubMeshes      []SubMesh
}

type SubMesh struct {
	Name           string
	IndexOffset    uint32
	IndexCount     uint32
	DiffuseTexture string
}

func EncodeMesh(out io.Writer, mesh *Mesh) error {
	return Encode(out, mesh)
}

func DecodeMesh(in io.Reader, mesh *Mesh) error {
	return Decode(in, mesh)
}
