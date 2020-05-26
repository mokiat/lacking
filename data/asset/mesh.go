package asset

import "io"

const UnspecifiedOffset = int16(-1)

type Mesh struct {
	VertexData     []byte
	VertexStride   int16
	CoordOffset    int16
	NormalOffset   int16
	TangentOffset  int16
	TexCoordOffset int16
	ColorOffset    int16
	IndexData      []byte
	SubMeshes      []SubMesh
}

type SubMesh struct {
	Name             string
	Primitive        Primitive
	IndexOffset      uint32
	IndexCount       uint32
	MaterialType     string
	Color            [4]float32
	ColorTexture     string
	Roughness        float32
	RoughnessTexture string
	NormalScale      float32
	NormalTexture    string
}

type Primitive uint8

const (
	PrimitivePoints Primitive = iota
	PrimitiveLines
	PrimitiveLineStrip
	PrimitiveLineLoop
	PrimitiveTriangles
	PrimitiveTriangleStrip
	PrimitiveTriangleFan
)

func EncodeMesh(out io.Writer, mesh *Mesh) error {
	return Encode(out, mesh)
}

func DecodeMesh(in io.Reader, mesh *Mesh) error {
	return Decode(in, mesh)
}
