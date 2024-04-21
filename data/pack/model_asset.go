package pack

import (
	"fmt"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/stod"
	"github.com/x448/float16"

	"github.com/mokiat/lacking/game/asset"
	newasset "github.com/mokiat/lacking/game/newasset"
)

type SaveModelAssetOption func(a *SaveModelAssetAction)

func WithCollisionMesh(collisionMesh bool) SaveModelAssetOption {
	return func(a *SaveModelAssetAction) {
		a.forceCollidable = collisionMesh
	}
}

type SaveModelAssetAction struct {
	resource        asset.Resource
	modelProvider   ModelProvider
	forceCollidable bool
}

func (a *SaveModelAssetAction) Describe() string {
	return fmt.Sprintf("save_model_asset(%q)", a.resource.Name())
}

func (a *SaveModelAssetAction) Run() error {
	conv := newConverter(a.forceCollidable)
	modelAsset := conv.BuildModel(a.modelProvider.Model())
	if err := a.resource.WriteContent(modelAsset); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	return nil
}

func newConverter(collisionMeshes bool) *converter {
	return &converter{
		forceCollidable:                       collisionMeshes,
		assetNodes:                            make([]newasset.Node, 0),
		assetNodeIndexFromNode:                make(map[*Node]int),
		assetGeometryShaderIndexFromMaterial:  make(map[*Material]int),
		assetShadowShaderIndexFromMaterial:    make(map[*Material]int),
		assetMaterialIndexFromMaterial:        make(map[*Material]int),
		assetArmatureIndexFromArmature:        make(map[*Armature]int),
		assetGeometriesFromMeshDefinition:     make(map[*MeshDefinition]int),
		assetMeshDefinitionFromMeshDefinition: make(map[*MeshDefinition]int),
		assetBodyMaterialFromMeshDefinition:   make(map[*MeshDefinition]int),
		assetBodyDefinitionFromMeshDefinition: make(map[*MeshDefinition]int),
	}
}

type converter struct {
	forceCollidable                       bool
	assetNodes                            []newasset.Node
	assetNodeIndexFromNode                map[*Node]int
	assetGeometryShaderIndexFromMaterial  map[*Material]int
	assetShadowShaderIndexFromMaterial    map[*Material]int
	assetMaterialIndexFromMaterial        map[*Material]int
	assetArmatureIndexFromArmature        map[*Armature]int
	assetGeometriesFromMeshDefinition     map[*MeshDefinition]int
	assetMeshDefinitionFromMeshDefinition map[*MeshDefinition]int
	assetBodyMaterialFromMeshDefinition   map[*MeshDefinition]int
	assetBodyDefinitionFromMeshDefinition map[*MeshDefinition]int
}

