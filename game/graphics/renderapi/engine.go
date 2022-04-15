package graphics

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/renderapi/internal"
	"github.com/mokiat/lacking/game/graphics/renderapi/plugin"
	"github.com/mokiat/lacking/render"
)

func NewEngine(api render.API, shaders plugin.ShaderCollection) *Engine {
	return &Engine{
		api:      api,
		shaders:  shaders,
		renderer: newRenderer(api, shaders),
	}
}

var _ graphics.Engine = (*Engine)(nil)

type Engine struct {
	api      render.API
	shaders  plugin.ShaderCollection
	renderer *Renderer
}

func (e *Engine) Create() {
	e.renderer.Allocate()
}

func (e *Engine) CreateScene() graphics.Scene {
	return newScene(e.renderer)
}

func (e *Engine) CreateTwoDTexture(definition graphics.TwoDTextureDefinition) graphics.TwoDTexture {
	return newTwoDTexture(e.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           definition.Width,
		Height:          definition.Height,
		Wrapping:        e.convertWrap(definition.WrapS),                                 // TODO: Remove WrapS / WrapT capability from definition
		Filtering:       e.convertFilter(definition.MinFilter, definition.UseAnisotropy), // TODO: Remove Min / Mag capability from definition
		Mipmapping:      e.needsMipmaps(definition.MinFilter),
		GammaCorrection: true,
		Format:          e.convertFormat(definition.InternalFormat, definition.DataFormat),
		Data:            definition.Data,
	}))
}

func (e *Engine) CreateCubeTexture(definition graphics.CubeTextureDefinition) graphics.CubeTexture {
	return newCubeTexture(e.api.CreateColorTextureCube(render.ColorTextureCubeInfo{
		Dimension:       definition.Dimension,
		Filtering:       e.convertFilter(definition.MinFilter, false), // TODO: Remove Min / Mag capability from definition
		Mipmapping:      e.needsMipmaps(definition.MinFilter),
		GammaCorrection: true,
		Format:          e.convertFormat(definition.InternalFormat, definition.DataFormat),
		FrontSideData:   definition.FrontSideData,
		BackSideData:    definition.BackSideData,
		LeftSideData:    definition.LeftSideData,
		RightSideData:   definition.RightSideData,
		TopSideData:     definition.TopSideData,
		BottomSideData:  definition.BottomSideData,
	}))
}

func (e *Engine) CreateMeshTemplate(definition graphics.MeshTemplateDefinition) graphics.MeshTemplate {
	vertexBuffer := e.api.CreateVertexBuffer(render.BufferInfo{
		Dynamic: false,
		Data:    definition.VertexData,
	})
	indexBuffer := e.api.CreateIndexBuffer(render.BufferInfo{
		Dynamic: false,
		Data:    definition.IndexData,
	})

	var attributes []render.VertexArrayAttributeInfo
	if definition.VertexFormat.HasCoord {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: coordAttributeIndex,
			Format:   render.VertexAttributeFormatRGB32F,
			Offset:   definition.VertexFormat.CoordOffsetBytes,
		})
	}
	if definition.VertexFormat.HasNormal {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: normalAttributeIndex,
			Format:   render.VertexAttributeFormatRGB32F,
			Offset:   definition.VertexFormat.NormalOffsetBytes,
		})
	}
	if definition.VertexFormat.HasTangent {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: tangentAttributeIndex,
			Format:   render.VertexAttributeFormatRGB32F,
			Offset:   definition.VertexFormat.TangentOffsetBytes,
		})
	}
	if definition.VertexFormat.HasTexCoord {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: texCoordAttributeIndex,
			Format:   render.VertexAttributeFormatRG32F,
			Offset:   definition.VertexFormat.TexCoordOffsetBytes,
		})
	}
	if definition.VertexFormat.HasColor {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: colorAttributeIndex,
			Format:   render.VertexAttributeFormatRGBA32F,
			Offset:   definition.VertexFormat.ColorOffsetBytes,
		})
	}

	vertexArray := e.api.CreateVertexArray(render.VertexArrayInfo{
		Bindings: []render.VertexArrayBindingInfo{
			{
				VertexBuffer: vertexBuffer,
				Stride:       definition.VertexFormat.CoordStrideBytes, // FIXME: Not accurate
			},
		},
		Attributes:  attributes,
		IndexBuffer: indexBuffer,
		IndexFormat: e.convertIndexType(definition.IndexFormat),
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
			topology:         e.convertPrimitive(subMesh.Primitive),
			indexCount:       subMesh.IndexCount,
			indexOffsetBytes: subMesh.IndexOffset,
		}
	}
	return result
}

