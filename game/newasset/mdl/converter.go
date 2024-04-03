package mdl

import (
	"bytes"
	"fmt"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/debug/log"
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
	assetSky := asset.Sky{
		NodeIndex: nodeIndex,
		Layers:    make([]asset.SkyLayer, len(sky.layers)),
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

	lslShader, err := c.parseShader(layer.shader)
	if err != nil {
		return asset.SkyLayer{}, fmt.Errorf("error parsing shader: %w", err)
	}
	// TODO: Run validation with sky "Globals"

	block := newUniformBlock()
	for _, field := range uniformFields(lslShader) {
		value := layer.Property(field.Name)
		switch field.Type {
		case lsl.TypeNameFloat:
			block.WriteFloat32(convert[float32](value))
		case lsl.TypeNameVec2:
			block.WriteSPVec2(convert[sprec.Vec2](value))
		case lsl.TypeNameVec3:
			block.WriteSPVec3(convert[sprec.Vec3](value))
		case lsl.TypeNameVec4:
			block.WriteSPVec4(convert[sprec.Vec4](value))
		default:
			return asset.SkyLayer{}, fmt.Errorf("unsupported uniform type %q", field.Type)
		}
	}

	return asset.SkyLayer{
		Blending:           layer.Blending(),
		Textures:           []asset.TextureBinding{}, // FIXME: Implement
		MaterialDataStd140: block.Data(),
		ShaderIndex:        shaderIndex,
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

func uniformFields(shader *lsl.Shader) []lsl.Field {
	for _, decl := range shader.Declarations {
		if uniformBlock, ok := decl.(*lsl.UniformBlockDeclaration); ok {
			return uniformBlock.Fields
		}
	}
	return nil
}

func convert[T any](value any) T {
	var result T
	if actual, ok := value.(T); ok {
		result = actual
	} else {
		log.Warn("failed to convert value %v to type %T", value, result)
	}
	return result
}

func newUniformBlock() *uniformBlock {
	data := &bytes.Buffer{}
	return &uniformBlock{
		data:   data,
		writer: gblob.NewLittleEndianWriter(data),
	}
}

type uniformBlock struct {
	data   *bytes.Buffer
	writer gblob.TypedWriter
	offset int
}

func (b *uniformBlock) Data() []byte {
	return b.data.Bytes()
}

func (b *uniformBlock) WriteFloat32(value float32) {
	b.ensureOffsetMultiple(4)
	b.writer.WriteFloat32(value)
	b.offset += 4
}

func (b *uniformBlock) WriteSPVec2(value sprec.Vec2) {
	b.ensureOffsetMultiple(8)
	b.writer.WriteFloat32(float32(value.X))
	b.writer.WriteFloat32(float32(value.Y))
	b.offset += 8
}

func (b *uniformBlock) WriteSPVec3(value sprec.Vec3) {
	b.ensureOffsetMultiple(16) // alignment is 16 bytes
	b.writer.WriteFloat32(float32(value.X))
	b.writer.WriteFloat32(float32(value.Y))
	b.writer.WriteFloat32(float32(value.Z))
	b.offset += 12
}

func (b *uniformBlock) WriteSPVec4(value sprec.Vec4) {
	b.ensureOffsetMultiple(16) // alignment is 16 bytes
	b.writer.WriteFloat32(float32(value.X))
	b.writer.WriteFloat32(float32(value.Y))
	b.writer.WriteFloat32(float32(value.Z))
	b.writer.WriteFloat32(float32(value.W))
	b.offset += 16
}

func (b *uniformBlock) ensureOffsetMultiple(multiple int) {
	remainder := b.offset % multiple
	if remainder != 0 {
		padding := multiple - remainder
		for range padding {
			b.writer.WriteUint8(0)
		}
		b.offset += padding
	}
}
