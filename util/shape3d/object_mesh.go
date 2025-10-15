package shape3d

import "slices"

// MeshInfo contains the information needed to create a mesh shape.
type MeshInfo[S any] struct {

	// ShapeInfo contains general shape information.
	ShapeInfo[S]

	// Mesh contains the mesh information.
	Mesh Mesh
}

type sceneMeshShape[S any] struct {
	sceneShape[S]
	meshSolver
}

func newMeshSolver(template Mesh) meshSolver {
	bs := template.BoundingSphere()
	return meshSolver{
		lsMesh:           template,
		lsBoundingSphere: bs,

		wsMesh: Mesh{
			Triangles: slices.Clone(template.Triangles),
		},
		wsBoundingSphere: bs,
	}
}

type meshSolver struct {
	lsMesh           Mesh
	lsBoundingSphere Sphere

	wsMesh           Mesh
	wsBoundingSphere Sphere
}

func (s *meshSolver) update(transform Transform) {
	for i := range s.wsMesh.Triangles {
		srcTriangle := &s.lsMesh.Triangles[i]
		s.wsMesh.Triangles[i] = Triangle{
			A: transform.Apply(srcTriangle.A),
			B: transform.Apply(srcTriangle.B),
			C: transform.Apply(srcTriangle.C),
		}
	}
	s.wsBoundingSphere = TransformedSphere(s.lsBoundingSphere, transform)
}

func (s *meshSolver) boundingSphere() Sphere {
	return s.wsBoundingSphere
}
