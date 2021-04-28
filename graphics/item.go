package graphics

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/mokiat/gomath/sprec"
)

type RenderPrimitive int

const (
	RenderPrimitivePoints RenderPrimitive = iota
	RenderPrimitiveLines
	RenderPrimitiveLineStrip
	RenderPrimitiveLineLoop
	RenderPrimitiveTriangles
	RenderPrimitiveTriangleStrip
	RenderPrimitiveTriangleFan
)

func createItem() Item {
	return Item{}
}

type Item struct {
	Program         *Program
	Primitive       RenderPrimitive
	BackfaceCulling bool

	// TODO: Make uniforms generic through usage of
	// uniform type specifiers and []byte buffers
	ModelMatrix              sprec.Mat4
	Metalness                float32
	MetalnessTwoDTexture     *TwoDTexture
	Roughness                float32
	RoughnessTwoDTexture     *TwoDTexture
	AlbedoColor              sprec.Vec4
	AlbedoTwoDTexture        *TwoDTexture
	AlbedoCubeTexture        *CubeTexture
	AmbientReflectionTexture *CubeTexture
	AmbientRefractionTexture *CubeTexture
	NormalScale              float32
	NormalTwoDTexture        *TwoDTexture
	LightDirectionWS         sprec.Vec3
	Exposure                 float32

	VertexArray *VertexArray
	IndexOffset int
	IndexCount  int32
}

func (i *Item) reset() {
	i.Program = nil
	i.Primitive = RenderPrimitiveTriangles
	i.BackfaceCulling = false
	i.Metalness = 0.0
	i.MetalnessTwoDTexture = nil
	i.Roughness = 0.5
	i.RoughnessTwoDTexture = nil
	i.AlbedoColor = sprec.NewVec4(0.5, 0.0, 0.5, 1.0)
	i.AlbedoTwoDTexture = nil
	i.AlbedoCubeTexture = nil
	i.AmbientReflectionTexture = nil
	i.AmbientRefractionTexture = nil
	i.NormalScale = 1.0
	i.NormalTwoDTexture = nil
	i.LightDirectionWS = sprec.NewVec3(0.0, 1.0, 0.0)
	i.Exposure = 1.0
	i.VertexArray = nil
	i.IndexOffset = 0
	i.IndexCount = 0
}

func (i *Item) glPrimitive() uint32 {
	switch i.Primitive {
	case RenderPrimitivePoints:
		return gl.POINTS
	case RenderPrimitiveLines:
		return gl.LINES
	case RenderPrimitiveLineStrip:
		return gl.LINE_STRIP
	case RenderPrimitiveLineLoop:
		return gl.LINE_LOOP
	case RenderPrimitiveTriangles:
		return gl.TRIANGLES
	case RenderPrimitiveTriangleStrip:
		return gl.TRIANGLE_STRIP
	case RenderPrimitiveTriangleFan:
		return gl.TRIANGLE_FAN
	default:
		panic(fmt.Errorf("unsupported primitive type: %d", i.Primitive))
	}
}