func (c *converter) BuildModel(model *Model) *asset.Model {
	for _, node := range model.RootNodes {
		c.BuildNode(-1, node)
	}

	assetGeometryShaders := make([]newasset.Shader, len(model.Materials))
	assetShadowShaders := make([]newasset.Shader, len(model.Materials))
	for i, material := range model.Materials {
		assetGeometryShaders[i] = c.BuildGeometryShader(material)
		c.assetGeometryShaderIndexFromMaterial[material] = i

		assetShadowShaders[i] = c.BuildShadowShader(material)
		c.assetShadowShaderIndexFromMaterial[material] = i
	}

	assetTextures := make([]newasset.Texture, len(model.Textures))
	for i, texture := range model.Textures {
		assetTextures[i] = BuildTwoDTextureAsset(texture)
	}

	assetAnimations := make([]newasset.Animation, len(model.Animations))
	for i, animation := range model.Animations {
		assetAnimations[i] = c.BuildAnimation(animation)
	}

	assetMaterials := make([]newasset.Material, len(model.Materials))
	for i, material := range model.Materials {
		assetMaterials[i] = c.BuildMaterial(material)
		c.assetMaterialIndexFromMaterial[material] = i
	}

	assetArmatures := make([]newasset.Armature, len(model.Armatures))
	for i, armature := range model.Armatures {
		assetArmatures[i] = c.BuildArmature(armature)
		c.assetArmatureIndexFromArmature[armature] = i
	}

	assetGeometries := make([]newasset.Geometry, len(model.MeshDefinitions))
	for i, meshDefinition := range model.MeshDefinitions {
		assetGeometries[i] = c.BuildGeometry(meshDefinition)
		c.assetGeometriesFromMeshDefinition[meshDefinition] = i
	}

	assetMeshDefinitions := make([]newasset.MeshDefinition, len(model.MeshDefinitions))
	for i, meshDefinition := range model.MeshDefinitions {
		assetMeshDefinitions[i] = c.BuildMeshDefinition(meshDefinition)
		c.assetMeshDefinitionFromMeshDefinition[meshDefinition] = i
	}

	assetMeshInstances := make([]newasset.Mesh, len(model.MeshInstances))
	for i, meshInstance := range model.MeshInstances {
		assetMeshInstances[i] = c.BuildMeshInstance(meshInstance)
	}

	assetBodyMaterials := make([]newasset.BodyMaterial, len(model.MeshDefinitions))
	for i, meshDefinition := range model.MeshDefinitions {
		assetBodyMaterials[i] = c.BuildBodyMaterial(meshDefinition)
		c.assetBodyMaterialFromMeshDefinition[meshDefinition] = i
	}

	assetBodyDefinitions := make([]newasset.BodyDefinition, len(model.MeshDefinitions))
	for i, meshDefinition := range model.MeshDefinitions {
		assetBodyDefinitions[i] = c.BuildBodyDefinition(meshDefinition)
		c.assetBodyDefinitionFromMeshDefinition[meshDefinition] = i
	}

	assetBodyInstances := make([]newasset.Body, 0, len(model.MeshInstances))
	for _, meshInstance := range model.MeshInstances {
		if c.forceCollidable || meshInstance.HasCollision() {
			assetBodyInstances = append(assetBodyInstances, c.BuildBodyInstance(meshInstance))
		}
	}

	assetPointLights := make([]newasset.PointLight, 0)
	assetSpotLights := make([]newasset.SpotLight, 0)
	assetDirectionalLights := make([]newasset.DirectionalLight, 0)
	for _, lightInstance := range model.LightInstances {
		lightDefinition := lightInstance.Definition
		nodeIndex, ok := c.assetNodeIndexFromNode[lightInstance.Node]
		if !ok {
			panic(fmt.Errorf("node %s not found", lightInstance.Node.Name))
		}
		switch lightDefinition.Type {
		case LightTypePoint:
			assetPointLights = append(assetPointLights, newasset.PointLight{
				NodeIndex:    uint32(nodeIndex),
				EmitColor:    lightDefinition.EmitColor,
				EmitDistance: lightDefinition.EmitRange,
				CastShadow:   false,
			})
		case LightTypeSpot:
			assetSpotLights = append(assetSpotLights, newasset.SpotLight{
				NodeIndex:      uint32(nodeIndex),
				EmitColor:      lightDefinition.EmitColor,
				EmitDistance:   lightDefinition.EmitRange,
				EmitAngleOuter: lightDefinition.EmitOuterConeAngle,
				EmitAngleInner: lightDefinition.EmitInnerConeAngle,
				CastShadow:     false,
			})
		case LightTypeDirectional:
			assetDirectionalLights = append(assetDirectionalLights, newasset.DirectionalLight{
				NodeIndex:  uint32(nodeIndex),
				EmitColor:  lightDefinition.EmitColor,
				CastShadow: false,
			})
		default:
			panic(fmt.Errorf("unknown light type %q", lightInstance.Definition.Type))
		}
	}

	return &asset.Model{
		Nodes:             c.assetNodes,
		GeometryShaders:   assetGeometryShaders,
		ShadowShaders:     assetShadowShaders,
		Animations:        assetAnimations,
		Armatures:         assetArmatures,
		Textures:          assetTextures,
		Materials:         assetMaterials,
		Geometries:        assetGeometries,
		MeshDefinitions:   assetMeshDefinitions,
		MeshInstances:     assetMeshInstances,
		BodyMaterials:     assetBodyMaterials,
		BodyDefinitions:   assetBodyDefinitions,
		BodyInstances:     assetBodyInstances,
		PointLights:       assetPointLights,
		SpotLights:        assetSpotLights,
		DirectionalLights: assetDirectionalLights,
	}
}

