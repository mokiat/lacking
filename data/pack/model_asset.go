package pack

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/stod"
	"github.com/mokiat/lacking/data"
	gameasset "github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/log"
	"github.com/x448/float16"
)

type SaveModelAssetOption func(a *SaveModelAssetAction)

func WithCollisionMesh(collisionMesh bool) SaveModelAssetOption {
	return func(a *SaveModelAssetAction) {
		a.collisionMesh = collisionMesh
	}
}

type SaveModelAssetAction struct {
	registry      gameasset.Registry
	id            string
	modelProvider ModelProvider
	collisionMesh bool
}

func (a *SaveModelAssetAction) Describe() string {
	return fmt.Sprintf("save_model_asset(id: %q)", a.id)
}

func (a *SaveModelAssetAction) Run() error {
	conv := newConverter(a.collisionMesh)
	modelAsset := conv.BuildModel(a.modelProvider.Model())
	resource := a.registry.ResourceByID(a.id)
	if resource == nil {
		resource = a.registry.CreateIDResource(a.id, "model", a.id)
	}
	if err := resource.WriteContent(modelAsset); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	if err := a.registry.Save(); err != nil {
		return fmt.Errorf("error saving resources: %w", err)
	}
	return nil
}

func newConverter(collisionMeshes bool) *converter {
	return &converter{
		collisionMeshes:                       collisionMeshes,
		assetNodes:                            make([]gameasset.Node, 0),
		assetNodeIndexFromNode:                make(map[*Node]int),
		assetMaterialIndexFromMaterial:        make(map[*Material]int),
		assetArmatureIndexFromArmature:        make(map[*Armature]int),
		assetMeshDefinitionFromMeshDefinition: make(map[*MeshDefinition]int),
	}
}

type converter struct {
	collisionMeshes                       bool
	assetNodes                            []gameasset.Node
	assetNodeIndexFromNode                map[*Node]int
	assetMaterialIndexFromMaterial        map[*Material]int
	assetArmatureIndexFromArmature        map[*Armature]int
	assetMeshDefinitionFromMeshDefinition map[*MeshDefinition]int
}

func (c *converter) BuildModel(model *Model) *gameasset.Model {
	for _, node := range model.RootNodes {
		c.BuildNode(-1, node)
	}

	var (
		assetAnimations = make([]gameasset.Animation, len(model.Animations))
	)
	for i, animation := range model.Animations {
		assetAnimations[i] = c.BuildAnimation(animation)
	}

	var (
		assetMaterials = make([]gameasset.Material, len(model.Materials))
	)
	for i, material := range model.Materials {
		assetMaterials[i] = c.BuildMaterial(material)
		c.assetMaterialIndexFromMaterial[material] = i
	}

	var (
		assetArmatures = make([]gameasset.Armature, len(model.Armatures))
	)
	for i, armature := range model.Armatures {
		assetArmatures[i] = c.BuildArmature(armature)
		c.assetArmatureIndexFromArmature[armature] = i
	}

	var (
		assetMeshDefinitions = make([]gameasset.MeshDefinition, len(model.MeshDefinitions))
	)
	for i, meshDefinition := range model.MeshDefinitions {
		assetMeshDefinitions[i] = c.BuildMeshDefinition(meshDefinition)
		c.assetMeshDefinitionFromMeshDefinition[meshDefinition] = i
	}

	var (
		assetMeshInstances = make([]gameasset.MeshInstance, len(model.MeshInstances))
	)
	for i, meshInstance := range model.MeshInstances {
		assetMeshInstances[i] = c.BuildMeshInstance(meshInstance)
	}

	var (
		assetBodyDefinitions []gameasset.BodyDefinition
		assetBodyInstances   []gameasset.BodyInstance
	)
	if c.collisionMeshes {
		assetBodyDefinitions = make([]gameasset.BodyDefinition, len(model.MeshDefinitions))
		for i, meshDefinition := range model.MeshDefinitions {
			assetBodyDefinitions[i] = c.BuildBodyDefinition(meshDefinition)
		}

		assetBodyInstances = make([]gameasset.BodyInstance, len(model.MeshInstances))
		for i, meshInstance := range model.MeshInstances {
			assetBodyInstances[i] = c.BuildBodyInstance(meshInstance)
		}
	}

	return &gameasset.Model{
		Nodes:           c.assetNodes,
		Animations:      assetAnimations,
		Armatures:       assetArmatures,
		Materials:       assetMaterials,
		MeshDefinitions: assetMeshDefinitions,
		MeshInstances:   assetMeshInstances,
		BodyDefinitions: assetBodyDefinitions,
		BodyInstances:   assetBodyInstances,
	}
}

