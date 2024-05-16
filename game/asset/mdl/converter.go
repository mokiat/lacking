package mdl

import (
	"fmt"
	"slices"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/asset"
	"github.com/x448/float16"
)

func NewConverter(model *Model) *Converter {
	return &Converter{
		model: model,

		convertedNodes:           make(map[*Node]uint32),
		convertedArmatures:       make(map[*Armature]uint32),
		convertedShaders:         make(map[*Shader]uint32),
		convertedTextures:        make(map[*Texture]uint32),
		convertedMaterials:       make(map[*Material]uint32),
		convertedGeometries:      make(map[*Geometry]uint32),
		convertedMeshDefinitions: make(map[*MeshDefinition]uint32),
		convertedBodyMaterials:   make(map[*BodyMaterial]uint32),
		convertedBodyDefinitions: make(map[*BodyDefinition]uint32),
	}
}

type Converter struct {
	model *Model

	assetNodes     []asset.Node
	convertedNodes map[*Node]uint32

	assetArmatures     []asset.Armature
	convertedArmatures map[*Armature]uint32

	assetShaders     []asset.Shader
	convertedShaders map[*Shader]uint32

	assetTextures     []asset.Texture
	convertedTextures map[*Texture]uint32

	assetMaterials     []asset.Material
	convertedMaterials map[*Material]uint32

	assetGeometries     []asset.Geometry
	convertedGeometries map[*Geometry]uint32

	assetMeshDefinitions     []asset.MeshDefinition
	convertedMeshDefinitions map[*MeshDefinition]uint32

	assetBodyMaterials     []asset.BodyMaterial
	convertedBodyMaterials map[*BodyMaterial]uint32

	assetBodyDefinitions     []asset.BodyDefinition
	convertedBodyDefinitions map[*BodyDefinition]uint32
}

func (c *Converter) Convert() (asset.Model, error) {
	return c.convertModel(c.model)
}

func (c *Converter) convertModel(s *Model) (asset.Model, error) {
	var (
		assetMeshes            []asset.Mesh
		assetBodies            []asset.Body
		assetAmbientLights     []asset.AmbientLight
		assetPointLights       []asset.PointLight
		assetSpotLights        []asset.SpotLight
		assetDirectionalLights []asset.DirectionalLight
		assetSkies             []asset.Sky
	)

	nodes := s.FlattenNodes()

	c.assetNodes = nil

	// First nodes pass, so that all nodes are tracked, otherwise
	// armature resolution will fail.
	for i, node := range nodes {
		c.convertedNodes[node] = uint32(i)
	}

	for i, node := range nodes {
		parentIndex := asset.UnspecifiedNodeIndex
		if pIndex, ok := c.convertedNodes[node.Parent()]; ok {
			parentIndex = int32(pIndex)
		}

		c.assetNodes = append(c.assetNodes, asset.Node{
			Name:        node.Name(),
			ParentIndex: parentIndex,
			Translation: node.Translation(),
			Rotation:    node.Rotation(),
			Scale:       node.Scale(),
			Mask:        asset.NodeMaskNone,
		})

		switch source := node.source.(type) {
		case *Body:
			assetBody, err := c.convertBody(uint32(i), source)
			if err != nil {
				return asset.Model{}, fmt.Errorf("error converting body %q: %w", node.Name(), err)
			}
			assetBodies = append(assetBodies, assetBody)
		}
		switch target := node.target.(type) {
		case *Mesh:
			assetMesh, err := c.convertMesh(uint32(i), target)
			if err != nil {
				return asset.Model{}, fmt.Errorf("error converting mesh %q: %w", node.Name(), err)
			}
			assetMeshes = append(assetMeshes, assetMesh)
		case *AmbientLight:
			ambientLightAsset, err := c.convertAmbientLight(uint32(i), target)
			if err != nil {
				return asset.Model{}, fmt.Errorf("error converting ambient light %q: %w", node.Name(), err)
			}
			assetAmbientLights = append(assetAmbientLights, ambientLightAsset)
		case *PointLight:
			pointLightAsset := c.convertPointLight(uint32(i), target)
			assetPointLights = append(assetPointLights, pointLightAsset)
		case *SpotLight:
			spotLightAsset := c.convertSpotLight(uint32(i), target)
			assetSpotLights = append(assetSpotLights, spotLightAsset)
		case *DirectionalLight:
			directionalLightAsset := c.convertDirectionalLight(uint32(i), target)
			assetDirectionalLights = append(assetDirectionalLights, directionalLightAsset)
		case *Sky:
			assetSky, err := c.convertSky(uint32(i), target)
			if err != nil {
				return asset.Model{}, fmt.Errorf("error converting sky %q: %w", node.Name(), err)
			}
			assetSkies = append(assetSkies, assetSky)
		}
	}

	assetAnimations := make([]asset.Animation, len(c.model.animations))
	for i, animation := range c.model.animations {
		assetAnimations[i] = c.convertAnimation(animation)
	}

	return asset.Model{
		Nodes:             c.assetNodes,
		Animations:        assetAnimations,
		Armatures:         c.assetArmatures,
		Shaders:           c.assetShaders,
		Textures:          c.assetTextures,
		Materials:         c.assetMaterials,
		Geometries:        c.assetGeometries,
		MeshDefinitions:   c.assetMeshDefinitions,
		Meshes:            assetMeshes,
		BodyMaterials:     c.assetBodyMaterials,
		BodyDefinitions:   c.assetBodyDefinitions,
		Bodies:            assetBodies,
		AmbientLights:     assetAmbientLights,
		PointLights:       assetPointLights,
		SpotLights:        assetSpotLights,
		DirectionalLights: assetDirectionalLights,
		Skies:             assetSkies,
	}, nil
}