func (c *converter) BuildNode(parentIndex int, node *Node) {
	result := newasset.Node{
		Name:        node.Name,
		ParentIndex: int32(parentIndex),
		Translation: node.Translation,
		Rotation:    node.Rotation,
		Scale:       node.Scale,
	}
	index := len(c.assetNodes)
	c.assetNodes = append(c.assetNodes, result)
	c.assetNodeIndexFromNode[node] = index
	for _, child := range node.Children {
		c.BuildNode(index, child)
	}
}

func (c *converter) BuildAnimation(animation *Animation) newasset.Animation {
	assetAnimation := newasset.Animation{
		Name:      animation.Name,
		StartTime: animation.StartTime,
		EndTime:   animation.EndTime,
		Bindings:  make([]newasset.AnimationBinding, len(animation.Bindings)),
	}
	for i, binding := range animation.Bindings {
		translationKeyframes := make([]newasset.AnimationKeyframe[dprec.Vec3], len(binding.TranslationKeyframes))
		for j, keyframe := range binding.TranslationKeyframes {
			translationKeyframes[j] = newasset.AnimationKeyframe[dprec.Vec3]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Translation,
			}
		}
		rotationKeyframes := make([]newasset.AnimationKeyframe[dprec.Quat], len(binding.RotationKeyframes))
		for j, keyframe := range binding.RotationKeyframes {
			rotationKeyframes[j] = newasset.AnimationKeyframe[dprec.Quat]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Rotation,
			}
		}
		scaleKeyframes := make([]newasset.AnimationKeyframe[dprec.Vec3], len(binding.ScaleKeyframes))
		for j, keyframe := range binding.ScaleKeyframes {
			scaleKeyframes[j] = newasset.AnimationKeyframe[dprec.Vec3]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Scale,
			}
		}
		assetAnimation.Bindings[i] = newasset.AnimationBinding{
			NodeName:             binding.Node.Name,
			TranslationKeyframes: translationKeyframes,
			RotationKeyframes:    rotationKeyframes,
			ScaleKeyframes:       scaleKeyframes,
		}
	}
	return assetAnimation
}

func (c *converter) BuildGeometryShader(material *Material) newasset.Shader {
	var sourceCode string

	var textureLines string
	if ref := material.ColorTexture; ref != nil {
		textureLines += "  baseColorSampler sampler2D,\n"
	}
	if ref := material.MetallicRoughnessTexture; ref != nil {
		textureLines += "  metallicRoughnessSampler sampler2D,\n"
	}
	if ref := material.NormalTexture; ref != nil {
		textureLines += "  normalSampler sampler2D,\n"
	}
	if textureLines != "" {
		sourceCode += "textures {\n" + textureLines + "}\n"
	}

	sourceCode += `
		uniforms {
			baseColor vec4,
			metallic float,
			roughness float,
			normalScale float,
		}
	`
	sourceCode += `
		func #fragment() {
	`

	if ref := material.ColorTexture; ref != nil {
		sourceCode += `
			#color = sample(baseColorSampler, #uv)
		`
	} else {
		sourceCode += `
			#color = baseColor
		`
	}

	sourceCode += `
			#metallic = metallic
			#roughness = roughness
		}
	`
	// TODO
	// 	AlphaTesting:   material.AlphaTesting,
	// 	AlphaThreshold: material.AlphaThreshold,
	// 	Blending:       material.Blending,

	return newasset.Shader{
		SourceCode: sourceCode,
	}
}

