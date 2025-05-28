package conv

import (
	"fmt"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/dto/meshdto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/storage/chunked"
	"github.com/x448/float16"
)

type MeshSource interface {
	AllArmatures() []*mdl.Armature
	AllGeometries() []*mdl.Geometry
	AllMeshDefinitions() []*mdl.MeshDefinition
	AllMeshPlacements() []mdl.Placed[*mdl.Mesh]
}

func NewMeshConverter() *MeshConverter {
	return &MeshConverter{}
}

type MeshConverter struct{}

func (c *MeshConverter) Convert(target *ds.List[chunked.Chunk], asset any) error {
	src, ok := asset.(MeshSource)
	if !ok {
		return nil
	}
	chunk, err := c.CreateMeshChunk(src)
	if err != nil {
		return err
	}
	target.Add(chunked.FromValue(meshdto.MeshChunkID, chunk))
	return nil
}

func (c *MeshConverter) CreateMeshChunk(src MeshSource) (*meshdto.MeshChunk, error) {
	allArmatures := src.AllArmatures()
	dtoArmatures := make([]meshdto.Armature, len(allArmatures))
	for i, armature := range allArmatures {
		var err error
		dtoArmatures[i], err = c.convertArmature(armature)
		if err != nil {
			return nil, fmt.Errorf("error converting armature: %w", err)
		}
	}

	allGeometries := src.AllGeometries()
	dtoGeometries := make([]meshdto.Geometry, len(allGeometries))
	for i, geometry := range allGeometries {
		var err error
		dtoGeometries[i], err = c.convertGeometry(geometry)
		if err != nil {
			return nil, fmt.Errorf("error converting geometry: %w", err)
		}
	}

	allMeshDefinitions := src.AllMeshDefinitions()
	dtoMeshDefinitions := make([]meshdto.MeshDefinition, len(allMeshDefinitions))
	for i, definition := range allMeshDefinitions {
		var err error
		dtoMeshDefinitions[i], err = c.convertMeshDefinition(definition)
		if err != nil {
			return nil, fmt.Errorf("error converting mesh definition: %w", err)
		}
	}

	allMeshPlacements := src.AllMeshPlacements()
	dtoMeshes := make([]meshdto.Mesh, len(allMeshPlacements))
	for i, placement := range allMeshPlacements {
		var err error
		dtoMeshes[i], err = c.convertMesh(placement.Node, placement.Value)
		if err != nil {
			return nil, fmt.Errorf("error converting mesh: %w", err)
		}
	}

	return &meshdto.MeshChunk{
		Armatures:       dtoArmatures,
		Geometries:      dtoGeometries,
		MeshDefinitions: dtoMeshDefinitions,
		Meshes:          dtoMeshes,
	}, nil
}

func (c *MeshConverter) convertArmature(armature *mdl.Armature) (meshdto.Armature, error) {
	return meshdto.Armature{
		ID: armature.ID(),
		Joints: gog.Map(armature.Joints(), func(joint *mdl.Joint) meshdto.Joint {
			return meshdto.Joint{
				NodeID:            joint.Node().ID(),
				InverseBindMatrix: joint.InverseBindMatrix(),
			}
		}),
	}, nil
}