func (c *converter) BuildNode(parentIndex int, node *Node) {
	result := gameasset.Node{
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

func (c *converter) BuildAnimation(animation *Animation) gameasset.Animation {
	assetAnimation := gameasset.Animation{
		Name:      animation.Name,
		StartTime: animation.StartTime,
		EndTime:   animation.EndTime,
		Bindings:  make([]gameasset.AnimationBinding, len(animation.Bindings)),
	}
	for i, binding := range animation.Bindings {
		translationKeyframes := make([]gameasset.TranslationKeyframe, len(binding.TranslationKeyframes))
		for j, keyframe := range binding.TranslationKeyframes {
			translationKeyframes[j] = gameasset.TranslationKeyframe{
				Timestamp:   keyframe.Timestamp,
				Translation: keyframe.Translation,
			}
		}
		rotationKeyframes := make([]gameasset.RotationKeyframe, len(binding.RotationKeyframes))
		for j, keyframe := range binding.RotationKeyframes {
			rotationKeyframes[j] = gameasset.RotationKeyframe{
				Timestamp: keyframe.Timestamp,
				Rotation:  keyframe.Rotation,
			}
		}
		scaleKeyframes := make([]gameasset.ScaleKeyframe, len(binding.ScaleKeyframes))
		for j, keyframe := range binding.ScaleKeyframes {
			scaleKeyframes[j] = gameasset.ScaleKeyframe{
				Timestamp: keyframe.Timestamp,
				Scale:     keyframe.Scale,
			}
		}
		assetAnimation.Bindings[i] = gameasset.AnimationBinding{
			NodeIndex:            int32(c.assetNodeIndexFromNode[binding.Node]),
			TranslationKeyframes: translationKeyframes,
			RotationKeyframes:    rotationKeyframes,
			ScaleKeyframes:       scaleKeyframes,
		}
	}
	return assetAnimation
}

func (c *converter) BuildMaterial(material *Material) gameasset.Material {
	assetMaterial := gameasset.Material{
		Name:            material.Name,
		Type:            gameasset.MaterialTypePBR,
		BackfaceCulling: material.BackfaceCulling,
		AlphaTesting:    material.AlphaTesting,
		AlphaThreshold:  material.AlphaThreshold,
		Blending:        material.Blending,
		ScalarMask:      0xFFFFFFFF, // TODO: In PBRView
	}
	pbr := gameasset.NewPBRMaterialView(&assetMaterial)
	pbr.SetBaseColor(material.Color)
	pbr.SetBaseColorTexture(material.ColorTexture)
	pbr.SetMetallic(material.Metallic)
	pbr.SetRoughness(material.Roughness)
	pbr.SetMetallicRoughnessTexture(material.MetallicRoughnessTexture)
	pbr.SetNormalScale(material.NormalScale)
	pbr.SetNormalTexture(material.NormalTexture)
	return assetMaterial
}

func (c *converter) BuildArmature(armature *Armature) gameasset.Armature {
	assetArmature := gameasset.Armature{
		Joints: make([]gameasset.Joint, len(armature.Joints)),
	}
	for i, joint := range armature.Joints {
		assetArmature.Joints[i] = gameasset.Joint{
			NodeIndex:         int32(c.assetNodeIndexFromNode[joint.Node]),
			InverseBindMatrix: joint.InverseBindMatrix.ColumnMajorArray(),
		}
	}
	return assetArmature
}

func (c *converter) BuildMeshDefinition(meshDefinition *MeshDefinition) gameasset.MeshDefinition {
	const (
		sizeUnsignedByte  = 1
		sizeUnsignedShort = 2
		sizeUnsignedInt   = 4
		sizeHalfFloat     = 2
		sizeFloat         = 4
	)

	var (
		stride         int32
		coordOffset    int32
		normalOffset   int32
		tangentOffset  int32
		texCoordOffset int32
		colorOffset    int32
		weightsOffset  int32
		jointsOffset   int32
	)

	layout := meshDefinition.VertexLayout
	if layout.HasCoords {
		coordOffset = stride
		stride += 3 * sizeFloat
	} else {
		coordOffset = gameasset.UnspecifiedOffset
	}
	if layout.HasNormals {
		normalOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		normalOffset = gameasset.UnspecifiedOffset
	}
	if layout.HasTangents {
		tangentOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		tangentOffset = gameasset.UnspecifiedOffset
	}
	if layout.HasTexCoords {
		texCoordOffset = stride
		stride += 2 * sizeHalfFloat
	} else {
		texCoordOffset = gameasset.UnspecifiedOffset
	}
	if layout.HasColors {
		colorOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		colorOffset = gameasset.UnspecifiedOffset
	}
	if layout.HasWeights {
		weightsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		weightsOffset = gameasset.UnspecifiedOffset
	}
	if layout.HasJoints {
		jointsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		jointsOffset = gameasset.UnspecifiedOffset
	}

	var (
		vertexData = data.Buffer(make([]byte, len(meshDefinition.Vertices)*int(stride)))
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
		indexLayout gameasset.IndexLayout
		indexData   data.Buffer
		indexSize   int
	)
	if len(meshDefinition.Vertices) >= 0xFFFF {
		indexSize = sizeUnsignedInt
		indexLayout = gameasset.IndexLayoutUint32
		indexData = data.Buffer(make([]byte, len(meshDefinition.Indices)*sizeUnsignedInt))
		for i, index := range meshDefinition.Indices {
			indexData.SetUint32(i*sizeUnsignedInt, uint32(index))
		}
	} else {
		indexSize = sizeUnsignedShort
		indexLayout = gameasset.IndexLayoutUint16
		indexData = data.Buffer(make([]byte, len(meshDefinition.Indices)*sizeUnsignedShort))
		for i, index := range meshDefinition.Indices {
			indexData.SetUint16(i*sizeUnsignedShort, uint16(index))
		}
	}

	var (
		fragments = make([]gameasset.MeshFragment, len(meshDefinition.Fragments))
	)
	for i, fragment := range meshDefinition.Fragments {
		fragments[i] = c.BuildFragment(fragment, indexSize)
	}

	var boundingSphereRadius float64
	for _, vertex := range meshDefinition.Vertices {
		boundingSphereRadius = dprec.Max(
			boundingSphereRadius,
			float64(vertex.Coord.Length()),
		)
	}

	return gameasset.MeshDefinition{
		Name: meshDefinition.Name,
		VertexLayout: gameasset.VertexLayout{
			CoordOffset:    coordOffset,
			CoordStride:    stride,
			NormalOffset:   normalOffset,
			NormalStride:   stride,
			TangentOffset:  tangentOffset,
			TangentStride:  stride,
			TexCoordOffset: texCoordOffset,
			TexCoordStride: stride,
			ColorOffset:    colorOffset,
			ColorStride:    stride,
			WeightsOffset:  weightsOffset,
			WeightsStride:  stride,
			JointsOffset:   jointsOffset,
			JointsStride:   stride,
		},
		VertexData:           vertexData,
		IndexLayout:          indexLayout,
		IndexData:            indexData,
		Fragments:            fragments,
		BoundingSphereRadius: boundingSphereRadius,
	}
}

func (c *converter) BuildFragment(fragment MeshFragment, indexSize int) gameasset.MeshFragment {
	var topology gameasset.MeshTopology
	switch fragment.Primitive {
	case PrimitivePoints:
		topology = gameasset.MeshTopologyPoints
	case PrimitiveLines:
		topology = gameasset.MeshTopologyLines
	case PrimitiveLineStrip:
		topology = gameasset.MeshTopologyLineStrip
	case PrimitiveLineLoop:
		topology = gameasset.MeshTopologyLineLoop
	case PrimitiveTriangles:
		topology = gameasset.MeshTopologyTriangles
	case PrimitiveTriangleStrip:
		topology = gameasset.MeshTopologyTriangleStrip
	case PrimitiveTriangleFan:
		topology = gameasset.MeshTopologyTriangleFan
	default:
		panic(fmt.Errorf("unsupported primitive type: %d", fragment.Primitive))
	}

	var materialIndex int32
	if index, ok := c.assetMaterialIndexFromMaterial[fragment.Material]; ok {
		materialIndex = int32(index)
	} else {
		panic(fmt.Errorf("material %s not found", fragment.Material.Name))
	}

	return gameasset.MeshFragment{
		Topology:      topology,
		IndexOffset:   uint32(fragment.IndexOffset * indexSize),
		IndexCount:    uint32(fragment.IndexCount),
		MaterialIndex: materialIndex,
	}
}

func (c *converter) BuildMeshInstance(meshInstance *MeshInstance) gameasset.MeshInstance {
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
	var armatureIndex int32 = gameasset.UnspecifiedArmatureIndex
	if meshInstance.Armature != nil {
		if index, ok := c.assetArmatureIndexFromArmature[meshInstance.Armature]; ok {
			armatureIndex = int32(index)
		} else {
			panic(fmt.Errorf("armature not found"))
		}
	}
	return gameasset.MeshInstance{
		Name:            meshInstance.Name,
		NodeIndex:       nodeIndex,
		ArmatureIndex:   armatureIndex,
		DefinitionIndex: definitionIndex,
	}
}

func (c *converter) BuildBodyDefinition(meshDefinition *MeshDefinition) gameasset.BodyDefinition {
	var triangles []gameasset.CollisionTriangle

	for _, fragment := range meshDefinition.Fragments {
		if fragment.Primitive != PrimitiveTriangles {
			log.Warn("Skipping collision mesh due to primitive no being triangles")
			continue
		}
		for i := fragment.IndexOffset; i < fragment.IndexOffset+fragment.IndexCount; i += 3 {
			indexA := meshDefinition.Indices[i+0]
			indexB := meshDefinition.Indices[i+1]
			indexC := meshDefinition.Indices[i+2]

			coordA := meshDefinition.Vertices[indexA].Coord
			coordB := meshDefinition.Vertices[indexB].Coord
			coordC := meshDefinition.Vertices[indexC].Coord

			triangles = append(triangles, gameasset.CollisionTriangle{
				A: stod.Vec3(coordA),
				B: stod.Vec3(coordB),
				C: stod.Vec3(coordC),
			})
		}
	}

	return gameasset.BodyDefinition{
		Name: meshDefinition.Name,
		CollisionMeshes: []gameasset.CollisionMesh{
			{
				Translation: dprec.ZeroVec3(),
				Rotation:    dprec.IdentityQuat(),
				Triangles:   triangles,
			},
		},
	}
}

func (c *converter) BuildBodyInstance(meshInstance *MeshInstance) gameasset.BodyInstance {
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
	return gameasset.BodyInstance{
		Name:      meshInstance.Name,
		NodeIndex: nodeIndex,
		BodyIndex: definitionIndex,
	}
}