func (c *converter) BuildShadowShader(material *Material) newasset.Shader {
	return newasset.Shader{
		SourceCode: `
			// Use default.
		`,
	}
}

func (c *converter) BuildMaterial(material *Material) newasset.Material {
	geometryShaderIndex := uint32(c.assetGeometryShaderIndexFromMaterial[material])
	shadowShaderIndex := uint32(c.assetShadowShaderIndexFromMaterial[material])

	var culling = newasset.CullModeNone
	if material.BackfaceCulling {
		culling = newasset.CullModeBack
	}

	var textures []newasset.TextureBinding
	if ref := material.ColorTexture; ref != nil {
		textures = append(textures, newasset.TextureBinding{
			BindingName:  "baseColorSampler",
			TextureIndex: uint32(ref.TextureIndex),
			Wrapping:     newasset.WrapModeClamp,
			Filtering:    newasset.FilterModeLinear,
			Mipmapping:   true,
		})
	}
	if ref := material.MetallicRoughnessTexture; ref != nil {
		textures = append(textures, newasset.TextureBinding{
			BindingName:  "metallicRoughnessSampler",
			TextureIndex: uint32(ref.TextureIndex),
			Wrapping:     newasset.WrapModeClamp,
			Filtering:    newasset.FilterModeLinear,
			Mipmapping:   true,
		})
	}
	if ref := material.NormalTexture; ref != nil {
		textures = append(textures, newasset.TextureBinding{
			BindingName:  "normalSampler",
			TextureIndex: uint32(ref.TextureIndex),
			Wrapping:     newasset.WrapModeClamp,
			Filtering:    newasset.FilterModeLinear,
			Mipmapping:   true,
		})
	}

	return newasset.Material{
		Name:     material.Name,
		Textures: textures,
		Properties: []newasset.PropertyBinding{
			c.convertProperty("baseColor", material.Color),
			c.convertProperty("metallic", material.Metallic),
			c.convertProperty("roughness", material.Roughness),
			c.convertProperty("normalScale", material.NormalScale),
		},
		GeometryPasses: []newasset.GeometryPass{
			{
				Layer:           0,
				Culling:         culling,
				FrontFace:       newasset.FaceOrientationCCW,
				DepthTest:       true,
				DepthWrite:      true,
				DepthComparison: newasset.ComparisonLess,
				ShaderIndex:     geometryShaderIndex,
			},
		},
		ShadowPasses: []newasset.ShadowPass{
			{
				Layer:       0,
				Culling:     culling,
				FrontFace:   newasset.FaceOrientationCCW,
				ShaderIndex: shadowShaderIndex,
			},
		},
		ForwardPasses: []newasset.ForwardPass{},
	}

}

func (c *converter) convertProperty(name string, value any) newasset.PropertyBinding {
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
		panic(fmt.Errorf("unsupported property type %T", value))
	}
	return newasset.PropertyBinding{
		BindingName: name,
		Data:        data,
	}
}

func (c *converter) BuildArmature(armature *Armature) newasset.Armature {
	return newasset.Armature{
		Joints: gog.Map(armature.Joints, func(joint Joint) newasset.Joint {
			return newasset.Joint{
				NodeIndex:         uint32(c.assetNodeIndexFromNode[joint.Node]),
				InverseBindMatrix: joint.InverseBindMatrix,
			}
		}),
	}
}