func (c *MeshConverter) convertGeometry(geometry *mdl.Geometry) (meshdto.Geometry, error) {
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

	layout := geometry.Format()
	if layout&mdl.VertexFormatCoord != 0 {
		coordBufferIndex = 0
		coordOffset = stride
		stride += 3 * sizeFloat
	} else {
		coordBufferIndex = meshdto.UnspecifiedBufferIndex
	}
	if layout&mdl.VertexFormatNormal != 0 {
		normalBufferIndex = 0
		normalOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		normalBufferIndex = meshdto.UnspecifiedBufferIndex
	}
	if layout&mdl.VertexFormatTangent != 0 {
		tangentBufferIndex = 0
		tangentOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		tangentBufferIndex = meshdto.UnspecifiedBufferIndex
	}
	if layout&mdl.VertexFormatTexCoord != 0 {
		texCoordBufferIndex = 0
		texCoordOffset = stride
		stride += 2 * sizeHalfFloat
	} else {
		texCoordBufferIndex = meshdto.UnspecifiedBufferIndex
	}
	if layout&mdl.VertexFormatColor != 0 {
		colorBufferIndex = 0
		colorOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		colorBufferIndex = meshdto.UnspecifiedBufferIndex
	}
	if layout&mdl.VertexFormatWeights != 0 {
		weightsBufferIndex = 0
		weightsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		weightsBufferIndex = meshdto.UnspecifiedBufferIndex
	}
	if layout&mdl.VertexFormatJoints != 0 {
		jointsBufferIndex = 0
		jointsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		jointsBufferIndex = meshdto.UnspecifiedBufferIndex
	}

	vertexData := gblob.LittleEndianBlock(make([]byte, len(geometry.Vertices())*int(stride)))
	if layout&mdl.VertexFormatCoord != 0 {
		offset := int(coordOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetFloat32(offset+0*sizeFloat, vertex.Coord.X)
			vertexData.SetFloat32(offset+1*sizeFloat, vertex.Coord.Y)
			vertexData.SetFloat32(offset+2*sizeFloat, vertex.Coord.Z)
			offset += int(stride)
		}
	}
	if layout&mdl.VertexFormatNormal != 0 {
		offset := int(normalOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetUint16(offset+0*sizeHalfFloat, float16.Fromfloat32(vertex.Normal.X).Bits())
			vertexData.SetUint16(offset+1*sizeHalfFloat, float16.Fromfloat32(vertex.Normal.Y).Bits())
			vertexData.SetUint16(offset+2*sizeHalfFloat, float16.Fromfloat32(vertex.Normal.Z).Bits())
			offset += int(stride)
		}
	}
	if layout&mdl.VertexFormatTangent != 0 {
		offset := int(tangentOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetUint16(offset+0*sizeHalfFloat, float16.Fromfloat32(vertex.Tangent.X).Bits())
			vertexData.SetUint16(offset+1*sizeHalfFloat, float16.Fromfloat32(vertex.Tangent.Y).Bits())
			vertexData.SetUint16(offset+2*sizeHalfFloat, float16.Fromfloat32(vertex.Tangent.Z).Bits())
			offset += int(stride)
		}
	}
	if layout&mdl.VertexFormatTexCoord != 0 {
		offset := int(texCoordOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetUint16(offset+0*sizeHalfFloat, float16.Fromfloat32(vertex.TexCoord.X).Bits())
			vertexData.SetUint16(offset+1*sizeHalfFloat, float16.Fromfloat32(vertex.TexCoord.Y).Bits())
			offset += int(stride)
		}
	}
	if layout&mdl.VertexFormatColor != 0 {
		offset := int(colorOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetUint8(offset+0*sizeUnsignedByte, uint8(vertex.Color.X*255.0))
			vertexData.SetUint8(offset+1*sizeUnsignedByte, uint8(vertex.Color.Y*255.0))
			vertexData.SetUint8(offset+2*sizeUnsignedByte, uint8(vertex.Color.Z*255.0))
			vertexData.SetUint8(offset+3*sizeUnsignedByte, uint8(vertex.Color.W*255.0))
			offset += int(stride)
		}
	}
	if layout&mdl.VertexFormatWeights != 0 {
		offset := int(weightsOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetUint8(offset+0*sizeUnsignedByte, uint8(vertex.Weights.X*255.0))
			vertexData.SetUint8(offset+1*sizeUnsignedByte, uint8(vertex.Weights.Y*255.0))
			vertexData.SetUint8(offset+2*sizeUnsignedByte, uint8(vertex.Weights.Z*255.0))
			vertexData.SetUint8(offset+3*sizeUnsignedByte, uint8(vertex.Weights.W*255.0))
			offset += int(stride)
		}
	}
	if layout&mdl.VertexFormatJoints != 0 {
		offset := int(jointsOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetUint8(offset+0*sizeUnsignedByte, uint8(vertex.Joints[0]))
			vertexData.SetUint8(offset+1*sizeUnsignedByte, uint8(vertex.Joints[1]))
			vertexData.SetUint8(offset+2*sizeUnsignedByte, uint8(vertex.Joints[2]))
			vertexData.SetUint8(offset+3*sizeUnsignedByte, uint8(vertex.Joints[3]))
			offset += int(stride)
		}
	}

	var (
		indexLayout meshdto.IndexLayout
		indexData   gblob.LittleEndianBlock
		indexSize   int
	)
	if len(geometry.Vertices()) >= 0xFFFF {
		indexSize = sizeUnsignedInt
		indexLayout = meshdto.IndexLayoutUint32
		indexData = gblob.LittleEndianBlock(make([]byte, len(geometry.Indices())*sizeUnsignedInt))
		for i, index := range geometry.Indices() {
			indexData.SetUint32(i*sizeUnsignedInt, uint32(index))
		}
	} else {
		indexSize = sizeUnsignedShort
		indexLayout = meshdto.IndexLayoutUint16
		indexData = gblob.LittleEndianBlock(make([]byte, len(geometry.Indices())*sizeUnsignedShort))
		for i, index := range geometry.Indices() {
			indexData.SetUint16(i*sizeUnsignedShort, uint16(index))
		}
	}

	assetFragments := make([]meshdto.Fragment, 0, len(geometry.Fragments()))
	for _, fragment := range geometry.Fragments() {
		assetFragments = append(assetFragments, meshdto.Fragment{
			Name:            fragment.Name(),
			Topology:        fragment.Topology(),
			IndexByteOffset: uint32(fragment.IndexOffset() * indexSize),
			IndexCount:      uint32(fragment.IndexCount()),
		})
	}

	var boundingSphereRadius float64
	for _, vertex := range geometry.Vertices() {
		boundingSphereRadius = dprec.Max(
			boundingSphereRadius,
			float64(vertex.Coord.Length()),
		)
	}

	return meshdto.Geometry{
		ID: geometry.ID(),
		VertexBuffers: []meshdto.VertexBuffer{
			{
				Stride: stride,
				Data:   vertexData,
			},
		},
		VertexLayout: meshdto.VertexLayout{
			Coord: meshdto.VertexAttribute{
				BufferIndex: coordBufferIndex,
				ByteOffset:  coordOffset,
				Format:      meshdto.VertexAttributeFormatRGB32F,
			},
			Normal: meshdto.VertexAttribute{
				BufferIndex: normalBufferIndex,
				ByteOffset:  normalOffset,
				Format:      meshdto.VertexAttributeFormatRGB16F,
			},
			Tangent: meshdto.VertexAttribute{
				BufferIndex: tangentBufferIndex,
				ByteOffset:  tangentOffset,
				Format:      meshdto.VertexAttributeFormatRGB16F,
			},
			TexCoord: meshdto.VertexAttribute{
				BufferIndex: texCoordBufferIndex,
				ByteOffset:  texCoordOffset,
				Format:      meshdto.VertexAttributeFormatRG16F,
			},
			Color: meshdto.VertexAttribute{
				BufferIndex: colorBufferIndex,
				ByteOffset:  colorOffset,
				Format:      meshdto.VertexAttributeFormatRGBA8UN,
			},
			Weights: meshdto.VertexAttribute{
				BufferIndex: weightsBufferIndex,
				ByteOffset:  weightsOffset,
				Format:      meshdto.VertexAttributeFormatRGBA8UN,
			},
			Joints: meshdto.VertexAttribute{
				BufferIndex: jointsBufferIndex,
				ByteOffset:  jointsOffset,
				Format:      meshdto.VertexAttributeFormatRGBA8IU,
			},
		},
		IndexBuffer: meshdto.IndexBuffer{
			IndexLayout: indexLayout,
			Data:        indexData,
		},
		Fragments:            assetFragments,
		BoundingSphereRadius: boundingSphereRadius,
		MinDistance:          geometry.MinDistance(),
		MaxDistance:          geometry.MaxDistance(),
		MaxCascade:           uint8(geometry.MaxCascade()),
	}, nil
}

func (c *MeshConverter) convertMeshDefinition(definition *mdl.MeshDefinition) (meshdto.MeshDefinition, error) {
	geometry := definition.Geometry()

	var materialBindings []meshdto.MaterialBinding
	for i, fragment := range geometry.Fragments() {
		material, ok := definition.MaterialBindings()[fragment.Name()]
		if !ok {
			continue // likely invisible fragment.
		}
		materialBindings = append(materialBindings, meshdto.MaterialBinding{
			FragmentIndex: uint32(i),
			MaterialID:    material.ID(),
		})
	}

	return meshdto.MeshDefinition{
		ID:               definition.ID(),
		GeometryID:       geometry.ID(),
		MaterialBindings: materialBindings,
	}, nil
}

func (c *MeshConverter) convertMesh(node *mdl.Node, mesh *mdl.Mesh) (meshdto.Mesh, error) {
	armatureID := meshdto.UnspecifiedArmatureID
	if armature := mesh.Armature(); armature != nil {
		armatureID = armature.ID()
	}
	return meshdto.Mesh{
		ID:               mesh.ID(),
		NodeID:           node.ID(),
		MeshDefinitionID: mesh.Definition().ID(),
		ArmatureID:       armatureID,
	}, nil
}
