package shape3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/mem"
)

type SphereID struct {
	internalID mem.SparseID
}

func (i SphereID) IsNil() bool {
	return i == (SphereID{})
}

type SphereShape struct {
	id           mem.SparseID
	objectID     mem.SparseID
	nextSphereID mem.SparseID
	template     Sphere
	solver       sphereSolver
}

func (s *SphereShape) Init(id mem.SparseID, template Sphere) {
	s.id = id
	s.template = template
}

func (s *SphereShape) Transform(transform Transform) {
	s.solver.Update(s.template, transform)
}

func (s *SphereShape) BoundingSphere() Sphere {
	return s.template
}

type sphereSolver struct {
	position dprec.Vec3
	radius   float64
}

func (s *sphereSolver) Update(template Sphere, transform Transform) {
	s.position = transform.Apply(template.Position)
	s.radius = template.Radius
}
