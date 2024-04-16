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
		assetMaterialIndexFromMaterial:        make(map[*Material]int),
		assetArmatureIndexFromArmature:        make(map[*Armature]int),
		assetMeshDefinitionFromMeshDefinition: make(map[*MeshDefinition]int),
		assetBodyDefinitionFromMeshDefinition: make(map[*MeshDefinition]int),
	}
}

type converter struct {
	forceCollidable                       bool
	assetNodes                            []newasset.Node
	assetNodeIndexFromNode                map[*Node]int
	assetMaterialIndexFromMaterial        map[*Material]int
	assetArmatureIndexFromArmature        map[*Armature]int
	assetMeshDefinitionFromMeshDefinition map[*MeshDefinition]int
	assetBodyDefinitionFromMeshDefinition map[*MeshDefinition]int
}

func (c *converter) BuildModel(model *Model) *asset.Model {
	for _, node := range model.RootNodes {
		c.BuildNode(-1, node)
	}

	assetTextures := make([]newasset.Texture, len(model.Textures))
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

	assetArmatures := make([]newasset.Armature, len(model.Armatures))
	for i, armature := range model.Armatures {
		assetArmatures[i] = c.BuildArmature(armature)
		c.assetArmatureIndexFromArmature[armature] = i
	}

	assetMeshDefinitions := make([]asset.MeshDefinition, len(model.MeshDefinitions))
	for i, meshDefinition := range model.MeshDefinitions {
		assetMeshDefinitions[i] = c.BuildMeshDefinition(meshDefinition)
		c.assetMeshDefinitionFromMeshDefinition[meshDefinition] = i
	}

	assetMeshInstances := make([]asset.MeshInstance, len(model.MeshInstances))
	for i, meshInstance := range model.MeshInstances {
		assetMeshInstances[i] = c.BuildMeshInstance(meshInstance)
	}

	assetBodyDefinitions := make([]asset.BodyDefinition, len(model.MeshDefinitions))
	for i, meshDefinition := range model.MeshDefinitions {
		assetBodyDefinitions[i] = c.BuildBodyDefinition(meshDefinition)
		c.assetBodyDefinitionFromMeshDefinition[meshDefinition] = i
	}

	assetBodyInstances := make([]asset.BodyInstance, 0, len(model.MeshInstances))
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
		Animations:        assetAnimations,
		Armatures:         assetArmatures,
		Textures:          assetTextures,
		Materials:         assetMaterials,
		MeshDefinitions:   assetMeshDefinitions,
		MeshInstances:     assetMeshInstances,
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

func (c *converter) BuildAnimation(animation *Animation) asset.Animation {
	assetAnimation := asset.Animation{
		Name:      animation.Name,
		StartTime: animation.StartTime,
		EndTime:   animation.EndTime,
		Bindings:  make([]asset.AnimationBinding, len(animation.Bindings)),
	}
	for i, binding := range animation.Bindings {
		translationKeyframes := make([]asset.TranslationKeyframe, len(binding.TranslationKeyframes))
		for j, keyframe := range binding.TranslationKeyframes {
			translationKeyframes[j] = asset.TranslationKeyframe{
				Timestamp:   keyframe.Timestamp,
				Translation: keyframe.Translation,
			}
		}
		rotationKeyframes := make([]asset.RotationKeyframe, len(binding.RotationKeyframes))
		for j, keyframe := range binding.RotationKeyframes {
			rotationKeyframes[j] = asset.RotationKeyframe{
				Timestamp: keyframe.Timestamp,
				Rotation:  keyframe.Rotation,
			}
		}
		scaleKeyframes := make([]asset.ScaleKeyframe, len(binding.ScaleKeyframes))
		for j, keyframe := range binding.ScaleKeyframes {
			scaleKeyframes[j] = asset.ScaleKeyframe{
				Timestamp: keyframe.Timestamp,
				Scale:     keyframe.Scale,
			}
		}
		assetAnimation.Bindings[i] = asset.AnimationBinding{
			NodeIndex:            int32(c.assetNodeIndexFromNode[binding.Node]),
			TranslationKeyframes: translationKeyframes,
			RotationKeyframes:    rotationKeyframes,
			ScaleKeyframes:       scaleKeyframes,
		}
	}
	return assetAnimation
}

func (c *converter) BuildMaterial(material *Material) asset.Material {
	assetMaterial := asset.Material{
		Name:            material.Name,
		Type:            asset.MaterialTypePBR,
		BackfaceCulling: material.BackfaceCulling,
		AlphaTesting:    material.AlphaTesting,
		AlphaThreshold:  material.AlphaThreshold,
		Blending:        material.Blending,
		ScalarMask:      0xFFFFFFFF, // TODO: In PBRView
	}
	for i := range assetMaterial.Textures {
		assetMaterial.Textures[i] = asset.TextureRef{
			TextureIndex: asset.UnspecifiedIndex,
		}
	}
	pbr := asset.NewPBRMaterialView(&assetMaterial)
	pbr.SetBaseColor(material.Color)
	if ref := material.ColorTexture; ref != nil {
		pbr.SetBaseColorTexture(asset.TextureRef{
			TextureIndex: int32(ref.TextureIndex),
		})
	}
	pbr.SetMetallic(material.Metallic)
	pbr.SetRoughness(material.Roughness)
	if ref := material.MetallicRoughnessTexture; ref != nil {
		pbr.SetMetallicRoughnessTexture(asset.TextureRef{
			TextureIndex: int32(ref.TextureIndex),
		})
	}
	pbr.SetNormalScale(material.NormalScale)
	if ref := material.NormalTexture; ref != nil {
		pbr.SetNormalTexture(asset.TextureRef{
			TextureIndex: int32(ref.TextureIndex),
		})
	}
	return assetMaterial
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

func (c *converter) BuildMeshDefinition(meshDefinition *MeshDefinition) asset.MeshDefinition {
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
		coordOffset = asset.UnspecifiedOffset
	}
	if layout.HasNormals {
		normalOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		normalOffset = asset.UnspecifiedOffset
	}
	if layout.HasTangents {
		tangentOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		tangentOffset = asset.UnspecifiedOffset
	}
	if layout.HasTexCoords {
		texCoordOffset = stride
		stride += 2 * sizeHalfFloat
	} else {
		texCoordOffset = asset.UnspecifiedOffset
	}
	if layout.HasColors {
		colorOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		colorOffset = asset.UnspecifiedOffset
	}
	if layout.HasWeights {
		weightsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		weightsOffset = asset.UnspecifiedOffset
	}
	if layout.HasJoints {
		jointsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		jointsOffset = asset.UnspecifiedOffset
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
		fragments = make([]asset.MeshFragment, 0, len(meshDefinition.Fragments))
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

	return asset.MeshDefinition{
		Name: meshDefinition.Name,
		VertexLayout: asset.VertexLayout{
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

func (c *converter) BuildFragment(fragment MeshFragment, indexSize int) asset.MeshFragment {
	var topology asset.MeshTopology
	switch fragment.Primitive {
	case PrimitivePoints:
		topology = asset.MeshTopologyPoints
	case PrimitiveLines:
		topology = asset.MeshTopologyLines
	case PrimitiveLineStrip:
		topology = asset.MeshTopologyLineStrip
	case PrimitiveLineLoop:
		topology = asset.MeshTopologyLineLoop
	case PrimitiveTriangles:
		topology = asset.MeshTopologyTriangles
	case PrimitiveTriangleStrip:
		topology = asset.MeshTopologyTriangleStrip
	case PrimitiveTriangleFan:
		topology = asset.MeshTopologyTriangleFan
	default:
		panic(fmt.Errorf("unsupported primitive type: %d", fragment.Primitive))
	}

	var materialIndex int32
	if index, ok := c.assetMaterialIndexFromMaterial[fragment.Material]; ok {
		materialIndex = int32(index)
	} else {
		panic(fmt.Errorf("material %s not found", fragment.Material.Name))
	}

	return asset.MeshFragment{
		Topology:      topology,
		IndexOffset:   uint32(fragment.IndexOffset * indexSize),
		IndexCount:    uint32(fragment.IndexCount),
		MaterialIndex: materialIndex,
	}
}

func (c *converter) BuildMeshInstance(meshInstance *MeshInstance) asset.MeshInstance {
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
	return asset.MeshInstance{
		Name:            meshInstance.Name,
		NodeIndex:       nodeIndex,
		ArmatureIndex:   armatureIndex,
		DefinitionIndex: definitionIndex,
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
		Name:            meshDefinition.Name,
		CollisionMeshes: meshes,
	}
}

func (c *converter) BuildBodyInstance(meshInstance *MeshInstance) asset.BodyInstance {
	var nodeIndex int32
	if index, ok := c.assetNodeIndexFromNode[meshInstance.Node]; ok {
		nodeIndex = int32(index)
	} else {
		panic(fmt.Errorf("node %s not found", meshInstance.Node.Name))
	}
	var definitionIndex int32
	if index, ok := c.assetBodyDefinitionFromMeshDefinition[meshInstance.Definition]; ok {
		definitionIndex = int32(index)
	} else {
		panic(fmt.Errorf("body definition %s not found", meshInstance.Definition.Name))
	}
	return asset.BodyInstance{
		Name:      meshInstance.Name,
		NodeIndex: nodeIndex,
		BodyIndex: definitionIndex,
	}
}