func (c *Converter) convertAnimation(animation *Animation) asset.Animation {
	assetAnimation := asset.Animation{
		Name:      animation.name,
		StartTime: animation.startTime,
		EndTime:   animation.endTime,
		Bindings:  make([]asset.AnimationBinding, len(animation.bindings)),
	}
	for i, binding := range animation.bindings {
		translationKeyframes := make([]asset.AnimationKeyframe[dprec.Vec3], len(binding.translationKeyframes))
		for j, keyframe := range binding.translationKeyframes {
			translationKeyframes[j] = asset.AnimationKeyframe[dprec.Vec3]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Value,
			}
		}
		rotationKeyframes := make([]asset.AnimationKeyframe[dprec.Quat], len(binding.rotationKeyframes))
		for j, keyframe := range binding.rotationKeyframes {
			rotationKeyframes[j] = asset.AnimationKeyframe[dprec.Quat]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Value,
			}
		}
		scaleKeyframes := make([]asset.AnimationKeyframe[dprec.Vec3], len(binding.scaleKeyframes))
		for j, keyframe := range binding.scaleKeyframes {
			scaleKeyframes[j] = asset.AnimationKeyframe[dprec.Vec3]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Value,
			}
		}
		assetAnimation.Bindings[i] = asset.AnimationBinding{
			NodeName:             binding.nodeName,
			TranslationKeyframes: translationKeyframes,
			RotationKeyframes:    rotationKeyframes,
			ScaleKeyframes:       scaleKeyframes,
		}
	}
	return assetAnimation
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

func (c *Converter) convertBodyMaterial(material *BodyMaterial) (uint32, error) {
	if index, ok := c.convertedBodyMaterials[material]; ok {
		return index, nil
	}

	assetMaterial := asset.BodyMaterial{
		FrictionCoefficient:    material.frictionCoefficient,
		RestitutionCoefficient: material.restitutionCoefficient,
	}

	index := uint32(len(c.assetBodyMaterials))
	c.assetBodyMaterials = append(c.assetBodyMaterials, assetMaterial)
	c.convertedBodyMaterials[material] = index
	return index, nil
}

