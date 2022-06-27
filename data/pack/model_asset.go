package pack

import (
	"fmt"

	"github.com/mokiat/lacking/data"
	gameasset "github.com/mokiat/lacking/game/asset"
	"github.com/x448/float16"
)

type SaveModelAssetAction struct {
	registry      gameasset.Registry
	id            string
	modelProvider ModelProvider
}

func (a *SaveModelAssetAction) Describe() string {
	return fmt.Sprintf("save_model_asset(id: %q)", a.id)
}

func (a *SaveModelAssetAction) Run() error {
	conv := newConverter()
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

func newConverter() *converter {
	return &converter{
		assetNodes:                            make([]gameasset.Node, 0),
		assetNodeIndexFromNode:                make(map[*Node]int),
		assetMaterialIndexFromMaterial:        make(map[*Material]int),
		assetMeshDefinitionFromMeshDefinition: make(map[*MeshDefinition]int),
	}
}

type converter struct {
	assetNodes                            []gameasset.Node
	assetNodeIndexFromNode                map[*Node]int
	assetMaterialIndexFromMaterial        map[*Material]int
	assetMeshDefinitionFromMeshDefinition map[*MeshDefinition]int
}

func (c *converter) BuildModel(model *Model) *gameasset.Model {
	for _, node := range model.RootNodes {
		c.BuildNode(-1, node)
	}

	var (
		assetMaterials = make([]gameasset.Material, len(model.Materials))
	)
	for i, material := range model.Materials {
		assetMaterials[i] = c.BuildMaterial(material)
		c.assetMaterialIndexFromMaterial[material] = i
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

	return &gameasset.Model{
		Nodes:           c.assetNodes,
		Materials:       assetMaterials,
		MeshDefinitions: assetMeshDefinitions,
		MeshInstances:   assetMeshInstances,
	}
}

func (c *converter) BuildNode(parentIndex int, node *Node) {
	result := gameasset.Node{
		Name:        node.Name,
		ParentIndex: int32(parentIndex),
		Translation: node.Translation.Array(),
		Rotation: [4]float32{
			node.Rotation.X,
			node.Rotation.Y,
			node.Rotation.Z,
			node.Rotation.W,
		},
		Scale: node.Scale.Array(),
	}
	index := len(c.assetNodes)
	c.assetNodes = append(c.assetNodes, result)
	c.assetNodeIndexFromNode[node] = index
	for _, child := range node.Children {
		c.BuildNode(index, child)
	}
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
		VertexData:  vertexData,
		IndexLayout: indexLayout,
		IndexData:   indexData,
		Fragments:   fragments,
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

	return gameasset.MeshInstance{
		Name:            meshInstance.Name,
		NodeIndex:       nodeIndex,
		DefinitionIndex: definitionIndex,
	}
}