func (c *converter) BuildGeometry(meshDefinition *MeshDefinition) newasset.Geometry {
	const (
		sizeUnsignedByte  = 1
		sizeUnsignedShort = 2
		sizeUnsignedInt   = 4
		sizeHalfFloat     = 2
		sizeFloat         = 4
	)

	var (
		stride              uint32
		coordBufferIndex    int32
		coordOffset         uint32
		normalBufferIndex   int32
		normalOffset        uint32
		tangentBufferIndex  int32
		tangentOffset       uint32
		texCoordBufferIndex int32
		texCoordOffset      uint32
		colorBufferIndex    int32
		colorOffset         uint32
		weightsBufferIndex  int32
		weightsOffset       uint32
		jointsBufferIndex   int32
		jointsOffset        uint32
	)

	layout := meshDefinition.VertexLayout
	if layout.HasCoords {
		coordBufferIndex = 0
		coordOffset = stride
		stride += 3 * sizeFloat
	} else {
		coordBufferIndex = newasset.UnspecifiedBufferIndex
	}
	if layout.HasNormals {
		normalBufferIndex = 0
		normalOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		normalBufferIndex = newasset.UnspecifiedBufferIndex
	}
	if layout.HasTangents {
		tangentBufferIndex = 0
		tangentOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		tangentBufferIndex = newasset.UnspecifiedBufferIndex
	}
	if layout.HasTexCoords {
		texCoordBufferIndex = 0
		texCoordOffset = stride
		stride += 2 * sizeHalfFloat
	} else {
		texCoordBufferIndex = newasset.UnspecifiedBufferIndex
	}
	if layout.HasColors {
		colorBufferIndex = 0
		colorOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		colorBufferIndex = newasset.UnspecifiedBufferIndex
	}
	if layout.HasWeights {
		weightsBufferIndex = 0
		weightsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		weightsBufferIndex = newasset.UnspecifiedBufferIndex
	}
	if layout.HasJoints {
		jointsBufferIndex = 0
		jointsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		jointsBufferIndex = newasset.UnspecifiedBufferIndex
	}

	var (
		vertexData = gblob.LittleEndianBlock(make([]byte, len(meshDefinition.Vertices)*int(stride)))
	)
	if layout.HasCoords {
		offset := int(coordOffset)
		for _, vertex := range meshDefinition.Vertices {
			vertexData.SetFloat32(offset+0*sizeFloat, vertex.Coord.X)
			vertexData.SetFloat32(offset+1*sizeFloat, vertex.Coord.Y)
			vertexData.SetFloat32(offset+2*sizeFloat, vertex.Coord.Z)
			offset += int(stride)
		}
	}
	if layout.HasNormals {
		offset := int(normalOffset)
		for _, vertex := range meshDefinition.Vertices {
			vertexData.SetUint16(offset+0*sizeHalfFloat, float16.Fromfloat32(vertex.Normal.X).Bits())
			vertexData.SetUint16(offset+1*sizeHalfFloat, float16.Fromfloat32(vertex.Normal.Y).Bits())
			vertexData.SetUint16(offset+2*sizeHalfFloat, float16.Fromfloat32(vertex.Normal.Z).Bits())
			offset += int(stride)
		}
	}
	if layout.HasTangents {
		offset := int(tangentOffset)
		for _, vertex := range meshDefinition.Vertices {
			vertexData.SetUint16(offset+0*sizeHalfFloat, float16.Fromfloat32(vertex.Tangent.X).Bits())
			vertexData.SetUint16(offset+1*sizeHalfFloat, float16.Fromfloat32(vertex.Tangent.Y).Bits())
			vertexData.SetUint16(offset+2*sizeHalfFloat, float16.Fromfloat32(vertex.Tangent.Z).Bits())
			offset += int(stride)
		}
	}
	if layout.HasTexCoords {
		offset := int(texCoordOffset)
		for _, vertex := range meshDefinition.Vertices {
			vertexData.SetUint16(offset+0*sizeHalfFloat, float16.Fromfloat32(vertex.TexCoord.X).Bits())
			vertexData.SetUint16(offset+1*sizeHalfFloat, float16.Fromfloat32(vertex.TexCoord.Y).Bits())
			offset += int(stride)
		}
	}
	if layout.HasColors {
		offset := int(colorOffset)
		for _, vertex := range meshDefinition.Vertices {
			vertexData.SetUint8(offset+0*sizeUnsignedByte, uint8(vertex.Color.X*255.0))
			vertexData.SetUint8(offset+1*sizeUnsignedByte, uint8(vertex.Color.Y*255.0))
			vertexData.SetUint8(offset+2*sizeUnsignedByte, uint8(vertex.Color.Z*255.0))
			vertexData.SetUint8(offset+3*sizeUnsignedByte, uint8(vertex.Color.W*255.0))
			offset += int(stride)
		}
	}
	if layout.HasWeights {
		offset := int(weightsOffset)
		for _, vertex := range meshDefinition.Vertices {
			vertexData.SetUint8(offset+0*sizeUnsignedByte, uint8(vertex.Weights.X*255.0))
			vertexData.SetUint8(offset+1*sizeUnsignedByte, uint8(vertex.Weights.Y*255.0))
			vertexData.SetUint8(offset+2*sizeUnsignedByte, uint8(vertex.Weights.Z*255.0))
			vertexData.SetUint8(offset+3*sizeUnsignedByte, uint8(vertex.Weights.W*255.0))
			offset += int(stride)
		}
	}
	if layout.HasJoints {
		offset := int(jointsOffset)
		for _, vertex := range meshDefinition.Vertices {
			vertexData.SetUint8(offset+0*sizeUnsignedByte, uint8(vertex.Joints[0]))
			vertexData.SetUint8(offset+1*sizeUnsignedByte, uint8(vertex.Joints[1]))
			vertexData.SetUint8(offset+2*sizeUnsignedByte, uint8(vertex.Joints[2]))
			vertexData.SetUint8(offset+3*sizeUnsignedByte, uint8(vertex.Joints[3]))
			offset += int(stride)
		}
	}

	var (
		indexLayout newasset.IndexLayout
		indexData   gblob.LittleEndianBlock
		indexSize   int
	)
	if len(meshDefinition.Vertices) >= 0xFFFF {
		indexSize = sizeUnsignedInt
		indexLayout = newasset.IndexLayoutUint32
		indexData = gblob.LittleEndianBlock(make([]byte, len(meshDefinition.Indices)*sizeUnsignedInt))
		for i, index := range meshDefinition.Indices {
			indexData.SetUint32(i*sizeUnsignedInt, uint32(index))
		}
	} else {
		indexSize = sizeUnsignedShort
		indexLayout = newasset.IndexLayoutUint16
		indexData = gblob.LittleEndianBlock(make([]byte, len(meshDefinition.Indices)*sizeUnsignedShort))
		for i, index := range meshDefinition.Indices {
			indexData.SetUint16(i*sizeUnsignedShort, uint16(index))
		}
	}

	var (
		fragments = make([]newasset.Fragment, 0, len(meshDefinition.Fragments))
	)
	for _, fragment := range meshDefinition.Fragments {
		if fragment.Material != nil && !fragment.Material.IsInvisible() {
			fragments = append(fragments, c.BuildFragment(fragment, indexSize))
		}
	}

	var boundingSphereRadius float64
	for _, vertex := range meshDefinition.Vertices {
		boundingSphereRadius = dprec.Max(
			boundingSphereRadius,
			float64(vertex.Coord.Length()),
		)
	}

	return newasset.Geometry{
		VertexBuffers: []newasset.VertexBuffer{
			{
				Stride: stride,
				Data:   vertexData,
			},
		},
		VertexLayout: newasset.VertexLayout{
			Coord: newasset.VertexAttribute{
				BufferIndex: coordBufferIndex,
				ByteOffset:  coordOffset,
				Format:      newasset.VertexAttributeFormatRGB32F,
			},
			Normal: newasset.VertexAttribute{
				BufferIndex: normalBufferIndex,
				ByteOffset:  normalOffset,
				Format:      newasset.VertexAttributeFormatRGB16F,
			},
			Tangent: newasset.VertexAttribute{
				BufferIndex: tangentBufferIndex,
				ByteOffset:  tangentOffset,
				Format:      newasset.VertexAttributeFormatRGB16F,
			},
			TexCoord: newasset.VertexAttribute{
				BufferIndex: texCoordBufferIndex,
				ByteOffset:  texCoordOffset,
				Format:      newasset.VertexAttributeFormatRG16F,
			},
			Color: newasset.VertexAttribute{
				BufferIndex: colorBufferIndex,
				ByteOffset:  colorOffset,
				Format:      newasset.VertexAttributeFormatRGBA8UN,
			},
			Weights: newasset.VertexAttribute{
				BufferIndex: weightsBufferIndex,
				ByteOffset:  weightsOffset,
				Format:      newasset.VertexAttributeFormatRGBA8UN,
			},
			Joints: newasset.VertexAttribute{
				BufferIndex: jointsBufferIndex,
				ByteOffset:  jointsOffset,
				Format:      newasset.VertexAttributeFormatRGBA8IU,
			},
		},
		IndexBuffer: newasset.IndexBuffer{
			IndexLayout: indexLayout,
			Data:        indexData,
		},
		Fragments:            fragments,
		BoundingSphereRadius: boundingSphereRadius,
	}
}

