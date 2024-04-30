package mdl

import (
	"fmt"
	"slices"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/asset"
)

func NewConverter(model *Model) *Converter {
	return &Converter{
		model: model,

		convertedShaders:   make(map[*Shader]uint32),
		convertedTextures:  make(map[*Texture]uint32),
		convertedMaterials: make(map[*Material]uint32),
	}
}

type Converter struct {
	model *Model

	assetShaders     []asset.Shader
	convertedShaders map[*Shader]uint32

	assetTextures     []asset.Texture
	convertedTextures map[*Texture]uint32

	assetMaterials     []asset.Material
	convertedMaterials map[*Material]uint32
}

func (c *Converter) Convert() (asset.Model, error) {
	return c.convertModel(c.model)
}

func (c *Converter) convertModel(s *Model) (asset.Model, error) {
	var (
		assetNodes             []asset.Node
		assetAmbientLights     []asset.AmbientLight
		assetPointLights       []asset.PointLight
		assetSpotLights        []asset.SpotLight
		assetDirectionalLights []asset.DirectionalLight
		assetSkies             []asset.Sky
	)

	nodes := s.FlattenNodes()

	nodeIndex := make(map[Node]uint32)

	for i, node := range nodes {
		nodeIndex[node] = uint32(i)

		parentIndex := asset.UnspecifiedNodeIndex
		if pIndex, ok := nodeIndex[node.Parent()]; ok {
			parentIndex = int32(pIndex)
		}

		assetNodes = append(assetNodes, asset.Node{
			Name:        node.Name(),
			ParentIndex: parentIndex,
			Translation: node.Translation(),
			Rotation:    node.Rotation(),
			Scale:       node.Scale(),
			Mask:        asset.NodeMaskNone,
		})

		switch essence := node.(type) {
		case *AmbientLight:
			ambientLightAsset, err := c.convertAmbientLight(uint32(i), essence)
			if err != nil {
				return asset.Model{}, fmt.Errorf("error converting ambient light %q: %w", node.Name(), err)
			}
			assetAmbientLights = append(assetAmbientLights, ambientLightAsset)
		case *PointLight:
			pointLightAsset := c.convertPointLight(uint32(i), essence)
			assetPointLights = append(assetPointLights, pointLightAsset)
		case *SpotLight:
			spotLightAsset := c.convertSpotLight(uint32(i), essence)
			assetSpotLights = append(assetSpotLights, spotLightAsset)
		case *DirectionalLight:
			directionalLightAsset := c.convertDirectionalLight(uint32(i), essence)
			assetDirectionalLights = append(assetDirectionalLights, directionalLightAsset)
		case *Sky:
			assetSky, err := c.convertSky(uint32(i), essence)
			if err != nil {
				return asset.Model{}, fmt.Errorf("error converting sky %q: %w", node.Name(), err)
			}
			assetSkies = append(assetSkies, assetSky)
		}
	}

	return asset.Model{
		Nodes:             assetNodes,
		Shaders:           c.assetShaders,
		Textures:          c.assetTextures,
		Materials:         c.assetMaterials,
		AmbientLights:     assetAmbientLights,
		PointLights:       assetPointLights,
		SpotLights:        assetSpotLights,
		DirectionalLights: assetDirectionalLights,
		Skies:             assetSkies,
	}, nil
}

func (c *Converter) convertMaterialPass(pass *MaterialPass) (asset.MaterialPass, error) {
	shaderIndex, err := c.convertShader(pass.shader)
	if err != nil {
		return asset.MaterialPass{}, fmt.Errorf("error converting shader: %w", err)
	}
	return asset.MaterialPass{
		Layer:           int32(pass.layer),
		Culling:         pass.culling,
		FrontFace:       pass.frontFace,
		DepthTest:       pass.depthTest,
		DepthWrite:      pass.depthWrite,
		DepthComparison: pass.depthComparison,
		Blending:        pass.blending,
		ShaderIndex:     shaderIndex,
	}, nil
}

func (c *Converter) convertMaterial(material *Material) (uint32, error) {
	if index, ok := c.convertedMaterials[material]; ok {
		return index, nil
	}

	textures, err := c.convertSamplers(material.samplers)
	if err != nil {
		return 0, fmt.Errorf("error converting samplers: %w", err)
	}

	properties, err := c.convertProperties(material.properties)
	if err != nil {
		return 0, fmt.Errorf("error converting properties: %w", err)
	}

	assetMaterial := asset.Material{
		Name:                 material.name,
		Textures:             textures,
		Properties:           properties,
		GeometryPasses:       make([]asset.MaterialPass, len(material.geometryPasses)),
		ShadowPasses:         make([]asset.MaterialPass, len(material.shadowPasses)),
		ForwardPasses:        make([]asset.MaterialPass, len(material.forwardPasses)),
		SkyPasses:            make([]asset.MaterialPass, len(material.skyPasses)),
		PostprocessingPasses: make([]asset.MaterialPass, len(material.postprocessingPasses)),
	}
	for i, pass := range material.geometryPasses {
		assetPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return 0, fmt.Errorf("error converting material pass: %w", err)
		}
		assetMaterial.GeometryPasses[i] = assetPass
	}
	for i, pass := range material.shadowPasses {
		assetPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return 0, fmt.Errorf("error converting material pass: %w", err)
		}
		assetMaterial.ShadowPasses[i] = assetPass
	}
	for i, pass := range material.forwardPasses {
		assetPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return 0, fmt.Errorf("error converting material pass: %w", err)
		}
		assetMaterial.ForwardPasses[i] = assetPass
	}
	for i, pass := range material.skyPasses {
		assetPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return 0, fmt.Errorf("error converting material pass: %w", err)
		}
		assetMaterial.SkyPasses[i] = assetPass
	}
	for i, pass := range material.postprocessingPasses {
		assetPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return 0, fmt.Errorf("error converting material pass: %w", err)
		}
		assetMaterial.PostprocessingPasses[i] = assetPass
	}

	index := uint32(len(c.assetMaterials))
	c.assetMaterials = append(c.assetMaterials, assetMaterial)
	c.convertedMaterials[material] = index
	return index, nil

}

