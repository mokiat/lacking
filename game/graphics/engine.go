package graphics

import (
	"encoding/binary"
	"fmt"

	"github.com/mokiat/lacking/data/buffer"
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

	freeFragmentID int
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

// PBRShading returns the Shading implementaton for phisically-based rendering.
func (e *Engine) PBRShading() Shading {
	return &pbrShading{
		api:     e.api,
		shaders: e.shaders,
	}
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

// CreateMaterialDefinition creates a new MaterialDefinition from the specified
// info object.
func (e *Engine) CreateMaterialDefinition(info MaterialDefinitionInfo) *MaterialDefinition {
	data := make([]byte, 4*4*len(info.Vectors))
	plotter := buffer.NewPlotter(data, binary.LittleEndian)
	for _, vec := range info.Vectors {
		plotter.PlotVec4(vec)
	}
	return &MaterialDefinition{
		revision:        1,
		backfaceCulling: info.BackfaceCulling,
		alphaTesting:    info.AlphaTesting,
		alphaBlending:   info.AlphaBlending,
		alphaThreshold:  info.AlphaThreshold,
		uniformData:     data,
		twoDTextures:    info.TwoDTextures,
		cubeTextures:    info.CubeTextures,
		shading:         info.Shading,
	}
}

// CreateMeshDefinition creates a new MeshDefinition using the specified
// info object.
func (e *Engine) CreateMeshDefinition(info MeshDefinitionInfo) *MeshDefinition {
	vertexBuffer := e.api.CreateVertexBuffer(render.BufferInfo{
		Dynamic: false,
		Data:    info.VertexData,
	})
	indexBuffer := e.api.CreateIndexBuffer(render.BufferInfo{
		Dynamic: false,
		Data:    info.IndexData,
	})

	var attributes []render.VertexArrayAttributeInfo
	if info.VertexFormat.HasCoord {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: internal.CoordAttributeIndex,
			Format:   render.VertexAttributeFormatRGB32F,
			Offset:   info.VertexFormat.CoordOffsetBytes,
		})
	}
	if info.VertexFormat.HasNormal {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: internal.NormalAttributeIndex,
			Format:   render.VertexAttributeFormatRGB16F,
			Offset:   info.VertexFormat.NormalOffsetBytes,
		})
	}
	if info.VertexFormat.HasTangent {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: internal.TangentAttributeIndex,
			Format:   render.VertexAttributeFormatRGB16F,
			Offset:   info.VertexFormat.TangentOffsetBytes,
		})
	}
	if info.VertexFormat.HasTexCoord {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: internal.TexCoordAttributeIndex,
			Format:   render.VertexAttributeFormatRG16F,
			Offset:   info.VertexFormat.TexCoordOffsetBytes,
		})
	}
	if info.VertexFormat.HasColor {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: internal.ColorAttributeIndex,
			Format:   render.VertexAttributeFormatRGBA8UN,
			Offset:   info.VertexFormat.ColorOffsetBytes,
		})
	}
	if info.VertexFormat.HasWeights {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: internal.WeightsAttributeIndex,
			Format:   render.VertexAttributeFormatRGBA8UN,
			Offset:   info.VertexFormat.WeightsOffsetBytes,
		})
	}
	if info.VertexFormat.HasJoints {
		attributes = append(attributes, render.VertexArrayAttributeInfo{
			Binding:  0,
			Location: internal.JointsAttributeIndex,
			Format:   render.VertexAttributeFormatRGBA8IU,
			Offset:   info.VertexFormat.JointsOffsetBytes,
		})
	}

	vertexArray := e.api.CreateVertexArray(render.VertexArrayInfo{
		Bindings: []render.VertexArrayBindingInfo{
			{
				VertexBuffer: vertexBuffer,
				Stride:       info.VertexFormat.CoordStrideBytes, // FIXME: Not accurate
			},
		},
		Attributes:  attributes,
		IndexBuffer: indexBuffer,
		IndexFormat: e.convertIndexType(info.IndexFormat),
	})

	result := &MeshDefinition{
		vertexBuffer:         vertexBuffer,
		indexBuffer:          indexBuffer,
		vertexArray:          vertexArray,
		fragments:            make([]MeshFragmentDefinition, len(info.Fragments)),
		boundingSphereRadius: info.BoundingSphereRadius,
		hasArmature:          info.HasArmature(),
	}
	for i, fragmentInfo := range info.Fragments {
		materialDef := fragmentInfo.Material
		fragmentDef := MeshFragmentDefinition{
			id:               e.freeFragmentID,
			mesh:             result,
			topology:         e.convertPrimitive(fragmentInfo.Primitive),
			indexCount:       fragmentInfo.IndexCount,
			indexOffsetBytes: fragmentInfo.IndexOffset,
			material: &Material{
				definitionRevision: materialDef.revision,
				definition:         materialDef,
			},
		}
		fragmentDef.rebuildPipelines()
		result.fragments[i] = fragmentDef
		e.freeFragmentID++
	}
	return result
}

// CreatePBRMaterialDefinition creates a new Material that is based on PBR
// definition.
// TODO: Remove this and create a PBR builder over MaterialDefinitionInfo.
func (e *Engine) CreatePBRMaterialDefinition(info PBRMaterialInfo) *MaterialDefinition {
	extractTwoDTexture := func(src *TwoDTexture) render.Texture {
		if src == nil {
			return nil
		}
		return src.texture
	}

	uniformData := make([]byte, 2*4*4)
	plotter := buffer.NewPlotter(uniformData, binary.LittleEndian)
	plotter.PlotVec4(info.AlbedoColor)
	plotter.PlotFloat32(info.AlphaThreshold)
	plotter.PlotFloat32(info.NormalScale)
	plotter.PlotFloat32(info.Metallic)
	plotter.PlotFloat32(info.Roughness)

	return &MaterialDefinition{
		revision:        1,
		backfaceCulling: info.BackfaceCulling,
		alphaBlending:   info.AlphaBlending,
		alphaTesting:    info.AlphaTesting,
		alphaThreshold:  info.AlphaThreshold,
		uniformData:     uniformData,
		twoDTextures: []render.Texture{
			extractTwoDTexture(info.AlbedoTexture),
			extractTwoDTexture(info.NormalTexture),
			extractTwoDTexture(info.MetallicRoughnessTexture),
		},
		cubeTextures: []render.Texture{},
		shading:      e.PBRShading(),
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