func (c *converter) BuildMeshDefinition(meshDefinition *MeshDefinition) newasset.MeshDefinition {
	var materialBindings []newasset.MaterialBinding
	for _, fragment := range meshDefinition.Fragments {
		if fragment.Material != nil && !fragment.Material.IsInvisible() {
			materialBindings = append(materialBindings, newasset.MaterialBinding{
				MaterialIndex: uint32(c.assetMaterialIndexFromMaterial[fragment.Material]),
			})
		}
	}

	return newasset.MeshDefinition{
		GeometryIndex:    uint32(c.assetGeometriesFromMeshDefinition[meshDefinition]),
		MaterialBindings: materialBindings,
	}
}

func (c *converter) BuildFragment(fragment MeshFragment, indexSize int) newasset.Fragment {
	var topology newasset.Topology
	switch fragment.Primitive {
	case PrimitivePoints:
		topology = newasset.TopologyPoints
	case PrimitiveLines:
		topology = newasset.TopologyLineList
	case PrimitiveLineStrip:
		topology = newasset.TopologyLineStrip
	case PrimitiveTriangles:
		topology = newasset.TopologyTriangleList
	case PrimitiveTriangleStrip:
		topology = newasset.TopologyTriangleStrip
	default:
		panic(fmt.Errorf("unsupported primitive type: %d", fragment.Primitive))
	}

	return newasset.Fragment{
		Name:            "", // TODO: Take from material
		Topology:        topology,
		IndexByteOffset: uint32(fragment.IndexOffset * indexSize),
		IndexCount:      uint32(fragment.IndexCount),
	}
}

