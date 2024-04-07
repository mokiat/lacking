package graphics

import (
	"github.com/mokiat/gblob"
	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/x448/float16"
)

// MeshGeometryBuilderWithCoords is a MeshGeometryBuilderOption that enables the
// construction of vertices with coordinates.
func MeshGeometryBuilderWithCoords() MeshGeometryBuilderOption {
	return func(b *MeshGeometryBuilder) {
		b.hasCoords = true
	}
}

// MeshGeometryBuilderWithNormals is a MeshGeometryBuilderOption that enables the
// construction of vertices with normals.
func MeshGeometryBuilderWithNormals() MeshGeometryBuilderOption {
	return func(b *MeshGeometryBuilder) {
		b.hasNormals = true
	}
}

// MeshGeometryBuilderWithTangents is a MeshGeometryBuilderOption that enables the
// construction of vertices with tangents.
func MeshGeometryBuilderWithTangents() MeshGeometryBuilderOption {
	return func(b *MeshGeometryBuilder) {
		b.hasTangents = true
	}
}

// MeshGeometryBuilderWithTexCoords is a MeshGeometryBuilderOption that enables the
// construction of vertices with texture coordinates.
func MeshGeometryBuilderWithTexCoords() MeshGeometryBuilderOption {
	return func(b *MeshGeometryBuilder) {
		b.hasTexCoords = true
	}
}

// MeshGeometryBuilderWithColors is a MeshGeometryBuilderOption that enables the
// construction of vertices with colors.
func MeshGeometryBuilderWithColors() MeshGeometryBuilderOption {
	return func(b *MeshGeometryBuilder) {
		b.hasColors = true
	}
}

// MeshGeometryBuilderWithJoints is a MeshGeometryBuilderOption that enables the
// construction of vertices with joint indices.
func MeshGeometryBuilderWithJoints() MeshGeometryBuilderOption {
	return func(b *MeshGeometryBuilder) {
		b.hasJoints = true
	}
}

// MeshGeometryBuilderWithWeights is a MeshGeometryBuilderOption that enables the
// construction of vertices with joint weights.
func MeshGeometryBuilderWithWeights() MeshGeometryBuilderOption {
	return func(b *MeshGeometryBuilder) {
		b.hasWeights = true
	}
}

// MeshGeometryBuilderOption is a function that modifies a MeshGeometryBuilder.
type MeshGeometryBuilderOption func(*MeshGeometryBuilder)

// NewMeshGeometryBuilder creates a new MeshGeometryBuilder with the provided options.
func NewMeshGeometryBuilder(opts ...MeshGeometryBuilderOption) *MeshGeometryBuilder {
	result := &MeshGeometryBuilder{
		transform: sprec.IdentityMat4(),
	}
	for _, opt := range opts {
		opt(result)
	}
	return result
}

// MeshGeometryBuilder is a helper for constructing a MeshDefinitionInfo.
type MeshGeometryBuilder struct {
	hasCoords    bool
	hasNormals   bool
	hasTangents  bool
	hasTexCoords bool
	hasColors    bool
	hasJoints    bool
	hasWeights   bool

	transform sprec.Mat4
	vertices  []meshBuilderVertex
	indices   []uint32
	fragments []meshBuilderFragment
}

// Transform sets the transformation matrix that will be applied to
// future vertices added to the mesh.
func (mb *MeshGeometryBuilder) Transform(transform sprec.Mat4) {
	mb.transform = transform
}

// VertexOffset returns the index of the first vertex that will be added
// by the next call to Vertex.
func (mb *MeshGeometryBuilder) VertexOffset() uint32 {
	return uint32(len(mb.vertices))
}

// Vertex returns a builder for the next vertex to be added to the mesh.
func (mb *MeshGeometryBuilder) Vertex() VertexBuilder {
	position := uint32(len(mb.vertices))
	mb.vertices = append(mb.vertices, meshBuilderVertex{})
	return VertexBuilder{
		builder: mb,
		vertex:  &mb.vertices[position],
	}
}

// IndexOffset returns the index of the first index that will be added
// by the next call to Index.
func (mb *MeshGeometryBuilder) IndexOffset() uint32 {
	return uint32(len(mb.indices))
}

// Index adds an index to the mesh.
func (mb *MeshGeometryBuilder) Index(index uint32) uint32 {
	position := uint32(len(mb.indices))
	mb.indices = append(mb.indices, index)
	return position
}

// IndexLine adds two indices to the mesh.
func (mb *MeshGeometryBuilder) IndexLine(a, b uint32) (uint32, uint32) {
	position := uint32(len(mb.indices))
	mb.indices = append(mb.indices, a, b)
	return position, position + 2
}

// IndexTriangle adds indices to the mesh to form a triangle.
func (mb *MeshGeometryBuilder) IndexTriangle(a, b, c uint32) (uint32, uint32) {
	position := uint32(len(mb.indices))
	mb.indices = append(mb.indices, a, b, c)
	return position, position + 3
}

// IndexQuad adds indices to the mesh to form a quad based off of two triangles.
func (mb *MeshGeometryBuilder) IndexQuad(a, b, c, d uint32) (uint32, uint32) {
	position := uint32(len(mb.indices))
	mb.indices = append(mb.indices, a, b, c, a, c, d)
	return position, position + 6
}