func (e *Engine) CreatePBRMaterial(definition graphics.PBRMaterialDefinition) graphics.Material {
	extractTwoDTexture := func(src graphics.TwoDTexture) render.Texture {
		if src == nil {
			return nil
		}
		return src.(*TwoDTexture).Texture
	}
	return &Material{
		backfaceCulling: definition.BackfaceCulling,
		alphaBlending:   definition.AlphaBlending,
		alphaTesting:    definition.AlphaTesting,
		alphaThreshold:  definition.AlphaThreshold,
		twoDTextures: []render.Texture{
			extractTwoDTexture(definition.AlbedoTexture),
			extractTwoDTexture(definition.NormalTexture),
			extractTwoDTexture(definition.MetalnessTexture),
			extractTwoDTexture(definition.RoughnessTexture),
		},
		cubeTextures: []render.Texture{},
		vectors: []sprec.Vec4{
			definition.AlbedoColor,
			sprec.NewVec4(definition.NormalScale, definition.Metalness, definition.Roughness, 0.0),
		},
		geometryPresentation: internal.NewPBRGeometryPresentation(e.api, e.shaders.PBRShaderSet(definition)),
		shadowPresentation:   nil, // TODO
	}
}

func (e *Engine) Destroy() {
	e.renderer.Release()
}

func (e *Engine) convertWrap(wrap graphics.Wrap) render.WrapMode {
	switch wrap {
	case graphics.WrapClampToEdge:
		return render.WrapModeClamp
	case graphics.WrapRepeat:
		return render.WrapModeRepeat
	case graphics.WrapMirroredRepat:
		return render.WrapModeMirroredRepeat
	default:
		panic(fmt.Errorf("unknown wrap mode: %d", wrap))
	}
}

func (e *Engine) needsMipmaps(filter graphics.Filter) bool {
	switch filter {
	case graphics.FilterNearestMipmapNearest:
		fallthrough
	case graphics.FilterNearestMipmapLinear:
		fallthrough
	case graphics.FilterLinearMipmapNearest:
		fallthrough
	case graphics.FilterLinearMipmapLinear:
		return true
	default:
		return false
	}
}

func (e *Engine) convertFilter(filter graphics.Filter, anisotropic bool) render.FilterMode {
	switch filter {
	case graphics.FilterNearest:
		return render.FilterModeNearest
	case graphics.FilterLinear:
		if anisotropic {
			return render.FilterModeAnisotropic
		}
		return render.FilterModeLinear
	case graphics.FilterNearestMipmapNearest:
		return render.FilterModeNearest
	case graphics.FilterNearestMipmapLinear:
		return render.FilterModeNearest
	case graphics.FilterLinearMipmapNearest:
		return render.FilterModeLinear
	case graphics.FilterLinearMipmapLinear:
		return render.FilterModeLinear
	default:
		panic(fmt.Errorf("unknown min filter mode: %d", filter))
	}
}

func (e *Engine) convertFormat(internalFormat graphics.InternalFormat, dataFormat graphics.DataFormat) render.DataFormat {
	switch dataFormat {
	case graphics.DataFormatRGBA8:
		return render.DataFormatRGBA8
	case graphics.DataFormatRGBA32F:
		return render.DataFormatRGBA32F
	default:
		panic(fmt.Errorf("unknown data format: %d", dataFormat))
	}
}

func (e *Engine) convertPrimitive(primitive graphics.Primitive) render.Topology {
	switch primitive {
	case graphics.PrimitivePoints:
		return render.TopologyPoints
	case graphics.PrimitiveLines:
		return render.TopologyLines
	case graphics.PrimitiveLineStrip:
		return render.TopologyLineStrip
	case graphics.PrimitiveLineLoop:
		return render.TopologyLineLoop
	case graphics.PrimitiveTriangles:
		return render.TopologyTriangles
	case graphics.PrimitiveTriangleStrip:
		return render.TopologyTriangleStrip
	case graphics.PrimitiveTriangleFan:
		return render.TopologyTriangleFan
	default:
		panic(fmt.Errorf("unknown primitive: %d", primitive))
	}
}

func (e *Engine) convertIndexType(indexFormat graphics.IndexFormat) render.IndexFormat {
	switch indexFormat {
	case graphics.IndexFormatU16:
		return render.IndexFormatUnsignedShort
	case graphics.IndexFormatU32:
		return render.IndexFormatUnsignedInt
	default:
		panic(fmt.Errorf("unknown index format: %d", indexFormat))
	}
}
