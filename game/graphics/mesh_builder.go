package graphics

import (
	"github.com/mokiat/gblob"
	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/x448/float16"
)

func MeshBuilderWithCoords() MeshBuilderOption {
	return func(b *MeshBuilder) {
		b.hasCoords = true
	}
}

func MeshBuilderWithNormals() MeshBuilderOption {
	return func(b *MeshBuilder) {
		b.hasNormals = true
	}
}

func MeshBuilderWithTangents() MeshBuilderOption {
	return func(b *MeshBuilder) {
		b.hasTangents = true
	}
}

func MeshBuilderWithTexCoords() MeshBuilderOption {
	return func(b *MeshBuilder) {
		b.hasTexCoords = true
	}
}

func MeshBuilderWithColors() MeshBuilderOption {
	return func(b *MeshBuilder) {
		b.hasColors = true
	}
}

func MeshBuilderWithJoints() MeshBuilderOption {
	return func(b *MeshBuilder) {
		b.hasJoints = true
	}
}

func MeshBuilderWithWeights() MeshBuilderOption {
	return func(b *MeshBuilder) {
		b.hasWeights = true
	}
}

type MeshBuilderOption func(*MeshBuilder)

func NewMeshBuilder(opts ...MeshBuilderOption) *MeshBuilder {
	fragments := []meshBuilderFragment{
		{
			primitive:   -1,
			material:    nil,
			indexOffset: 0,
			indexCount:  0,
		},
	}
	result := &MeshBuilder{
		fragments:       fragments,
		currentFragment: &fragments[0],
	}
	for _, opt := range opts {
		opt(result)
	}
	return result
}

type MeshBuilder struct {
	hasCoords    bool
	hasNormals   bool
	hasTangents  bool
	hasTexCoords bool
	hasColors    bool
	hasJoints    bool
	hasWeights   bool

	vertices      []meshBuilderVertex
	currentVertex *meshBuilderVertex

	indices []uint32

	fragments       []meshBuilderFragment
	currentFragment *meshBuilderFragment
}

func (mb *MeshBuilder) AddVertex() {
	mb.vertices = append(mb.vertices, meshBuilderVertex{})
	mb.currentVertex = &mb.vertices[len(mb.vertices)-1]
}

func (mb *MeshBuilder) Coord(x, y, z float32) {
	mb.CoordVec3(sprec.NewVec3(x, y, z))
}

func (mb *MeshBuilder) CoordVec3(vec sprec.Vec3) {
	mb.currentVertex.coord = vec
}

func (mb *MeshBuilder) Normal(x, y, z float32) {
	mb.NormalVec3(sprec.NewVec3(x, y, z))
}

func (mb *MeshBuilder) NormalVec3(vec sprec.Vec3) {
	mb.currentVertex.normal = vec
}

func (mb *MeshBuilder) Tangent(x, y, z float32) {
	mb.TangentVec3(sprec.NewVec3(x, y, z))
}

func (mb *MeshBuilder) TangentVec3(vec sprec.Vec3) {
	mb.currentVertex.tangent = vec
}

func (mb *MeshBuilder) TexCoord(x, y float32) {
	mb.TexCoordVec2(sprec.NewVec2(x, y))
}

func (mb *MeshBuilder) TexCoordVec2(vec sprec.Vec2) {
	mb.currentVertex.texCoord = vec
}

func (mb *MeshBuilder) Color(r, g, b, a float32) {
	mb.ColorVec4(sprec.NewVec4(r, g, b, a))
}

func (b *MeshBuilder) ColorVec4(vec sprec.Vec4) {
	b.currentVertex.color = vec
}

func (mb *MeshBuilder) Joints(a, b, c, d uint8) {
	mb.currentVertex.joints = [4]uint8{a, b, c, d}
}

func (mb *MeshBuilder) Weights(a, b, c, d float32) {
	mb.WeightsVec4(sprec.NewVec4(a, b, c, d))
}

func (mb *MeshBuilder) WeightsVec4(vec sprec.Vec4) {
	mb.currentVertex.weights = vec
}

