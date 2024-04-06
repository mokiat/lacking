package mdl

import (
	"fmt"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics/lsl"
	asset "github.com/mokiat/lacking/game/newasset"
)

func NewConverter(model *Model) *Converter {
	return &Converter{
		model: model,

		convertedSkyShaders: make(map[*Shader]uint32),

		parsedShaders: make(map[*Shader]*lsl.Shader),
	}
}

type Converter struct {
	model *Model

	assetSkyShaders     []asset.Shader
	convertedSkyShaders map[*Shader]uint32

	parsedShaders map[*Shader]*lsl.Shader
}

func (c *Converter) Convert() (asset.Model, error) {
	return c.convertModel(c.model)
}

func (c *Converter) convertModel(s *Model) (asset.Model, error) {
	var (
		assetNodes       []asset.Node
		assetPointLights []asset.PointLight
		assetSpotLights  []asset.SpotLight
		assetSkies       []asset.Sky
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
		case *PointLight:
			pointLightAsset := c.convertPointLight(uint32(i), essence)
			assetPointLights = append(assetPointLights, pointLightAsset)
		case *SpotLight:
			spotLightAsset := c.convertSpotLight(uint32(i), essence)
			assetSpotLights = append(assetSpotLights, spotLightAsset)
		case *Sky:
			assetSky, err := c.convertSky(uint32(i), essence)
			if err != nil {
				return asset.Model{}, fmt.Errorf("error converting sky %q: %w", node.Name(), err)
			}
			assetSkies = append(assetSkies, assetSky)
		}
	}

	return asset.Model{
		Nodes:       assetNodes,
		SkyShaders:  c.assetSkyShaders,
		PointLights: assetPointLights,
		SpotLights:  assetSpotLights,
		Skies:       assetSkies,
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

func (c *Converter) convertSky(nodeIndex uint32, sky *Sky) (asset.Sky, error) {
	properties, err := c.convertProperties(sky.properties)
	if err != nil {
		return asset.Sky{}, fmt.Errorf("error converting sky properties: %w", err)
	}

	assetSky := asset.Sky{
		NodeIndex:  nodeIndex,
		Textures:   []asset.TextureBinding{}, // TODO
		Properties: properties,
		Layers:     make([]asset.SkyLayer, len(sky.layers)),
	}
	for i, layer := range sky.layers {
		assetLayer, err := c.convertSkyLayer(layer)
		if err != nil {
			return asset.Sky{}, fmt.Errorf("error converting sky layer: %w", err)
		}
		assetSky.Layers[i] = assetLayer
	}
	return assetSky, nil
}

func (c *Converter) convertSkyLayer(layer SkyLayer) (asset.SkyLayer, error) {
	shaderIndex, err := c.convertSkyShader(layer.shader)
	if err != nil {
		return asset.SkyLayer{}, fmt.Errorf("error converting shader: %w", err)
	}

	_, err = c.parseShader(layer.shader)
	if err != nil {
		return asset.SkyLayer{}, fmt.Errorf("error parsing shader: %w", err)
	}
	// TODO: Run validation with sky "Globals"

	return asset.SkyLayer{
		Blending:    layer.Blending(),
		ShaderIndex: shaderIndex,
	}, nil
}

func (c *Converter) convertSkyShader(shader *Shader) (uint32, error) {
	if index, ok := c.convertedSkyShaders[shader]; ok {
		return index, nil
	}
	shaderIndex := uint32(len(c.assetSkyShaders))
	assetShader := asset.Shader{
		SourceCode: shader.SourceCode(),
	}
	c.convertedSkyShaders[shader] = shaderIndex
	c.assetSkyShaders = append(c.assetSkyShaders, assetShader)
	return shaderIndex, nil
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

func (c *Converter) parseShader(shader *Shader) (*lsl.Shader, error) {
	if parsed, ok := c.parsedShaders[shader]; ok {
		return parsed, nil
	}
	parsed, err := lsl.Parse(shader.SourceCode())
	if err != nil {
		return nil, err
	}
	c.parsedShaders[shader] = parsed
	return parsed, nil
}
