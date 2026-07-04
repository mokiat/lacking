package placement2d

import (
	"slices"

	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// MeshInfo contains the information needed to create a mesh shape.
type MeshInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Mesh contains the mesh information.
	Mesh shape2d.Mesh
}

type sceneMeshShape[S any] struct {
	sceneShape[S]
	meshSolver
}

func newMeshSolver(template shape2d.Mesh) meshSolver {
	bc := template.BoundingCircle()
	return meshSolver{
		lsMesh:           template,
		lsBoundingCircle: bc,

		wsMesh:           shape2d.NewMesh(slices.Clone(template.Edges)),
		wsBoundingCircle: bc,
	}
}

type meshSolver struct {
	lsMesh           shape2d.Mesh
	lsBoundingCircle shape2d.Circle

	wsMesh           shape2d.Mesh
	wsBoundingCircle shape2d.Circle
}

func (s *meshSolver) update(transform shape2d.Transform) {
	for i := range s.wsMesh.Edges {
		s.wsMesh.Edges[i] = shape2d.TransformedEdge(
			s.lsMesh.Edges[i],
			transform,
		)
	}
	s.wsBoundingCircle = shape2d.TransformedCircle(s.lsBoundingCircle, transform)
}

func (s *meshSolver) boundingCircle() shape2d.Circle {
	return s.wsBoundingCircle
}
