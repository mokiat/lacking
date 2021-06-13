package graphics

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/framework/opengl/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics"
)

func NewEngine() *Engine {
	return &Engine{
		renderer: newRenderer(),
	}
}

var _ graphics.Engine = (*Engine)(nil)

type Engine struct {
	renderer *Renderer
}

func (e *Engine) Create() {
	e.renderer.Allocate()
}

func (e *Engine) CreateScene() graphics.Scene {
	return newScene(e.renderer)
}

func (e *Engine) CreateTwoDTexture(definition graphics.TwoDTextureDefinition) graphics.TwoDTexture {
	allocateInfo := opengl.TwoDTextureAllocateInfo{
		Width:             int32(definition.Width),
		Height:            int32(definition.Height),
		WrapS:             e.convertWrap(definition.WrapS),
		WrapT:             e.convertWrap(definition.WrapT),
		MinFilter:         e.convertMinFilter(definition.MinFilter),
		MagFilter:         e.convertMagFilter(definition.MagFilter),
		UseAnisotropy:     definition.UseAnisotropy,
		GenerateMipmaps:   definition.GenerateMipmaps,
		DataFormat:        e.convertDataFormat(definition.DataFormat),
		DataComponentType: e.convertDataComponentType(definition.DataFormat),
		InternalFormat:    e.convertInternalFormat(definition.InternalFormat),
		Data:              definition.Data,
	}
	result := newTwoDTexture()
	result.TwoDTexture.Allocate(allocateInfo)
	return result
}

func (e *Engine) CreateCubeTexture(definition graphics.CubeTextureDefinition) graphics.CubeTexture {
	allocateInfo := opengl.CubeTextureAllocateInfo{
		Dimension:         int32(definition.Dimension),
		WrapS:             e.convertWrap(definition.WrapS),
		WrapT:             e.convertWrap(definition.WrapT),
		MinFilter:         e.convertMinFilter(definition.MinFilter),
		MagFilter:         e.convertMagFilter(definition.MagFilter),
		DataFormat:        e.convertDataFormat(definition.DataFormat),
		DataComponentType: e.convertDataComponentType(definition.DataFormat),
		InternalFormat:    e.convertInternalFormat(definition.InternalFormat),
		FrontSideData:     definition.FrontSideData,
		BackSideData:      definition.BackSideData,
		LeftSideData:      definition.LeftSideData,
		RightSideData:     definition.RightSideData,
		TopSideData:       definition.TopSideData,
		BottomSideData:    definition.BottomSideData,
	}
	result := newCubeTexture()
	result.CubeTexture.Allocate(allocateInfo)
	return result
}

func (e *Engine) CreateMeshTemplate(definition graphics.MeshTemplateDefinition) graphics.MeshTemplate {
	vertexBuffer := opengl.NewBuffer()
	vertexBuffer.Allocate(opengl.BufferAllocateInfo{
		Dynamic: false,
		Data:    definition.VertexData,
	})

	indexBuffer := opengl.NewBuffer()
	indexBuffer.Allocate(opengl.BufferAllocateInfo{
		Dynamic: false,
		Data:    definition.IndexData,
	})

	var attributes []opengl.VertexArrayAttribute
	if definition.VertexFormat.HasCoord {
		attributes = append(attributes, opengl.VertexArrayAttribute{
			Index:          coordAttributeIndex,
			ComponentCount: 3,
			ComponentType:  gl.FLOAT,
			Normalized:     false,
			OffsetBytes:    uint32(definition.VertexFormat.CoordOffsetBytes),
			BufferBinding:  0,
		})
	}
	if definition.VertexFormat.HasNormal {
		attributes = append(attributes, opengl.VertexArrayAttribute{
			Index:          normalAttributeIndex,
			ComponentCount: 3,
			ComponentType:  gl.FLOAT,
			Normalized:     false,
			OffsetBytes:    uint32(definition.VertexFormat.NormalOffsetBytes),
			BufferBinding:  0,
		})
	}
	if definition.VertexFormat.HasTangent {
		attributes = append(attributes, opengl.VertexArrayAttribute{
			Index:          tangentAttributeIndex,
			ComponentCount: 3,
			ComponentType:  gl.FLOAT,
			Normalized:     false,
			OffsetBytes:    uint32(definition.VertexFormat.TangentOffsetBytes),
			BufferBinding:  0,
		})
	}
	if definition.VertexFormat.HasTexCoord {
		attributes = append(attributes, opengl.VertexArrayAttribute{
			Index:          texCoordAttributeIndex,
			ComponentCount: 2,
			ComponentType:  gl.FLOAT,
			Normalized:     false,
			OffsetBytes:    uint32(definition.VertexFormat.TexCoordOffsetBytes),
			BufferBinding:  0,
		})
	}
	if definition.VertexFormat.HasColor {
		attributes = append(attributes, opengl.VertexArrayAttribute{
			Index:          colorAttributeIndex,
			ComponentCount: 4,
			ComponentType:  gl.FLOAT,
			Normalized:     false,
			OffsetBytes:    uint32(definition.VertexFormat.ColorOffsetBytes),
			BufferBinding:  0,
		})
	}

	vertexArray := opengl.NewVertexArray()
	vertexArray.Allocate(opengl.VertexArrayAllocateInfo{
		BufferBindings: []opengl.VertexArrayBufferBinding{
			{
				VertexBuffer: vertexBuffer,
				OffsetBytes:  0,
				StrideBytes:  int32(definition.VertexFormat.CoordStrideBytes), // FIXME: Not accurate
			},
		},
		Attributes:  attributes,
		IndexBuffer: indexBuffer,
	})

	result := &MeshTemplate{
		vertexBuffer: vertexBuffer,
		indexBuffer:  indexBuffer,
		vertexArray:  vertexArray,
		subMeshes:    make([]SubMeshTemplate, len(definition.SubMeshes)),
	}
	for i, subMesh := range definition.SubMeshes {
		result.subMeshes[i] = SubMeshTemplate{
			material:         subMesh.Material.(*Material),
			primitive:        e.convertPrimitive(subMesh.Primitive),
			indexCount:       int32(subMesh.IndexCount),
			indexOffsetBytes: subMesh.IndexOffset,
		}
	}
	return result
}

