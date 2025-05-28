package conv

import (
	"fmt"
	"slices"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/game/graphics/lsl"
	"github.com/mokiat/lacking/storage/chunked"
)

type ShadingSource interface {
	AllShaders() []*mdl.Shader
	AllTextures() []*mdl.Texture
	AllMaterials() []*mdl.Material
}

func NewShadingConverter() *ShadingConverter {
	return &ShadingConverter{}
}

type ShadingConverter struct{}

func (c *ShadingConverter) Convert(target *ds.List[chunked.Chunk], asset any) error {
	src, ok := asset.(ShadingSource)
	if !ok {
		return nil
	}
	chunk, err := c.CreateShadingChunk(src)
	if err != nil {
		return err
	}
	target.Add(chunked.FromValue(dto.ShadingChunkID, chunk))
	return nil
}

func (c *ShadingConverter) CreateShadingChunk(src ShadingSource) (*dto.ShadingChunk, error) {
	allShaders := src.AllShaders()
	dtoShaders := make([]dto.Shader, len(allShaders))
	for i, shader := range allShaders {
		var err error
		dtoShaders[i], err = c.convertShader(shader)
		if err != nil {
			return nil, fmt.Errorf("error converting shader: %w", err)
		}
	}

	allTextures := src.AllTextures()
	dtoTextures := make([]dto.Texture, len(allTextures))
	for i, texture := range allTextures {
		var err error
		dtoTextures[i], err = c.convertTexture(texture)
		if err != nil {
			return nil, fmt.Errorf("error converting texture: %w", err)
		}
	}

	allMaterials := src.AllMaterials()
	dtoMaterials := make([]dto.Material, len(allMaterials))
	for i, material := range allMaterials {
		var err error
		dtoMaterials[i], err = c.convertMaterial(material)
		if err != nil {
			return nil, fmt.Errorf("error converting material: %w", err)
		}
	}

	return &dto.ShadingChunk{
		Shaders:   dtoShaders,
		Textures:  dtoTextures,
		Materials: dtoMaterials,
	}, nil
}

func (c *ShadingConverter) convertShader(shader *mdl.Shader) (dto.Shader, error) {
	ast, err := lsl.Parse(shader.SourceCode())
	if err != nil {
		return dto.Shader{}, fmt.Errorf("error parsing shader: %w", err)
	}
	var schema lsl.Schema
	switch shader.ShaderType() {
	case mdl.ShaderTypeGeometry:
		schema = lsl.GeometrySchema()
	case mdl.ShaderTypeShadow:
		schema = lsl.ShadowSchema()
	case mdl.ShaderTypeForward:
		schema = lsl.ForwardSchema()
	case mdl.ShaderTypeSky:
		schema = lsl.SkySchema()
	case mdl.ShaderTypePostprocess:
		schema = lsl.PostprocessSchema()
	default:
		schema = lsl.DefaultSchema()
	}
	if err := lsl.Validate(ast, schema); err != nil {
		return dto.Shader{}, fmt.Errorf("error validating shader: %w", err)
	}
	return dto.Shader{
		ID:         shader.ID(),
		ShaderType: shader.ShaderType(),
		SourceCode: shader.SourceCode(),
	}, nil
}

func (c *ShadingConverter) convertTexture(texture *mdl.Texture) (dto.Texture, error) {
	var flags dto.TextureFlag
	switch texture.Kind() {
	case mdl.TextureKind2D:
		flags = dto.TextureFlag2D
	case mdl.TextureKind2DArray:
		flags = dto.TextureFlag2DArray
	case mdl.TextureKind3D:
		flags = dto.TextureFlag3D
	case mdl.TextureKindCube:
		flags = dto.TextureFlagCubeMap
	default:
		return dto.Texture{}, fmt.Errorf("unsupported texture kind %d", texture.Kind())
	}
	if c.isLikelyLinearSpace(texture.Format()) || texture.Linear() {
		flags |= dto.TextureFlagLinearSpace
	}
	if texture.GenerateMipmaps() {
		flags |= dto.TextureFlagMipmapping
	}
	return dto.Texture{
		ID:     texture.ID(),
		Format: texture.Format(),
		Flags:  flags,
		MipmapLayers: gog.Map(texture.MipmapLayers(), func(mipLayer mdl.MipmapLayer) dto.MipmapLayer {
			return dto.MipmapLayer{
				Width:  uint32(mipLayer.Width()),
				Height: uint32(mipLayer.Height()),
				Depth:  uint32(mipLayer.Depth()),
				Layers: gog.Map(mipLayer.Layers(), func(layer mdl.TextureLayer) dto.TextureLayer {
					return dto.TextureLayer{
						Data: layer.Data(),
					}
				}),
			}
		}),
	}, nil
}

