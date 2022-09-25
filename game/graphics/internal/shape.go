package internal

import "github.com/mokiat/lacking/render"

const (
	CoordAttributeIndex    = 0
	NormalAttributeIndex   = 1
	TangentAttributeIndex  = 2
	TexCoordAttributeIndex = 3
	ColorAttributeIndex    = 4
	WeightsAttributeIndex  = 5
	JointsAttributeIndex   = 6
)

// Shape defines a simple 3D mesh that does not have any materials.
type Shape struct {
	vertexBuffer render.Buffer
	indexBuffer  render.Buffer

	vertexArray render.VertexArray

	topology   render.Topology
	indexCount int
}

// VertexArray returns the VertexArray that contains the vertices and indices
// for this Shape.
func (s *Shape) VertexArray() render.VertexArray {
	return s.vertexArray
}

// Topology returns the mesh topology of this Shape.
func (s *Shape) Topology() render.Topology {
	return s.topology
}

// IndexCount returns the number of indices that comprise this Shape.
func (s *Shape) IndexCount() int {
	return s.indexCount
}

// Release releases all resources allocated by this Shape.
func (s *Shape) Release() {
	defer s.vertexBuffer.Release()
	defer s.indexBuffer.Release()

	defer s.vertexArray.Release()
}