func (c *Converter) convertBodyDefinition(definition *BodyDefinition) (uint32, error) {
	if index, ok := c.convertedBodyDefinitions[definition]; ok {
		return index, nil
	}

	materialIndex, err := c.convertBodyMaterial(definition.material)
	if err != nil {
		return 0, fmt.Errorf("error converting body material: %w", err)
	}

	assetDefinition := asset.BodyDefinition{
		MaterialIndex:     materialIndex,
		Mass:              definition.mass,
		MomentOfInertia:   definition.momentOfInertia,
		DragFactor:        definition.dragFactor,
		AngularDragFactor: definition.angularDragFactor,
		CollisionBoxes: gog.Map(definition.collisionBoxes, func(box *CollisionBox) asset.CollisionBox {
			return asset.CollisionBox{
				Translation: box.Translation(),
				Rotation:    box.Rotation(),
				Width:       box.Width(),
				Height:      box.Height(),
				Length:      box.Length(),
			}
		}),
		CollisionSpheres: gog.Map(definition.collisionSpheres, func(sphere *CollisionSphere) asset.CollisionSphere {
			return asset.CollisionSphere{
				Translation: sphere.Translation(),
				Radius:      sphere.Radius(),
			}
		}),
		CollisionMeshes: gog.Map(definition.collisionMeshes, func(mesh *CollisionMesh) asset.CollisionMesh {
			return asset.CollisionMesh{
				Translation: mesh.Translation(),
				Rotation:    mesh.Rotation(),
				Triangles: gog.Map(mesh.Triangles(), func(triangle CollisionTriangle) asset.CollisionTriangle {
					return asset.CollisionTriangle{
						A: triangle.A,
						B: triangle.B,
						C: triangle.C,
					}
				}),
			}
		}),
	}

	index := uint32(len(c.assetBodyDefinitions))
	c.assetBodyDefinitions = append(c.assetBodyDefinitions, assetDefinition)
	c.convertedBodyDefinitions[definition] = index
	return index, nil
}

func (c *Converter) convertBody(nodeIndex uint32, body *Body) (asset.Body, error) {
	bodyDefinitionIndex, err := c.convertBodyDefinition(body.definition)
	if err != nil {
		return asset.Body{}, fmt.Errorf("error converting body definition: %w", err)
	}
	return asset.Body{
		NodeIndex:           nodeIndex,
		BodyDefinitionIndex: bodyDefinitionIndex,
	}, nil
}

func (c *Converter) convertMesh(nodeIndex uint32, mesh *Mesh) (asset.Mesh, error) {
	meshDefinitionIndex, err := c.convertMeshDefinition(mesh.definition)
	if err != nil {
		return asset.Mesh{}, fmt.Errorf("error converting mesh definition: %w", err)
	}

	var armatureIndex = asset.UnspecifiedArmatureIndex
	if mesh.armature != nil {
		assetArmatureIndex, err := c.convertArmature(mesh.armature)
		if err != nil {
			return asset.Mesh{}, fmt.Errorf("error converting armature: %w", err)
		}
		armatureIndex = int32(assetArmatureIndex)
	}

	return asset.Mesh{
		NodeIndex:           nodeIndex,
		MeshDefinitionIndex: meshDefinitionIndex,
		ArmatureIndex:       armatureIndex,
	}, nil
}