func (c *ShadingConverter) isLikelyLinearSpace(format mdl.TextureFormat) bool {
	linearFormats := []mdl.TextureFormat{
		mdl.TextureFormatRGBA16F,
		mdl.TextureFormatRGBA32F,
	}
	return slices.Contains(linearFormats, format)
}

func (c *ShadingConverter) convertMaterial(material *mdl.Material) (dto.Material, error) {
	textures, err := c.convertSamplers(material.Samplers())
	if err != nil {
		return dto.Material{}, fmt.Errorf("error converting samplers: %w", err)
	}

	properties, err := c.convertProperties(material.Properties())
	if err != nil {
		return dto.Material{}, fmt.Errorf("error converting properties: %w", err)
	}

	dtoMaterial := dto.Material{
		ID:                   material.ID(),
		Name:                 material.Name(),
		Textures:             textures,
		Properties:           properties,
		GeometryPasses:       make([]dto.MaterialPass, len(material.GeometryPasses())),
		ShadowPasses:         make([]dto.MaterialPass, len(material.ShadowPasses())),
		ForwardPasses:        make([]dto.MaterialPass, len(material.ForwardPasses())),
		SkyPasses:            make([]dto.MaterialPass, len(material.SkyPasses())),
		PostprocessingPasses: make([]dto.MaterialPass, len(material.PostprocessingPasses())),
	}
	for i, pass := range material.GeometryPasses() {
		dtoPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return dto.Material{}, fmt.Errorf("error converting material pass: %w", err)
		}
		dtoMaterial.GeometryPasses[i] = dtoPass
	}
	for i, pass := range material.ShadowPasses() {
		dtoPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return dto.Material{}, fmt.Errorf("error converting material pass: %w", err)
		}
		dtoMaterial.ShadowPasses[i] = dtoPass
	}
	for i, pass := range material.ForwardPasses() {
		dtoPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return dto.Material{}, fmt.Errorf("error converting material pass: %w", err)
		}
		dtoMaterial.ForwardPasses[i] = dtoPass
	}
	for i, pass := range material.SkyPasses() {
		dtoPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return dto.Material{}, fmt.Errorf("error converting material pass: %w", err)
		}
		dtoMaterial.SkyPasses[i] = dtoPass
	}
	for i, pass := range material.PostprocessingPasses() {
		dtoPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return dto.Material{}, fmt.Errorf("error converting material pass: %w", err)
		}
		dtoMaterial.PostprocessingPasses[i] = dtoPass
	}

	return dtoMaterial, nil
}

func (c *ShadingConverter) convertSamplers(samplers map[string]*mdl.Sampler) ([]dto.TextureBinding, error) {
	bindings := make([]dto.TextureBinding, 0, len(samplers))
	for name, sampler := range samplers {
		bindings = append(bindings, dto.TextureBinding{
			BindingName: name,
			TextureID:   sampler.Texture().ID(),
			Wrapping:    sampler.WrapMode(),
			Filtering:   sampler.FilterMode(),
			Mipmapping:  sampler.Mipmapping(),
		})
	}
	return bindings, nil
}

func (c *ShadingConverter) convertProperties(properties map[string]interface{}) ([]dto.PropertyBinding, error) {
	bindings := make([]dto.PropertyBinding, 0, len(properties))
	for name, value := range properties {
		var data gblob.LittleEndianBlock
		switch value := value.(type) {
		case float32:
			data = make(gblob.LittleEndianBlock, 4)
			data.SetFloat32(0, value)
		case sprec.Vec2:
			data = make(gblob.LittleEndianBlock, 8)
			data.SetFloat32(0, value.X)
			data.SetFloat32(4, value.Y)
		case sprec.Vec3:
			data = make(gblob.LittleEndianBlock, 12)
			data.SetFloat32(0, value.X)
			data.SetFloat32(4, value.Y)
			data.SetFloat32(8, value.Z)
		case sprec.Vec4:
			data = make(gblob.LittleEndianBlock, 16)
			data.SetFloat32(0, value.X)
			data.SetFloat32(4, value.Y)
			data.SetFloat32(8, value.Z)
			data.SetFloat32(12, value.W)
		default:
			return nil, fmt.Errorf("unsupported property type %T", value)
		}
		bindings = append(bindings, dto.PropertyBinding{
			BindingName: name,
			Data:        data,
		})
	}
	return bindings, nil
}

func (c *ShadingConverter) convertMaterialPass(pass *mdl.MaterialPass) (dto.MaterialPass, error) {
	return dto.MaterialPass{
		Layer:           int32(pass.Layer()),
		Culling:         pass.Culling(),
		FrontFace:       pass.FrontFace(),
		DepthTest:       pass.DepthTest(),
		DepthWrite:      pass.DepthWrite(),
		DepthComparison: pass.DepthComparison(),
		Blending:        pass.Blending(),
		ShaderID:        pass.Shader().ID(),
	}, nil
}
