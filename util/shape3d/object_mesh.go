package shape3d

import (
	"slices"

	"github.com/mokiat/lacking/util/mem"
)

type MeshID struct {
	internalID mem.SparseID
}

func (i MeshID) IsNil() bool {
	return i == (MeshID{})
}

type MeshShape struct {
	template Mesh
	bs       Sphere
	solver   meshSolver
}

func (s *MeshShape) Init(template Mesh, transform Transform) {
	s.template = template
	s.bs = template.BoundingSphere()
	s.solver.wsTriangles = slices.Clone(template.Triangles)
	s.solver.Update(template, s.bs, transform)
}

func (s *MeshShape) SetTransform(transform Transform) {
	s.solver.Update(s.template, s.bs, transform)
}

func (s *MeshShape) BoundingSphere() Sphere {
	return s.bs
}

type meshSolver struct {
	wsBoundingSphere Sphere
	wsTriangles      []Triangle
}

func (s *meshSolver) Update(template Mesh, meshBS Sphere, transform Transform) {
	s.wsBoundingSphere = Sphere{
		Position: transform.Apply(meshBS.Position),
		Radius:   meshBS.Radius,
	}
	for i := range s.wsTriangles {
		srcTriangle := &template.Triangles[i]
		s.wsTriangles[i] = Triangle{
			A: transform.Apply(srcTriangle.A),
			B: transform.Apply(srcTriangle.B),
			C: transform.Apply(srcTriangle.C),
		}
	}
}