func (c *converter) BuildMeshInstance(meshInstance *MeshInstance) newasset.Mesh {
	var nodeIndex int32
	if index, ok := c.assetNodeIndexFromNode[meshInstance.Node]; ok {
		nodeIndex = int32(index)
	} else {
		panic(fmt.Errorf("node %s not found", meshInstance.Node.Name))
	}
	var definitionIndex int32
	if index, ok := c.assetMeshDefinitionFromMeshDefinition[meshInstance.Definition]; ok {
		definitionIndex = int32(index)
	} else {
		panic(fmt.Errorf("mesh definition %s not found", meshInstance.Definition.Name))
	}
	var armatureIndex int32 = newasset.UnspecifiedArmatureIndex
	if meshInstance.Armature != nil {
		if index, ok := c.assetArmatureIndexFromArmature[meshInstance.Armature]; ok {
			armatureIndex = int32(index)
		} else {
			panic(fmt.Errorf("armature not found"))
		}
	}
	return newasset.Mesh{
		NodeIndex:           uint32(nodeIndex),
		ArmatureIndex:       armatureIndex,
		MeshDefinitionIndex: uint32(definitionIndex),
	}
}

func (c *converter) BuildBodyMaterial(meshDefinition *MeshDefinition) newasset.BodyMaterial {
	return newasset.BodyMaterial{
		FrictionCoefficient:    0.9,
		RestitutionCoefficient: 0.5,
	}
}