func (c *Converter) convertAmbientLight(nodeIndex uint32, light *AmbientLight) (asset.AmbientLight, error) {
	reflectionTextureIndex, err := c.convertTexture(light.reflectionTexture)
	if err != nil {
		return asset.AmbientLight{}, fmt.Errorf("error converting reflection texture: %w", err)
	}

	refractionTextureIndex, err := c.convertTexture(light.refractionTexture)
	if err != nil {
		return asset.AmbientLight{}, fmt.Errorf("error converting refraction texture: %w", err)
	}

	return asset.AmbientLight{
		NodeIndex:              nodeIndex,
		ReflectionTextureIndex: reflectionTextureIndex,
		RefractionTextureIndex: refractionTextureIndex,
		CastShadow:             light.CastShadow(),
	}, nil
}

func (c *Converter) convertPointLight(nodeIndex uint32, light *PointLight) asset.PointLight {
	return asset.PointLight{
		NodeIndex:    nodeIndex,
		EmitColor:    light.EmitColor(),
		EmitDistance: light.EmitDistance(),
		CastShadow:   light.CastShadow(),
	}
}

func (c *Converter) convertSpotLight(nodeIndex uint32, light *SpotLight) asset.SpotLight {
	return asset.SpotLight{
		NodeIndex:      nodeIndex,
		EmitColor:      light.EmitColor(),
		EmitDistance:   light.EmitDistance(),
		EmitAngleOuter: light.EmitAngleOuter(),
		EmitAngleInner: light.EmitAngleInner(),
		CastShadow:     light.CastShadow(),
	}
}

func (c *Converter) convertDirectionalLight(nodeIndex uint32, light *DirectionalLight) asset.DirectionalLight {
	return asset.DirectionalLight{
		NodeIndex:  nodeIndex,
		EmitColor:  light.EmitColor(),
		CastShadow: light.CastShadow(),
	}
}

func (c *Converter) convertSky(nodeIndex uint32, sky *Sky) (asset.Sky, error) {

	materialIndex, err := c.convertMaterial(sky.material)
	if err != nil {
		return asset.Sky{}, fmt.Errorf("error converting material: %w", err)
	}

	assetSky := asset.Sky{
		NodeIndex:     nodeIndex,
		MaterialIndex: materialIndex,
	}
	return assetSky, nil
}

func (c *Converter) convertShader(shader *Shader) (uint32, error) {
	if index, ok := c.convertedShaders[shader]; ok {
		return index, nil
	}
	shaderIndex := uint32(len(c.assetShaders))
	assetShader := asset.Shader{
		ShaderType: shader.ShaderType(),
		SourceCode: shader.SourceCode(),
	}
	c.convertedShaders[shader] = shaderIndex
	c.assetShaders = append(c.assetShaders, assetShader)
	return shaderIndex, nil
}

func (c *Converter) convertSamplers(samplers map[string]*Sampler) ([]asset.TextureBinding, error) {
	bindings := make([]asset.TextureBinding, 0, len(samplers))
	for name, sampler := range samplers {
		textureIndex, err := c.convertTexture(sampler.texture)
		if err != nil {
			return nil, fmt.Errorf("error converting texture: %w", err)
		}
		bindings = append(bindings, asset.TextureBinding{
			BindingName:  name,
			TextureIndex: textureIndex,
			Wrapping:     sampler.wrapMode,
			Filtering:    sampler.filterMode,
			Mipmapping:   sampler.mipmapping,
		})
	}
	return bindings, nil
}

func isLikelyLinearSpace(format TextureFormat) bool {
	linearFormats := []TextureFormat{
		TextureFormatRGBA16F,
		TextureFormatRGBA32F,
	}
	return slices.Contains(linearFormats, format)
}

func (c *Converter) convertTexture(texture *Texture) (uint32, error) {
	if index, ok := c.convertedTextures[texture]; ok {
		return index, nil
	}

	var flags asset.TextureFlag
	switch texture.Kind() {
	case TextureKind2D:
		flags = asset.TextureFlag2D
	case TextureKind2DArray:
		flags = asset.TextureFlag2DArray
	case TextureKind3D:
		flags = asset.TextureFlag3D
	case TextureKindCube:
		flags = asset.TextureFlagCubeMap
	default:
		return 0, fmt.Errorf("unsupported texture kind %d", texture.Kind())
	}
	if isLikelyLinearSpace(texture.format) {
		flags |= asset.TextureFlagLinearSpace
	}
	assetTexture := asset.Texture{
		Width:  uint32(texture.Width()),
		Height: uint32(texture.Height()),
		Format: texture.Format(),
		Flags:  flags,
		Layers: gog.Map(texture.layers, func(layer TextureLayer) asset.TextureLayer {
			return asset.TextureLayer{
				Data: layer.Data(),
			}
		}),
	}

	index := uint32(len(c.assetTextures))
	c.assetTextures = append(c.assetTextures, assetTexture)
	c.convertedTextures[texture] = index
	return index, nil
}

func (c *Converter) convertProperties(properties map[string]interface{}) ([]asset.PropertyBinding, error) {
	bindings := make([]asset.PropertyBinding, 0, len(properties))
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
		bindings = append(bindings, asset.PropertyBinding{
			BindingName: name,
			Data:        data,
		})
	}
	return bindings, nil
}
