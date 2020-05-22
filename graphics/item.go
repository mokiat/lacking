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
	Primitive      RenderPrimitive
	Program        *Program
	ModelMatrix    sprec.Mat4
	SkyboxTexture  *CubeTexture
	DiffuseColor   sprec.Vec4
	DiffuseTexture *TwoDTexture
	VertexArray    *VertexArray
	IndexCount     int32
}

func (i *Item) reset() {
	i.Primitive = RenderPrimitiveTriangles
	i.Program = nil
	i.SkyboxTexture = nil
	i.DiffuseColor = sprec.NewVec4(0.0, 0.0, 0.0, 1.0)
	i.DiffuseTexture = nil
	i.VertexArray = nil
	i.IndexCount = 0
}
