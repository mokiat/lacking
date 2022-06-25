package graphics

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
)

func NewEngine(api render.API, shaders ShaderCollection) *Engine {
	return &Engine{
		api:      api,
		shaders:  shaders,
		renderer: newRenderer(api, shaders),
	}
}

// Engine represents an entrypoint to 3D graphics rendering.
type Engine struct {
	api      render.API
	shaders  ShaderCollection
	renderer *sceneRenderer
}

// Create initializes this 3D engine.
func (e *Engine) Create() {
	e.renderer.Allocate()
}

// Destroy releases resources allocated by this
// 3D engine.
func (e *Engine) Destroy() {
	e.renderer.Release()
}

// CreateScene creates a new 3D Scene. Entities managed
// within a given scene are isolated within that scene.
func (e *Engine) CreateScene() *Scene {
	return newScene(e.renderer)
}

// CreateTwoDTexture creates a new TwoDTexture using the
// specified definition.
func (e *Engine) CreateTwoDTexture(definition TwoDTextureDefinition) *TwoDTexture {
	return newTwoDTexture(e.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           definition.Width,
		Height:          definition.Height,
		Wrapping:        e.convertWrap(definition.Wrapping),
		Filtering:       e.convertFilter(definition.Filtering),
		Mipmapping:      definition.GenerateMipmaps,
		GammaCorrection: true,
		Format:          e.convertFormat(definition.InternalFormat, definition.DataFormat),
		Data:            definition.Data,
	}))
}

// CreateCubeTexture creates a new CubeTexture using the
// specified definition.
func (e *Engine) CreateCubeTexture(definition CubeTextureDefinition) *CubeTexture {
	return newCubeTexture(e.api.CreateColorTextureCube(render.ColorTextureCubeInfo{
		Dimension:       definition.Dimension,
		Filtering:       e.convertFilter(definition.Filtering),
		Mipmapping:      definition.GenerateMipmaps,
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

// CreateMeshTemplate creates a new MeshTemplate using the specified
// definition.
func (e *Engine) CreateMeshTemplate(definition MeshTemplateDefinition) *MeshTemplate {
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
			Location: internal.CoordAttributeIndex,
			Format:   render.VertexAttributeFormatRGB32F,
			Offset:   definition.VertexFormat.CoordOffsetBytes,
		})
	}
	if definition.VertexFormat.HasNormal {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: internal.NormalAttributeIndex,
			Format:   render.VertexAttributeFormatRGB16F,
			Offset:   definition.VertexFormat.NormalOffsetBytes,
		})
	}
	if definition.VertexFormat.HasTangent {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: internal.TangentAttributeIndex,
			Format:   render.VertexAttributeFormatRGB32F,
			Offset:   definition.VertexFormat.TangentOffsetBytes,
		})
	}
	if definition.VertexFormat.HasTexCoord {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: internal.TexCoordAttributeIndex,
			Format:   render.VertexAttributeFormatRG32F,
			Offset:   definition.VertexFormat.TexCoordOffsetBytes,
		})
	}
	if definition.VertexFormat.HasColor {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: internal.ColorAttributeIndex,
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
		subMeshes:    make([]subMeshTemplate, len(definition.SubMeshes)),
	}
	for i, subMesh := range definition.SubMeshes {
		result.subMeshes[i] = subMeshTemplate{
			material:         subMesh.Material,
			topology:         e.convertPrimitive(subMesh.Primitive),
			indexCount:       subMesh.IndexCount,
			indexOffsetBytes: subMesh.IndexOffset,
		}
	}
	return result
}

// CreatePBRMaterial creates a new Material that is based on PBR
// definition.
func (e *Engine) CreatePBRMaterial(definition PBRMaterialDefinition) *Material {
	extractTwoDTexture := func(src *TwoDTexture) render.Texture {
		if src == nil {
			return nil
		}
		return src.texture
	}
	shaderSet := e.shaders.PBRShaderSet(definition)
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
		geometryPresentation: internal.NewGeometryPresentation(e.api,
			shaderSet.VertexShader(),
			shaderSet.FragmentShader(),
		),
		shadowPresentation: nil, // TODO
	}
}

func (e *Engine) convertWrap(wrap Wrap) render.WrapMode {
	switch wrap {
	case WrapClampToEdge:
		return render.WrapModeClamp
	case WrapRepeat:
		return render.WrapModeRepeat
	case WrapMirroredRepat:
		return render.WrapModeMirroredRepeat
	default:
		panic(fmt.Errorf("unknown wrap mode: %d", wrap))
	}
}

func (e *Engine) convertFilter(filter Filter) render.FilterMode {
	switch filter {
	case FilterNearest:
		return render.FilterModeNearest
	case FilterLinear:
		return render.FilterModeLinear
	case FilterAnisotropic:
		return render.FilterModeAnisotropic
	default:
		panic(fmt.Errorf("unknown min filter mode: %d", filter))
	}
}

func (e *Engine) convertFormat(internalFormat InternalFormat, dataFormat DataFormat) render.DataFormat {
	switch dataFormat {
	case DataFormatRGBA8:
		return render.DataFormatRGBA8
	case DataFormatRGBA16F:
		return render.DataFormatRGBA16F
	case DataFormatRGBA32F:
		return render.DataFormatRGBA32F
	default:
		panic(fmt.Errorf("unknown data format: %d", dataFormat))
	}
}

func (e *Engine) convertPrimitive(primitive Primitive) render.Topology {
	switch primitive {
	case PrimitivePoints:
		return render.TopologyPoints
	case PrimitiveLines:
		return render.TopologyLines
	case PrimitiveLineStrip:
		return render.TopologyLineStrip
	case PrimitiveLineLoop:
		return render.TopologyLineLoop
	case PrimitiveTriangles:
		return render.TopologyTriangles
	case PrimitiveTriangleStrip:
		return render.TopologyTriangleStrip
	case PrimitiveTriangleFan:
		return render.TopologyTriangleFan
	default:
		panic(fmt.Errorf("unknown primitive: %d", primitive))
	}
}

func (e *Engine) convertIndexType(indexFormat IndexFormat) render.IndexFormat {
	switch indexFormat {
	case IndexFormatU16:
		return render.IndexFormatUnsignedShort
	case IndexFormatU32:
		return render.IndexFormatUnsignedInt
	default:
		panic(fmt.Errorf("unknown index format: %d", indexFormat))
	}
}

type ShaderCollection struct {
	ExposureSet         func() ShaderSet
	PostprocessingSet   func(mapping ToneMapping) ShaderSet
	DirectionalLightSet func() ShaderSet
	AmbientLightSet     func() ShaderSet
	SkyboxSet           func() ShaderSet
	SkycolorSet         func() ShaderSet
	PBRShaderSet        func(definition PBRMaterialDefinition) ShaderSet
}

type ShaderSet struct {
	VertexShader   func() string
	FragmentShader func() string
}

type ToneMapping string

const (
	ReinhardToneMapping    ToneMapping = "reinhard"
	ExponentialToneMapping ToneMapping = "exponential"
)