func (e *Engine) CreatePBRMaterial(definition graphics.PBRMaterialDefinition) graphics.Material {
	extractTwoDTexture := func(src graphics.TwoDTexture) *opengl.TwoDTexture {
		if src == nil {
			return nil
		}
		return src.(*TwoDTexture).TwoDTexture
	}
	return &Material{
		backfaceCulling: definition.BackfaceCulling,
		alphaBlending:   definition.AlphaBlending,
		alphaTesting:    definition.AlphaTesting,
		alphaThreshold:  definition.AlphaThreshold,
		twoDTextures: []*opengl.TwoDTexture{
			extractTwoDTexture(definition.AlbedoTexture),
			extractTwoDTexture(definition.NormalTexture),
			extractTwoDTexture(definition.MetalnessTexture),
			extractTwoDTexture(definition.RoughnessTexture),
		},
		cubeTextures: []*opengl.CubeTexture{},
		vectors: []sprec.Vec4{
			definition.AlbedoColor,
			sprec.NewVec4(definition.NormalScale, definition.Metalness, definition.Roughness, 0.0),
		},
		geometryPresentation: internal.NewPBRGeometryPresentation(definition),
		shadowPresentation:   nil, // TODO
	}
}

func (e *Engine) Destroy() {
	e.renderer.Release()
}

func (e *Engine) convertWrap(wrap graphics.Wrap) int32 {
	switch wrap {
	case graphics.WrapClampToEdge:
		return gl.CLAMP_TO_EDGE
	case graphics.WrapRepeat:
		return gl.REPEAT
	default:
		panic(fmt.Errorf("unknown wrap mode: %d", wrap))
	}
}

func (e *Engine) convertMinFilter(filter graphics.Filter) int32 {
	switch filter {
	case graphics.FilterNearest:
		return gl.NEAREST
	case graphics.FilterLinear:
		return gl.LINEAR
	case graphics.FilterNearestMipmapNearest:
		return gl.NEAREST_MIPMAP_NEAREST
	case graphics.FilterNearestMipmapLinear:
		return gl.NEAREST_MIPMAP_LINEAR
	case graphics.FilterLinearMipmapNearest:
		return gl.LINEAR_MIPMAP_NEAREST
	case graphics.FilterLinearMipmapLinear:
		return gl.LINEAR_MIPMAP_LINEAR
	default:
		panic(fmt.Errorf("unknown min filter mode: %d", filter))
	}
}

func (e *Engine) convertMagFilter(filter graphics.Filter) int32 {
	switch filter {
	case graphics.FilterNearest:
		return gl.NEAREST
	case graphics.FilterLinear:
		return gl.LINEAR
	default:
		panic(fmt.Errorf("unknown mag filter mode: %d", filter))
	}
}

func (e *Engine) convertDataFormat(format graphics.DataFormat) uint32 {
	switch format {
	case graphics.DataFormatRGBA8:
		return gl.RGBA
	case graphics.DataFormatRGBA32F:
		return gl.RGBA
	default:
		panic(fmt.Errorf("unknown data format: %d", format))
	}
}

func (e *Engine) convertDataComponentType(format graphics.DataFormat) uint32 {
	switch format {
	case graphics.DataFormatRGBA8:
		return gl.UNSIGNED_BYTE
	case graphics.DataFormatRGBA32F:
		return gl.FLOAT
	default:
		panic(fmt.Errorf("unknown data format: %d", format))
	}
}

func (e *Engine) convertInternalFormat(format graphics.InternalFormat) uint32 {
	switch format {
	case graphics.InternalFormatRGBA8:
		return gl.SRGB8_ALPHA8
	case graphics.InternalFormatRGBA32F:
		return gl.RGBA32F
	default:
		panic(fmt.Errorf("unknown internal format: %d", format))
	}
}

func (e *Engine) convertPrimitive(primitive graphics.Primitive) uint32 {
	switch primitive {
	case graphics.PrimitivePoints:
		return gl.POINTS
	case graphics.PrimitiveLines:
		return gl.LINES
	case graphics.PrimitiveLineStrip:
		return gl.LINE_STRIP
	case graphics.PrimitiveLineLoop:
		return gl.LINE_LOOP
	case graphics.PrimitiveTriangles:
		return gl.TRIANGLES
	case graphics.PrimitiveTriangleStrip:
		return gl.TRIANGLE_STRIP
	case graphics.PrimitiveTriangleFan:
		return gl.TRIANGLE_FAN
	default:
		panic(fmt.Errorf("unknown primitive: %d", primitive))
	}
}
