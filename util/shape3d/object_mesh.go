package shape3d

import (
	"slices"
)

type MeshInfo struct {
	ShapeInfo
	Mesh Mesh
}

type MeshShape struct {
	Shape
	meshSolver
}

func newMeshSolver(template Mesh) meshSolver {
	bs := template.BoundingSphere()
	return meshSolver{
		template:               template,
		templateBoundingSohere: bs,

		wsMesh: Mesh{
			Triangles: slices.Clone(template.Triangles),
		},
		wsBoundingSphere: bs,
	}
}

type meshSolver struct {
	template               Mesh
	templateBoundingSohere Sphere

	wsMesh           Mesh
	wsBoundingSphere Sphere
}

func (s *meshSolver) Update(transform Transform) {
	for i := range s.wsMesh.Triangles {
		srcTriangle := &s.template.Triangles[i]
		s.wsMesh.Triangles[i] = Triangle{
			A: transform.Apply(srcTriangle.A),
			B: transform.Apply(srcTriangle.B),
			C: transform.Apply(srcTriangle.C),
		}
	}
	s.wsBoundingSphere = Sphere{
		Position: transform.Apply(s.templateBoundingSohere.Position),
		Radius:   s.templateBoundingSohere.Radius,
	}
}

func (s *meshSolver) BoundingSphere() Sphere {
	return s.wsBoundingSphere
}