func (mb *MeshBuilder) UseMaterial(material *MaterialDefinition) {
	if material == mb.currentFragment.material {
		return
	}
	if mb.currentFragment.indexCount > 0 {
		mb.fragments = append(mb.fragments, meshBuilderFragment{
			primitive:   mb.currentFragment.primitive,
			material:    material,
			indexOffset: uint32(len(mb.indices)),
			indexCount:  0,
		})
		mb.currentFragment = &mb.fragments[len(mb.fragments)-1]
	} else {
		mb.currentFragment.material = material
	}
}

func (mb *MeshBuilder) UsePrimitive(primitive Primitive) {
	if primitive == mb.currentFragment.primitive {
		return
	}
	if mb.currentFragment.indexCount > 0 {
		mb.fragments = append(mb.fragments, meshBuilderFragment{
			primitive:   primitive,
			material:    mb.currentFragment.material,
			indexOffset: uint32(len(mb.indices)),
			indexCount:  0,
		})
		mb.currentFragment = &mb.fragments[len(mb.fragments)-1]
	} else {
		mb.currentFragment.primitive = primitive
	}
}

func (mb *MeshBuilder) AddIndex(index uint32) {
	mb.indices = append(mb.indices, index)
	mb.currentFragment.indexCount++
}

func (mb *MeshBuilder) BuildInfo() MeshDefinitionInfo {
	var (
		vertexFormat VertexFormat
		vertexOffset int
		vertexData   gblob.LittleEndianBlock
	)
	if mb.hasCoords {
		vertexFormat.HasCoord = true
		vertexFormat.CoordOffsetBytes = vertexOffset
		vertexOffset += 12
	}
	if mb.hasNormals {
		vertexFormat.HasNormal = true
		vertexFormat.NormalOffsetBytes = vertexOffset
		vertexOffset += 6
	}
	if mb.hasTangents {
		vertexFormat.HasTangent = true
		vertexFormat.TangentOffsetBytes = vertexOffset
		vertexOffset += 6
	}
	if mb.hasTexCoords {
		vertexFormat.HasTexCoord = true
		vertexFormat.TexCoordOffsetBytes = vertexOffset
		vertexOffset += 4
	}
	if mb.hasColors {
		vertexFormat.HasColor = true
		vertexFormat.ColorOffsetBytes = vertexOffset
		vertexOffset += 4
	}
	if mb.hasWeights {
		vertexFormat.HasWeights = true
		vertexFormat.WeightsOffsetBytes = vertexOffset
		vertexOffset += 4
	}
	if mb.hasJoints {
		vertexFormat.HasJoints = true
		vertexFormat.JointsOffsetBytes = vertexOffset
		vertexOffset += 4
	}

	vertexStride := vertexOffset
	vertexData = make(gblob.LittleEndianBlock, vertexStride*len(mb.vertices))

	if mb.hasCoords {
		vertexFormat.CoordStrideBytes = vertexOffset
		for i, vertex := range mb.vertices {
			offset := i*vertexStride + vertexFormat.CoordOffsetBytes
			vertexData.SetFloat32(offset+0, vertex.coord.X)
			vertexData.SetFloat32(offset+4, vertex.coord.Y)
			vertexData.SetFloat32(offset+8, vertex.coord.Z)
		}
	}
	if mb.hasNormals {
		vertexFormat.NormalStrideBytes = vertexOffset
		for i, vertex := range mb.vertices {
			offset := i*vertexStride + vertexFormat.NormalOffsetBytes
			vertexData.SetUint16(offset+0, float16.Fromfloat32(vertex.normal.X).Bits())
			vertexData.SetUint16(offset+2, float16.Fromfloat32(vertex.normal.Y).Bits())
			vertexData.SetUint16(offset+4, float16.Fromfloat32(vertex.normal.Z).Bits())
		}
	}
	if mb.hasTangents {
		vertexFormat.TangentStrideBytes = vertexOffset
		for i, vertex := range mb.vertices {
			offset := i*vertexStride + vertexFormat.TangentOffsetBytes
			vertexData.SetUint16(offset+0, float16.Fromfloat32(vertex.tangent.X).Bits())
			vertexData.SetUint16(offset+2, float16.Fromfloat32(vertex.tangent.Y).Bits())
			vertexData.SetUint16(offset+4, float16.Fromfloat32(vertex.tangent.Z).Bits())
		}
	}
	if mb.hasTexCoords {
		vertexFormat.TexCoordStrideBytes = vertexOffset
		for i, vertex := range mb.vertices {
			offset := i*vertexStride + vertexFormat.TexCoordOffsetBytes
			vertexData.SetUint16(offset+0, float16.Fromfloat32(vertex.texCoord.X).Bits())
			vertexData.SetUint16(offset+2, float16.Fromfloat32(vertex.texCoord.Y).Bits())
		}
	}
	if mb.hasColors {
		vertexFormat.ColorStrideBytes = vertexOffset
		for i, vertex := range mb.vertices {
			offset := i*vertexStride + vertexFormat.ColorOffsetBytes
			vertexData.SetUint8(offset+0, uint8(vertex.color.X*255))
			vertexData.SetUint8(offset+1, uint8(vertex.color.Y*255))
			vertexData.SetUint8(offset+2, uint8(vertex.color.Z*255))
			vertexData.SetUint8(offset+3, uint8(vertex.color.W*255))
		}
	}
	if mb.hasWeights {
		vertexFormat.WeightsStrideBytes = vertexOffset
		for i, vertex := range mb.vertices {
			offset := i*vertexStride + vertexFormat.WeightsOffsetBytes
			vertexData.SetUint8(offset+0, uint8(vertex.weights.X*255))
			vertexData.SetUint8(offset+1, uint8(vertex.weights.Y*255))
			vertexData.SetUint8(offset+2, uint8(vertex.weights.Z*255))
			vertexData.SetUint8(offset+3, uint8(vertex.weights.W*255))
		}
	}
	if mb.hasJoints {
		vertexFormat.JointsStrideBytes = vertexOffset
		for i, vertex := range mb.vertices {
			offset := i*vertexStride + vertexFormat.JointsOffsetBytes
			vertexData.SetUint8(offset+0, vertex.joints[0])
			vertexData.SetUint8(offset+1, vertex.joints[1])
			vertexData.SetUint8(offset+2, vertex.joints[2])
			vertexData.SetUint8(offset+3, vertex.joints[3])
		}
	}

	var (
		indexFormat IndexFormat
		indexData   gblob.LittleEndianBlock
		indexSize   int
	)
	if len(mb.indices) > 0xFFFF {
		indexFormat = IndexFormatU32
		indexSize = render.SizeU32
		indexData = make(gblob.LittleEndianBlock, indexSize*len(mb.indices))
		for i, index := range mb.indices {
			indexData.SetUint32(i*4, index)
		}
	} else {
		indexFormat = IndexFormatU16
		indexSize = render.SizeU16
		indexData = make(gblob.LittleEndianBlock, indexSize*len(mb.indices))
		for i, index := range mb.indices {
			indexData.SetUint16(i*2, uint16(index))
		}
	}

	return MeshDefinitionInfo{
		VertexFormat: vertexFormat,
		VertexData:   vertexData,
		Fragments: gog.Map(mb.fragments, func(fragment meshBuilderFragment) MeshFragmentDefinitionInfo {
			return MeshFragmentDefinitionInfo{
				Primitive:   fragment.primitive,
				Material:    fragment.material,
				IndexOffset: int(fragment.indexOffset) * indexSize,
				IndexCount:  int(fragment.indexCount),
			}
		}),
		IndexFormat:          indexFormat,
		IndexData:            indexData,
		BoundingSphereRadius: mb.bsRadius(),
	}
}

func (mb *MeshBuilder) bsRadius() float64 {
	var maxRadius float64
	for _, vertex := range mb.vertices {
		maxRadius = max(maxRadius, float64(vertex.coord.Length()))
	}
	return maxRadius
}

type meshBuilderVertex struct {
	coord    sprec.Vec3
	normal   sprec.Vec3
	tangent  sprec.Vec3
	texCoord sprec.Vec2
	color    sprec.Vec4
	joints   [4]uint8
	weights  sprec.Vec4
}

type meshBuilderFragment struct {
	primitive   Primitive
	material    *MaterialDefinition
	indexOffset uint32
	indexCount  uint32
}
