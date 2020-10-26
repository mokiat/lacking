package asset

import "io"

const UnspecifiedOffset = int16(-1)

type Mesh struct {
	Name         string
	VertexData   []byte
	VertexLayout VertexLayout
	IndexData    []byte
	SubMeshes    []SubMesh
}

type VertexLayout struct {
	CoordOffset    int16
	CoordStride    int16
	NormalOffset   int16
	NormalStride   int16
	TangentOffset  int16
	TangentStride  int16
	TexCoordOffset int16
	TexCoordStride int16
	ColorOffset    int16
	ColorStride    int16
}

type SubMesh struct {
	Primitive   Primitive
	IndexOffset uint32
	IndexCount  uint32
	Material    Material
}

type Material struct {
	Type             string
	BackfaceCulling  bool
	AlphaTesting     bool
	AlphaThreshold   float32
	Metalness        float32
	MetalnessTexture string
	Roughness        float32
	RoughnessTexture string
	Color            [4]float32
	ColorTexture     string
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
