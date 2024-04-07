package graphics

import "github.com/mokiat/lacking/render"

// MeshGeometryInfo contains everything needed to create a new MeshGeometry.
type MeshGeometryInfo struct {
	VertexData           []byte
	VertexFormat         VertexFormat
	IndexData            []byte
	IndexFormat          render.IndexFormat
	Fragments            []MeshGeometryFragmentInfo
	BoundingSphereRadius float64
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
	vertexBuffer         render.Buffer
	indexBuffer          render.Buffer
	vertexArray          render.VertexArray
	vertexFormat         VertexFormat
	fragments            []MeshGeometryFragment
	boundingSphereRadius float64
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
	defer g.vertexBuffer.Release()
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

// VertexFormat describes the data that is contained in a single vertex.
type VertexFormat struct {
	// TODO:
	// Stride uint32

	// TODO: VertexAttribute { enabled, offset, stride } // rethink stride as well
	//
	// Coord 				opt.T[VertexAttribute]

	HasCoord            bool
	CoordOffsetBytes    uint32
	CoordStrideBytes    uint32
	HasNormal           bool
	NormalOffsetBytes   uint32
	NormalStrideBytes   uint32
	HasTangent          bool
	TangentOffsetBytes  uint32
	TangentStrideBytes  uint32
	HasTexCoord         bool
	TexCoordOffsetBytes uint32
	TexCoordStrideBytes uint32
	HasColor            bool
	ColorOffsetBytes    uint32
	ColorStrideBytes    uint32
	HasWeights          bool
	WeightsOffsetBytes  uint32
	WeightsStrideBytes  uint32
	HasJoints           bool
	JointsOffsetBytes   uint32
	JointsStrideBytes   uint32
}
