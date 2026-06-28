package placement3d

import (
	"slices"

	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// MeshInfo contains the information needed to create a mesh shape.
type MeshInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Mesh contains the mesh information.
	Mesh shape3d.Mesh
}

type sceneMeshShape[S any] struct {
	sceneShape[S]
	meshSolver
}

func newMeshSolver(template shape3d.Mesh) meshSolver {
	bs := template.BoundingSphere()
	return meshSolver{
		lsMesh:           template,
		lsBoundingSphere: bs,

		wsMesh: shape3d.Mesh{
			Triangles: slices.Clone(template.Triangles),
		},
		wsBoundingSphere: bs,
	}
}

type meshSolver struct {
	lsMesh           shape3d.Mesh
	lsBoundingSphere shape3d.Sphere

	wsMesh           shape3d.Mesh
	wsBoundingSphere shape3d.Sphere
}

func (s *meshSolver) update(transform shape3d.Transform) {
	for i := range s.wsMesh.Triangles {
		s.wsMesh.Triangles[i] = shape3d.TransformedTriangle(
			s.lsMesh.Triangles[i],
			transform,
		)
	}
	s.wsBoundingSphere = shape3d.TransformedSphere(s.lsBoundingSphere, transform)
}

func (s *meshSolver) boundingSphere() shape3d.Sphere {
	return s.wsBoundingSphere
}
