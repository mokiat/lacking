package graphics

import "github.com/mokiat/gomath/sprec"

type RenderPrimitive int

const (
	RenderPrimitiveTriangles RenderPrimitive = iota
	RenderPrimitiveLines
)

func createItem() Item {
	return Item{}
}

type Item struct {
	Program   *Program
	Primitive RenderPrimitive

	// TODO: Make uniforms generic through usage of
	// uniform type specifiers and []byte buffers
	ModelMatrix       sprec.Mat4
	Metalness         float32
	Roughness         float32
	AlbedoColor       sprec.Vec4
	AlbedoTwoDTexture *TwoDTexture
	AlbedoCubeTexture *CubeTexture

	VertexArray *VertexArray
	IndexCount  int32
}

func (i *Item) reset() {
	i.Program = nil
	i.Primitive = RenderPrimitiveTriangles
	i.Metalness = 0.0
	i.Roughness = 0.5
	i.AlbedoColor = sprec.NewVec4(0.5, 0.0, 0.5, 1.0)
	i.AlbedoTwoDTexture = nil
	i.AlbedoCubeTexture = nil
	i.VertexArray = nil
	i.IndexCount = 0
}