// Fragment adds a mesh fragment to the mesh.
func (mb *MeshGeometryBuilder) Fragment(topology render.Topology, indexOffset, indexCount uint32) {
	mb.fragments = append(mb.fragments, meshBuilderFragment{
		topology:    topology,
		indexOffset: indexOffset,
		indexCount:  indexCount,
	})
}

// BuildInfo returns the MeshGeometryInfo that has been constructed
// by the builder.
func (mb *MeshGeometryBuilder) BuildInfo() MeshGeometryInfo {
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
		indexFormat render.IndexFormat
		indexData   gblob.LittleEndianBlock
		indexSize   int
	)
	if len(mb.indices) > 0xFFFF {
		indexFormat = render.IndexFormatUnsignedInt
		indexSize = render.SizeU32
		indexData = make(gblob.LittleEndianBlock, indexSize*len(mb.indices))
		for i, index := range mb.indices {
			indexData.SetUint32(i*4, index)
		}
	} else {
		indexFormat = render.IndexFormatUnsignedShort
		indexSize = render.SizeU16
		indexData = make(gblob.LittleEndianBlock, indexSize*len(mb.indices))
		for i, index := range mb.indices {
			indexData.SetUint16(i*2, uint16(index))
		}
	}

	return MeshGeometryInfo{
		VertexFormat: vertexFormat,
		VertexData:   vertexData,
		Fragments: gog.Map(mb.fragments, func(fragment meshBuilderFragment) MeshGeometryFragmentInfo {
			return MeshGeometryFragmentInfo{
				Topology:        fragment.topology,
				IndexByteOffset: fragment.indexOffset * uint32(indexSize),
				IndexCount:      fragment.indexCount,
			}
		}),
		IndexFormat:          indexFormat,
		IndexData:            indexData,
		BoundingSphereRadius: mb.bsRadius(),
	}
}

func (mb *MeshGeometryBuilder) bsRadius() float64 {
	var maxRadius float64
	for _, vertex := range mb.vertices {
		maxRadius = max(maxRadius, float64(vertex.coord.Length()))
	}
	return maxRadius
}

// VertexBuilder is a helper for constructing a vertex for a mesh.
type VertexBuilder struct {
	builder *MeshGeometryBuilder
	vertex  *meshBuilderVertex
}

// CoordVec3 sets the coordinate of the vertex.
func (vb VertexBuilder) CoordVec3(vec sprec.Vec3) VertexBuilder {
	transform := vb.builder.transform
	vb.vertex.coord = sprec.Mat4Vec3Transformation(transform, vec)
	return vb
}

// Coord sets the coordinate of the vertex.
func (vb VertexBuilder) Coord(x, y, z float32) VertexBuilder {
	return vb.CoordVec3(sprec.NewVec3(x, y, z))
}

// NormalVec3 sets the normal of the vertex.
func (vb VertexBuilder) NormalVec3(vec sprec.Vec3) VertexBuilder {
	vb.vertex.normal = vec
	return vb
}

// Normal sets the normal of the vertex.
func (vb VertexBuilder) Normal(x, y, z float32) VertexBuilder {
	return vb.NormalVec3(sprec.NewVec3(x, y, z))
}

// TangentVec3 sets the tangent of the vertex.
func (vb VertexBuilder) TangentVec3(vec sprec.Vec3) VertexBuilder {
	vb.vertex.tangent = vec
	return vb
}

// Tangent sets the tangent of the vertex.
func (vb VertexBuilder) Tangent(x, y, z float32) VertexBuilder {
	return vb.TangentVec3(sprec.NewVec3(x, y, z))
}

// TexCoordVec2 sets the texture coordinate of the vertex.
func (vb VertexBuilder) TexCoordVec2(vec sprec.Vec2) VertexBuilder {
	vb.vertex.texCoord = vec
	return vb
}

// TexCoord sets the texture coordinate of the vertex.
func (vb VertexBuilder) TexCoord(x, y float32) VertexBuilder {
	return vb.TexCoordVec2(sprec.NewVec2(x, y))
}

// ColorVec4 sets the color of the vertex.
func (vb VertexBuilder) ColorVec4(vec sprec.Vec4) VertexBuilder {
	vb.vertex.color = vec
	return vb
}

// Color sets the color of the vertex.
func (vb VertexBuilder) Color(r, g, b, a float32) VertexBuilder {
	return vb.ColorVec4(sprec.NewVec4(r, g, b, a))
}

// Joints sets the joint indices of the vertex.
func (vb VertexBuilder) Joints(a, b, c, d uint8) VertexBuilder {
	vb.vertex.joints = [4]uint8{a, b, c, d}
	return vb
}

// WeightsVec4 sets the joint weights of the vertex.
func (vb VertexBuilder) WeightsVec4(vec sprec.Vec4) VertexBuilder {
	vb.vertex.weights = vec
	return vb
}

// Weights sets the joint weights of the vertex.
func (vb VertexBuilder) Weights(a, b, c, d float32) VertexBuilder {
	return vb.WeightsVec4(sprec.NewVec4(a, b, c, d))
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
	topology    render.Topology
	indexOffset uint32
	indexCount  uint32
}