func (c *Converter) convertGeometry(geometry *Geometry) (uint32, error) {
	if index, ok := c.convertedGeometries[geometry]; ok {
		return index, nil
	}

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

	layout := geometry.vertexFormat
	if layout&VertexFormatCoord != 0 {
		coordBufferIndex = 0
		coordOffset = stride
		stride += 3 * sizeFloat
	} else {
		coordBufferIndex = asset.UnspecifiedBufferIndex
	}
	if layout&VertexFormatNormal != 0 {
		normalBufferIndex = 0
		normalOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		normalBufferIndex = asset.UnspecifiedBufferIndex
	}
	if layout&VertexFormatTangent != 0 {
		tangentBufferIndex = 0
		tangentOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		tangentBufferIndex = asset.UnspecifiedBufferIndex
	}
	if layout&VertexFormatTexCoord != 0 {
		texCoordBufferIndex = 0
		texCoordOffset = stride
		stride += 2 * sizeHalfFloat
	} else {
		texCoordBufferIndex = asset.UnspecifiedBufferIndex
	}
	if layout&VertexFormatColor != 0 {
		colorBufferIndex = 0
		colorOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		colorBufferIndex = asset.UnspecifiedBufferIndex
	}
	if layout&VertexFormatWeights != 0 {
		weightsBufferIndex = 0
		weightsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		weightsBufferIndex = asset.UnspecifiedBufferIndex
	}
	if layout&VertexFormatJoints != 0 {
		jointsBufferIndex = 0
		jointsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		jointsBufferIndex = asset.UnspecifiedBufferIndex
	}

	vertexData := gblob.LittleEndianBlock(make([]byte, len(geometry.vertices)*int(stride)))
	if layout&VertexFormatCoord != 0 {
		offset := int(coordOffset)
		for _, vertex := range geometry.vertices {
			vertexData.SetFloat32(offset+0*sizeFloat, vertex.Coord.X)
			vertexData.SetFloat32(offset+1*sizeFloat, vertex.Coord.Y)
			vertexData.SetFloat32(offset+2*sizeFloat, vertex.Coord.Z)
			offset += int(stride)
		}
	}
	if layout&VertexFormatNormal != 0 {
		offset := int(normalOffset)
		for _, vertex := range geometry.vertices {
			vertexData.SetUint16(offset+0*sizeHalfFloat, float16.Fromfloat32(vertex.Normal.X).Bits())
			vertexData.SetUint16(offset+1*sizeHalfFloat, float16.Fromfloat32(vertex.Normal.Y).Bits())
			vertexData.SetUint16(offset+2*sizeHalfFloat, float16.Fromfloat32(vertex.Normal.Z).Bits())
			offset += int(stride)
		}
	}
	if layout&VertexFormatTangent != 0 {
		offset := int(tangentOffset)
		for _, vertex := range geometry.vertices {
			vertexData.SetUint16(offset+0*sizeHalfFloat, float16.Fromfloat32(vertex.Tangent.X).Bits())
			vertexData.SetUint16(offset+1*sizeHalfFloat, float16.Fromfloat32(vertex.Tangent.Y).Bits())
			vertexData.SetUint16(offset+2*sizeHalfFloat, float16.Fromfloat32(vertex.Tangent.Z).Bits())
			offset += int(stride)
		}
	}
	if layout&VertexFormatTexCoord != 0 {
		offset := int(texCoordOffset)
		for _, vertex := range geometry.vertices {
			vertexData.SetUint16(offset+0*sizeHalfFloat, float16.Fromfloat32(vertex.TexCoord.X).Bits())
			vertexData.SetUint16(offset+1*sizeHalfFloat, float16.Fromfloat32(vertex.TexCoord.Y).Bits())
			offset += int(stride)
		}
	}
	if layout&VertexFormatColor != 0 {
		offset := int(colorOffset)
		for _, vertex := range geometry.vertices {
			vertexData.SetUint8(offset+0*sizeUnsignedByte, uint8(vertex.Color.X*255.0))
			vertexData.SetUint8(offset+1*sizeUnsignedByte, uint8(vertex.Color.Y*255.0))
			vertexData.SetUint8(offset+2*sizeUnsignedByte, uint8(vertex.Color.Z*255.0))
			vertexData.SetUint8(offset+3*sizeUnsignedByte, uint8(vertex.Color.W*255.0))
			offset += int(stride)
		}
	}
	if layout&VertexFormatWeights != 0 {
		offset := int(weightsOffset)
		for _, vertex := range geometry.vertices {
			vertexData.SetUint8(offset+0*sizeUnsignedByte, uint8(vertex.Weights.X*255.0))
			vertexData.SetUint8(offset+1*sizeUnsignedByte, uint8(vertex.Weights.Y*255.0))
			vertexData.SetUint8(offset+2*sizeUnsignedByte, uint8(vertex.Weights.Z*255.0))
			vertexData.SetUint8(offset+3*sizeUnsignedByte, uint8(vertex.Weights.W*255.0))
			offset += int(stride)
		}
	}
	if layout&VertexFormatJoints != 0 {
		offset := int(jointsOffset)
		for _, vertex := range geometry.vertices {
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
	if len(geometry.vertices) >= 0xFFFF {
		indexSize = sizeUnsignedInt
		indexLayout = asset.IndexLayoutUint32
		indexData = gblob.LittleEndianBlock(make([]byte, len(geometry.indices)*sizeUnsignedInt))
		for i, index := range geometry.indices {
			indexData.SetUint32(i*sizeUnsignedInt, uint32(index))
		}
	} else {
		indexSize = sizeUnsignedShort
		indexLayout = asset.IndexLayoutUint16
		indexData = gblob.LittleEndianBlock(make([]byte, len(geometry.indices)*sizeUnsignedShort))
		for i, index := range geometry.indices {
			indexData.SetUint16(i*sizeUnsignedShort, uint16(index))
		}
	}

	assetFragments := make([]asset.Fragment, 0, len(geometry.fragments))
	for _, fragment := range geometry.fragments {
		assetFragments = append(assetFragments, asset.Fragment{
			Name:            fragment.Name(),
			Topology:        fragment.topology,
			IndexByteOffset: uint32(fragment.indexOffset * indexSize),
			IndexCount:      uint32(fragment.indexCount),
		})
	}

	var boundingSphereRadius float64
	for _, vertex := range geometry.vertices {
		boundingSphereRadius = dprec.Max(
			boundingSphereRadius,
			float64(vertex.Coord.Length()),
		)
	}

	assetGeometry := asset.Geometry{
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
		Fragments:            assetFragments,
		BoundingSphereRadius: boundingSphereRadius,
		MinDistance:          geometry.minDistance,
		MaxDistance:          geometry.maxDistance,
	}

	index := uint32(len(c.assetGeometries))
	c.assetGeometries = append(c.assetGeometries, assetGeometry)
	c.convertedGeometries[geometry] = index
	return index, nil
}

func (c *Converter) convertMeshDefinition(definition *MeshDefinition) (uint32, error) {
	if index, ok := c.convertedMeshDefinitions[definition]; ok {
		return index, nil
	}

	geometryIndex, err := c.convertGeometry(definition.geometry)
	if err != nil {
		return 0, fmt.Errorf("error converting geometry: %w", err)
	}
	geometry := c.assetGeometries[geometryIndex]

	var materialBindings []asset.MaterialBinding
	for i, fragment := range geometry.Fragments {
		material, ok := definition.materialBindings[fragment.Name]
		if !ok {
			continue // likely invisible fragment.
		}
		materialIndex, err := c.convertMaterial(material)
		if err != nil {
			return 0, fmt.Errorf("error converting material: %w", err)
		}
		materialBindings = append(materialBindings, asset.MaterialBinding{
			FragmentIndex: uint32(i),
			MaterialIndex: materialIndex,
		})
	}

	assetDefinition := asset.MeshDefinition{
		GeometryIndex:    geometryIndex,
		MaterialBindings: materialBindings,
	}

	index := uint32(len(c.assetMeshDefinitions))
	c.assetMeshDefinitions = append(c.assetMeshDefinitions, assetDefinition)
	c.convertedMeshDefinitions[definition] = index
	return index, nil
}

func (c *Converter) convertArmature(armature *Armature) (uint32, error) {
	if index, ok := c.convertedArmatures[armature]; ok {
		return index, nil
	}

	assetArmature := asset.Armature{
		Joints: gog.Map(armature.joints, func(joint *Joint) asset.Joint {
			return asset.Joint{
				NodeIndex:         c.convertedNodes[joint.node],
				InverseBindMatrix: joint.inverseBindMatrix,
			}
		}),
	}

	index := uint32(len(c.assetArmatures))
	c.assetArmatures = append(c.assetArmatures, assetArmature)
	c.convertedArmatures[armature] = index
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
	if texture.generateMipmaps {
		flags |= asset.TextureFlagMipmapping
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
