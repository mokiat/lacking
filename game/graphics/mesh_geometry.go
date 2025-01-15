package graphics

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/render"
)

// MeshGeometryInfo contains everything needed to create a new MeshGeometry.
type MeshGeometryInfo struct {
	VertexBuffers        []MeshGeometryVertexBuffer
	VertexFormat         MeshGeometryVertexFormat
	IndexBuffer          MeshGeometryIndexBuffer
	Fragments            []MeshGeometryFragmentInfo
	BoundingSphereRadius float64
	MinDistance          opt.T[float64]
	MaxDistance          opt.T[float64]
	MaxCascade           opt.T[uint8]
}

// MeshGeometryFragmentInfo contains the information needed to represent a
// fragment of a mesh.
type MeshGeometryFragmentInfo struct {
	Name            string
	Topology        render.Topology
	IndexByteOffset uint32
	IndexCount      uint32
}

// MeshGeometry represents the raw geometry of a mesh, without any materials
// or shading.
type MeshGeometry struct {
	vertexBuffers        []render.Buffer
	indexBuffer          render.Buffer
	vertexArray          render.VertexArray
	vertexFormat         MeshGeometryVertexFormat
	fragments            []MeshGeometryFragment
	boundingSphereRadius float64
	minDistance          float64
	maxDistance          float64
	maxCascade           uint8
}

// FragmentCount returns the number of fragments that make up this mesh.
//
// Each fragment represents a portion of the mesh that is drawn with a specific
// material and topology, through the exact material is not specified here.
func (g *MeshGeometry) FragmentCount() int {
	return len(g.fragments)
}

// Fragment returns the fragment at the specified index.
func (g *MeshGeometry) Fragment(index int) *MeshGeometryFragment {
	return &g.fragments[index]
}

// Delete releases the resources that are associated with this mesh geometry.
func (g *MeshGeometry) Delete() {
	for _, buffer := range g.vertexBuffers {
		defer buffer.Release()
	}
	defer g.indexBuffer.Release()
	defer g.vertexArray.Release()
}

// MeshGeometryFragment represents a portion of a mesh that is drawn with a
// specific material and topology.
type MeshGeometryFragment struct {
	name            string
	topology        render.Topology
	indexByteOffset uint32
	indexCount      uint32
}

// Name returns the name of the fragment.
func (g *MeshGeometryFragment) Name() string {
	return g.name
}

// Topology returns the topology that is used to draw the fragment.
func (g *MeshGeometryFragment) Topology() render.Topology {
	return g.topology
}

// MeshGeometryVertexBuffer represents a buffer that contains vertex data.
type MeshGeometryVertexBuffer struct {
	ByteStride uint32
	Data       []byte
}

// MeshGeometryVertexFormat describes the data that is contained in a single vertex.
type MeshGeometryVertexFormat struct {
	Coord    opt.T[MeshGeometryVertexAttribute]
	Normal   opt.T[MeshGeometryVertexAttribute]
	Tangent  opt.T[MeshGeometryVertexAttribute]
	TexCoord opt.T[MeshGeometryVertexAttribute]
	Color    opt.T[MeshGeometryVertexAttribute]
	Weights  opt.T[MeshGeometryVertexAttribute]
	Joints   opt.T[MeshGeometryVertexAttribute]
}

// MeshGeometryVertexAttribute describes a single attribute of a vertex.
type MeshGeometryVertexAttribute struct {
	BufferIndex uint32
	ByteOffset  uint32
	Format      render.VertexAttributeFormat
}

// MeshGeometryIndexBuffer represents a buffer that contains index data.
type MeshGeometryIndexBuffer struct {
	Data   []byte
	Format render.IndexFormat
}