func (c *converter) BuildBodyDefinition(meshDefinition *MeshDefinition) newasset.BodyDefinition {
	var triangles []newasset.CollisionTriangle

	for _, fragment := range meshDefinition.Fragments {
		if fragment.Material != nil && fragment.Material.HasSkipCollision() {
			continue
		}
		if fragment.Primitive != PrimitiveTriangles {
			logger.Warn("Skipping collision mesh due to primitive not being triangles!")
			continue
		}
		for i := fragment.IndexOffset; i < fragment.IndexOffset+fragment.IndexCount; i += 3 {
			indexA := meshDefinition.Indices[i+0]
			indexB := meshDefinition.Indices[i+1]
			indexC := meshDefinition.Indices[i+2]

			coordA := meshDefinition.Vertices[indexA].Coord
			coordB := meshDefinition.Vertices[indexB].Coord
			coordC := meshDefinition.Vertices[indexC].Coord

			vecAB := sprec.Vec3Diff(coordB, coordA)
			vecAC := sprec.Vec3Diff(coordC, coordA)
			if sprec.Vec3Cross(vecAB, vecAC).Length() < 0.00001 {
				logger.Warn("Degenerate triangle omitted!")
				continue
			}

			triangles = append(triangles, newasset.CollisionTriangle{
				A: stod.Vec3(coordA),
				B: stod.Vec3(coordB),
				C: stod.Vec3(coordC),
			})
		}
	}

	// TODO: Dynamic grid size based on density
	const gridSize = 10

	type cell struct {
		X int
		Y int
		Z int
	}

	cells := gog.Partition(triangles, func(triangle newasset.CollisionTriangle) cell {
		centroid := dprec.Vec3Quot(dprec.Vec3Sum(dprec.Vec3Sum(triangle.A, triangle.B), triangle.C), 3.0)
		return cell{
			X: int(centroid.X) / gridSize,
			Y: int(centroid.Y) / gridSize,
			Z: int(centroid.Z) / gridSize,
		}
	})

	meshes := gog.Map(gog.Entries(cells), func(pair gog.KV[cell, []newasset.CollisionTriangle]) newasset.CollisionMesh {
		triangles := pair.Value

		center := dprec.Vec3Quot(gog.Reduce(triangles, dprec.ZeroVec3(), func(accum dprec.Vec3, triangle newasset.CollisionTriangle) dprec.Vec3 {
			return dprec.Vec3Sum(triangle.C, dprec.Vec3Sum(triangle.B, dprec.Vec3Sum(triangle.A, accum)))
		}), 3*float64(len(triangles)))

		triangles = gog.Map(triangles, func(triangle newasset.CollisionTriangle) newasset.CollisionTriangle {
			return newasset.CollisionTriangle{
				A: dprec.Vec3Diff(triangle.A, center),
				B: dprec.Vec3Diff(triangle.B, center),
				C: dprec.Vec3Diff(triangle.C, center),
			}
		})

		return newasset.CollisionMesh{
			Translation: center,
			Rotation:    dprec.IdentityQuat(),
			Triangles:   triangles,
		}
	})

	return newasset.BodyDefinition{
		MaterialIndex:   uint32(c.assetBodyMaterialFromMeshDefinition[meshDefinition]),
		CollisionMeshes: meshes,
	}
}

func (c *converter) BuildBodyInstance(meshInstance *MeshInstance) newasset.Body {
	var nodeIndex uint32
	if index, ok := c.assetNodeIndexFromNode[meshInstance.Node]; ok {
		nodeIndex = uint32(index)
	} else {
		panic(fmt.Errorf("node %s not found", meshInstance.Node.Name))
	}
	var definitionIndex uint32
	if index, ok := c.assetBodyDefinitionFromMeshDefinition[meshInstance.Definition]; ok {
		definitionIndex = uint32(index)
	} else {
		panic(fmt.Errorf("body definition %s not found", meshInstance.Definition.Name))
	}
	return newasset.Body{
		NodeIndex:           nodeIndex,
		BodyDefinitionIndex: definitionIndex,
	}
}
