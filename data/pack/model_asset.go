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
)

type SaveModelAssetOption func(a *SaveModelAssetAction)

func WithCollisionMesh(collisionMesh bool) SaveModelAssetOption {
	return func(a *SaveModelAssetAction) {
		a.forceCollidable = collisionMesh
	}
}

type SaveModelAssetAction struct {
	resource        *asset.Resource
	modelProvider   ModelProvider
	forceCollidable bool
}

func (a *SaveModelAssetAction) Describe() string {
	return fmt.Sprintf("save_model_asset(%q)", a.resource.Name())
}

func (a *SaveModelAssetAction) Run() error {
	conv := newConverter(a.forceCollidable)
	modelAsset := conv.BuildModel(a.modelProvider.Model())
	if err := a.resource.SaveContent(modelAsset); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	return nil
}

func newConverter(collisionMeshes bool) *converter {
	return &converter{
		forceCollidable:                       collisionMeshes,
		assetNodes:                            make([]asset.Node, 0),
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
	assetNodes                            []asset.Node
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

func (c *converter) BuildModel(model *Model) asset.Model {
	for _, node := range model.RootNodes {
		c.BuildNode(-1, node)
	}

	assetGeometryShaders := make([]asset.Shader, len(model.Materials))
	assetShadowShaders := make([]asset.Shader, len(model.Materials))
	for i, material := range model.Materials {
		assetGeometryShaders[i] = c.BuildGeometryShader(material)
		c.assetGeometryShaderIndexFromMaterial[material] = i

		assetShadowShaders[i] = c.BuildShadowShader(material)
		c.assetShadowShaderIndexFromMaterial[material] = i
	}

	assetTextures := make([]asset.Texture, len(model.Textures))
	for i, texture := range model.Textures {
		assetTextures[i] = BuildTwoDTextureAsset(texture)
	}

	assetAnimations := make([]asset.Animation, len(model.Animations))
	for i, animation := range model.Animations {
		assetAnimations[i] = c.BuildAnimation(animation)
	}

	assetMaterials := make([]asset.Material, len(model.Materials))
	for i, material := range model.Materials {
		assetMaterials[i] = c.BuildMaterial(material)
		c.assetMaterialIndexFromMaterial[material] = i
	}

	assetArmatures := make([]asset.Armature, len(model.Armatures))
	for i, armature := range model.Armatures {
		assetArmatures[i] = c.BuildArmature(armature)
		c.assetArmatureIndexFromArmature[armature] = i
	}

	assetGeometries := make([]asset.Geometry, len(model.MeshDefinitions))
	for i, meshDefinition := range model.MeshDefinitions {
		assetGeometries[i] = c.BuildGeometry(meshDefinition)
		c.assetGeometriesFromMeshDefinition[meshDefinition] = i
	}

	assetMeshDefinitions := make([]asset.MeshDefinition, len(model.MeshDefinitions))
	for i, meshDefinition := range model.MeshDefinitions {
		assetMeshDefinitions[i] = c.BuildMeshDefinition(meshDefinition)
		c.assetMeshDefinitionFromMeshDefinition[meshDefinition] = i
	}

	assetMeshInstances := make([]asset.Mesh, len(model.MeshInstances))
	for i, meshInstance := range model.MeshInstances {
		assetMeshInstances[i] = c.BuildMeshInstance(meshInstance)
	}

	assetBodyMaterials := make([]asset.BodyMaterial, len(model.MeshDefinitions))
	for i, meshDefinition := range model.MeshDefinitions {
		assetBodyMaterials[i] = c.BuildBodyMaterial(meshDefinition)
		c.assetBodyMaterialFromMeshDefinition[meshDefinition] = i
	}

	assetBodyDefinitions := make([]asset.BodyDefinition, len(model.MeshDefinitions))
	for i, meshDefinition := range model.MeshDefinitions {
		assetBodyDefinitions[i] = c.BuildBodyDefinition(meshDefinition)
		c.assetBodyDefinitionFromMeshDefinition[meshDefinition] = i
	}

	assetBodyInstances := make([]asset.Body, 0, len(model.MeshInstances))
	for _, meshInstance := range model.MeshInstances {
		if c.forceCollidable || meshInstance.HasCollision() {
			assetBodyInstances = append(assetBodyInstances, c.BuildBodyInstance(meshInstance))
		}
	}

	assetPointLights := make([]asset.PointLight, 0)
	assetSpotLights := make([]asset.SpotLight, 0)
	assetDirectionalLights := make([]asset.DirectionalLight, 0)
	for _, lightInstance := range model.LightInstances {
		lightDefinition := lightInstance.Definition
		nodeIndex, ok := c.assetNodeIndexFromNode[lightInstance.Node]
		if !ok {
			panic(fmt.Errorf("node %s not found", lightInstance.Node.Name))
		}
		switch lightDefinition.Type {
		case LightTypePoint:
			assetPointLights = append(assetPointLights, asset.PointLight{
				NodeIndex:    uint32(nodeIndex),
				EmitColor:    lightDefinition.EmitColor,
				EmitDistance: lightDefinition.EmitRange,
				CastShadow:   false,
			})
		case LightTypeSpot:
			assetSpotLights = append(assetSpotLights, asset.SpotLight{
				NodeIndex:      uint32(nodeIndex),
				EmitColor:      lightDefinition.EmitColor,
				EmitDistance:   lightDefinition.EmitRange,
				EmitAngleOuter: lightDefinition.EmitOuterConeAngle,
				EmitAngleInner: lightDefinition.EmitInnerConeAngle,
				CastShadow:     false,
			})
		case LightTypeDirectional:
			assetDirectionalLights = append(assetDirectionalLights, asset.DirectionalLight{
				NodeIndex:  uint32(nodeIndex),
				EmitColor:  lightDefinition.EmitColor,
				CastShadow: false,
			})
		default:
			panic(fmt.Errorf("unknown light type %q", lightInstance.Definition.Type))
		}
	}

	return asset.Model{
		Nodes:             c.assetNodes,
		GeometryShaders:   assetGeometryShaders,
		ShadowShaders:     assetShadowShaders,
		Animations:        assetAnimations,
		Armatures:         assetArmatures,
		Textures:          assetTextures,
		Materials:         assetMaterials,
		Geometries:        assetGeometries,
		MeshDefinitions:   assetMeshDefinitions,
		Meshes:            assetMeshInstances,
		BodyMaterials:     assetBodyMaterials,
		BodyDefinitions:   assetBodyDefinitions,
		Bodies:            assetBodyInstances,
		PointLights:       assetPointLights,
		SpotLights:        assetSpotLights,
		DirectionalLights: assetDirectionalLights,
	}
}

func (c *converter) BuildNode(parentIndex int, node *Node) {
	result := asset.Node{
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

func (c *converter) BuildAnimation(animation *Animation) asset.Animation {
	assetAnimation := asset.Animation{
		Name:      animation.Name,
		StartTime: animation.StartTime,
		EndTime:   animation.EndTime,
		Bindings:  make([]asset.AnimationBinding, len(animation.Bindings)),
	}
	for i, binding := range animation.Bindings {
		translationKeyframes := make([]asset.AnimationKeyframe[dprec.Vec3], len(binding.TranslationKeyframes))
		for j, keyframe := range binding.TranslationKeyframes {
			translationKeyframes[j] = asset.AnimationKeyframe[dprec.Vec3]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Translation,
			}
		}
		rotationKeyframes := make([]asset.AnimationKeyframe[dprec.Quat], len(binding.RotationKeyframes))
		for j, keyframe := range binding.RotationKeyframes {
			rotationKeyframes[j] = asset.AnimationKeyframe[dprec.Quat]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Rotation,
			}
		}
		scaleKeyframes := make([]asset.AnimationKeyframe[dprec.Vec3], len(binding.ScaleKeyframes))
		for j, keyframe := range binding.ScaleKeyframes {
			scaleKeyframes[j] = asset.AnimationKeyframe[dprec.Vec3]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Scale,
			}
		}
		assetAnimation.Bindings[i] = asset.AnimationBinding{
			NodeName:             binding.Node.Name,
			TranslationKeyframes: translationKeyframes,
			RotationKeyframes:    rotationKeyframes,
			ScaleKeyframes:       scaleKeyframes,
		}
	}
	return assetAnimation
}

func (c *converter) BuildGeometryShader(material *Material) asset.Shader {
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
			alphaThreshold float,
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

	if material.AlphaTesting {
		sourceCode += `
				if #color.a < alphaThreshold {
					discard
				}
		`
	}

	sourceCode += `
			#metallic = metallic
			#roughness = roughness
		}
	`

	return asset.Shader{
		SourceCode: sourceCode,
	}
}

func (c *converter) BuildShadowShader(material *Material) asset.Shader {
	return asset.Shader{
		SourceCode: `
			// Use default.
		`,
	}
}

func (c *converter) BuildMaterial(material *Material) asset.Material {
	geometryShaderIndex := uint32(c.assetGeometryShaderIndexFromMaterial[material])
	shadowShaderIndex := uint32(c.assetShadowShaderIndexFromMaterial[material])

	var culling = asset.CullModeNone
	if material.BackfaceCulling {
		culling = asset.CullModeBack
	}

	var textures []asset.TextureBinding
	if ref := material.ColorTexture; ref != nil {
		textures = append(textures, asset.TextureBinding{
			BindingName:  "baseColorSampler",
			TextureIndex: uint32(ref.TextureIndex),
			Wrapping:     asset.WrapModeClamp,
			Filtering:    asset.FilterModeLinear,
			Mipmapping:   true,
		})
	}
	if ref := material.MetallicRoughnessTexture; ref != nil {
		textures = append(textures, asset.TextureBinding{
			BindingName:  "metallicRoughnessSampler",
			TextureIndex: uint32(ref.TextureIndex),
			Wrapping:     asset.WrapModeClamp,
			Filtering:    asset.FilterModeLinear,
			Mipmapping:   true,
		})
	}
	if ref := material.NormalTexture; ref != nil {
		textures = append(textures, asset.TextureBinding{
			BindingName:  "normalSampler",
			TextureIndex: uint32(ref.TextureIndex),
			Wrapping:     asset.WrapModeClamp,
			Filtering:    asset.FilterModeLinear,
			Mipmapping:   true,
		})
	}

	return asset.Material{
		Name:     material.Name,
		Textures: textures,
		Properties: []asset.PropertyBinding{
			c.convertProperty("baseColor", material.Color),
			c.convertProperty("metallic", material.Metallic),
			c.convertProperty("roughness", material.Roughness),
			c.convertProperty("normalScale", material.NormalScale),
			c.convertProperty("alphaThreshold", material.AlphaThreshold),
		},
		GeometryPasses: []asset.GeometryPass{
			{
				Layer:           0,
				Culling:         culling,
				FrontFace:       asset.FaceOrientationCCW,
				DepthTest:       true,
				DepthWrite:      true,
				DepthComparison: asset.ComparisonLess,
				ShaderIndex:     geometryShaderIndex,
			},
		},
		ShadowPasses: []asset.ShadowPass{
			{
				Layer:       0,
				Culling:     culling,
				FrontFace:   asset.FaceOrientationCCW,
				ShaderIndex: shadowShaderIndex,
			},
		},
		ForwardPasses: []asset.ForwardPass{},
	}

}

func (c *converter) convertProperty(name string, value any) asset.PropertyBinding {
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
	return asset.PropertyBinding{
		BindingName: name,
		Data:        data,
	}
}

func (c *converter) BuildArmature(armature *Armature) asset.Armature {
	return asset.Armature{
		Joints: gog.Map(armature.Joints, func(joint Joint) asset.Joint {
			return asset.Joint{
				NodeIndex:         uint32(c.assetNodeIndexFromNode[joint.Node]),
				InverseBindMatrix: joint.InverseBindMatrix,
			}
		}),
	}
}

func (c *converter) BuildGeometry(meshDefinition *MeshDefinition) asset.Geometry {
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
		coordBufferIndex = asset.UnspecifiedBufferIndex
	}
	if layout.HasNormals {
		normalBufferIndex = 0
		normalOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		normalBufferIndex = asset.UnspecifiedBufferIndex
	}
	if layout.HasTangents {
		tangentBufferIndex = 0
		tangentOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		tangentBufferIndex = asset.UnspecifiedBufferIndex
	}
	if layout.HasTexCoords {
		texCoordBufferIndex = 0
		texCoordOffset = stride
		stride += 2 * sizeHalfFloat
	} else {
		texCoordBufferIndex = asset.UnspecifiedBufferIndex
	}
	if layout.HasColors {
		colorBufferIndex = 0
		colorOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		colorBufferIndex = asset.UnspecifiedBufferIndex
	}
	if layout.HasWeights {
		weightsBufferIndex = 0
		weightsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		weightsBufferIndex = asset.UnspecifiedBufferIndex
	}
	if layout.HasJoints {
		jointsBufferIndex = 0
		jointsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		jointsBufferIndex = asset.UnspecifiedBufferIndex
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
		indexLayout asset.IndexLayout
		indexData   gblob.LittleEndianBlock
		indexSize   int
	)
	if len(meshDefinition.Vertices) >= 0xFFFF {
		indexSize = sizeUnsignedInt
		indexLayout = asset.IndexLayoutUint32
		indexData = gblob.LittleEndianBlock(make([]byte, len(meshDefinition.Indices)*sizeUnsignedInt))
		for i, index := range meshDefinition.Indices {
			indexData.SetUint32(i*sizeUnsignedInt, uint32(index))
		}
	} else {
		indexSize = sizeUnsignedShort
		indexLayout = asset.IndexLayoutUint16
		indexData = gblob.LittleEndianBlock(make([]byte, len(meshDefinition.Indices)*sizeUnsignedShort))
		for i, index := range meshDefinition.Indices {
			indexData.SetUint16(i*sizeUnsignedShort, uint16(index))
		}
	}

	var (
		fragments = make([]asset.Fragment, 0, len(meshDefinition.Fragments))
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

	return asset.Geometry{
		VertexBuffers: []asset.VertexBuffer{
			{
				Stride: stride,
				Data:   vertexData,
			},
		},
		VertexLayout: asset.VertexLayout{
			Coord: asset.VertexAttribute{
				BufferIndex: coordBufferIndex,
				ByteOffset:  coordOffset,
				Format:      asset.VertexAttributeFormatRGB32F,
			},
			Normal: asset.VertexAttribute{
				BufferIndex: normalBufferIndex,
				ByteOffset:  normalOffset,
				Format:      asset.VertexAttributeFormatRGB16F,
			},
			Tangent: asset.VertexAttribute{
				BufferIndex: tangentBufferIndex,
				ByteOffset:  tangentOffset,
				Format:      asset.VertexAttributeFormatRGB16F,
			},
			TexCoord: asset.VertexAttribute{
				BufferIndex: texCoordBufferIndex,
				ByteOffset:  texCoordOffset,
				Format:      asset.VertexAttributeFormatRG16F,
			},
			Color: asset.VertexAttribute{
				BufferIndex: colorBufferIndex,
				ByteOffset:  colorOffset,
				Format:      asset.VertexAttributeFormatRGBA8UN,
			},
			Weights: asset.VertexAttribute{
				BufferIndex: weightsBufferIndex,
				ByteOffset:  weightsOffset,
				Format:      asset.VertexAttributeFormatRGBA8UN,
			},
			Joints: asset.VertexAttribute{
				BufferIndex: jointsBufferIndex,
				ByteOffset:  jointsOffset,
				Format:      asset.VertexAttributeFormatRGBA8IU,
			},
		},
		IndexBuffer: asset.IndexBuffer{
			IndexLayout: indexLayout,
			Data:        indexData,
		},
		Fragments:            fragments,
		BoundingSphereRadius: boundingSphereRadius,
	}
}

func (c *converter) BuildMeshDefinition(meshDefinition *MeshDefinition) asset.MeshDefinition {
	var materialBindings []asset.MaterialBinding
	for _, fragment := range meshDefinition.Fragments {
		if fragment.Material != nil && !fragment.Material.IsInvisible() {
			materialBindings = append(materialBindings, asset.MaterialBinding{
				MaterialIndex: uint32(c.assetMaterialIndexFromMaterial[fragment.Material]),
			})
		}
	}

	return asset.MeshDefinition{
		GeometryIndex:    uint32(c.assetGeometriesFromMeshDefinition[meshDefinition]),
		MaterialBindings: materialBindings,
	}
}

func (c *converter) BuildFragment(fragment MeshFragment, indexSize int) asset.Fragment {
	var topology asset.Topology
	switch fragment.Primitive {
	case PrimitivePoints:
		topology = asset.TopologyPoints
	case PrimitiveLines:
		topology = asset.TopologyLineList
	case PrimitiveLineStrip:
		topology = asset.TopologyLineStrip
	case PrimitiveTriangles:
		topology = asset.TopologyTriangleList
	case PrimitiveTriangleStrip:
		topology = asset.TopologyTriangleStrip
	default:
		panic(fmt.Errorf("unsupported primitive type: %d", fragment.Primitive))
	}

	return asset.Fragment{
		Name:            "", // TODO: Take from material
		Topology:        topology,
		IndexByteOffset: uint32(fragment.IndexOffset * indexSize),
		IndexCount:      uint32(fragment.IndexCount),
	}
}

func (c *converter) BuildMeshInstance(meshInstance *MeshInstance) asset.Mesh {
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
	var armatureIndex int32 = asset.UnspecifiedArmatureIndex
	if meshInstance.Armature != nil {
		if index, ok := c.assetArmatureIndexFromArmature[meshInstance.Armature]; ok {
			armatureIndex = int32(index)
		} else {
			panic(fmt.Errorf("armature not found"))
		}
	}
	return asset.Mesh{
		NodeIndex:           uint32(nodeIndex),
		ArmatureIndex:       armatureIndex,
		MeshDefinitionIndex: uint32(definitionIndex),
	}
}

func (c *converter) BuildBodyMaterial(meshDefinition *MeshDefinition) asset.BodyMaterial {
	return asset.BodyMaterial{
		FrictionCoefficient:    0.9,
		RestitutionCoefficient: 0.5,
	}
}

func (c *converter) BuildBodyDefinition(meshDefinition *MeshDefinition) asset.BodyDefinition {
	var triangles []asset.CollisionTriangle

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

			triangles = append(triangles, asset.CollisionTriangle{
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

	cells := gog.Partition(triangles, func(triangle asset.CollisionTriangle) cell {
		centroid := dprec.Vec3Quot(dprec.Vec3Sum(dprec.Vec3Sum(triangle.A, triangle.B), triangle.C), 3.0)
		return cell{
			X: int(centroid.X) / gridSize,
			Y: int(centroid.Y) / gridSize,
			Z: int(centroid.Z) / gridSize,
		}
	})

	meshes := gog.Map(gog.Entries(cells), func(pair gog.KV[cell, []asset.CollisionTriangle]) asset.CollisionMesh {
		triangles := pair.Value

		center := dprec.Vec3Quot(gog.Reduce(triangles, dprec.ZeroVec3(), func(accum dprec.Vec3, triangle asset.CollisionTriangle) dprec.Vec3 {
			return dprec.Vec3Sum(triangle.C, dprec.Vec3Sum(triangle.B, dprec.Vec3Sum(triangle.A, accum)))
		}), 3*float64(len(triangles)))

		triangles = gog.Map(triangles, func(triangle asset.CollisionTriangle) asset.CollisionTriangle {
			return asset.CollisionTriangle{
				A: dprec.Vec3Diff(triangle.A, center),
				B: dprec.Vec3Diff(triangle.B, center),
				C: dprec.Vec3Diff(triangle.C, center),
			}
		})

		return asset.CollisionMesh{
			Translation: center,
			Rotation:    dprec.IdentityQuat(),
			Triangles:   triangles,
		}
	})

	return asset.BodyDefinition{
		MaterialIndex:   uint32(c.assetBodyMaterialFromMeshDefinition[meshDefinition]),
		CollisionMeshes: meshes,
	}
}

func (c *converter) BuildBodyInstance(meshInstance *MeshInstance) asset.Body {
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
	return asset.Body{
		NodeIndex:           nodeIndex,
		BodyDefinitionIndex: definitionIndex,
	}
}
