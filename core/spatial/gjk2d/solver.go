package gjk2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk2d/internal"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// Solver performs intersection and contact queries between two [Shape]
// values. It holds reusable internal state to avoid allocations, hence it
// is not safe for concurrent use; each goroutine should use its own
// instance.
//
// The zero value is not usable; construct instances through [NewSolver].
type Solver struct {
	gjkSolver *internal.GJKSolver
	epaSolver *internal.EPASolver
}

// NewSolver creates a new [Solver] instance.
func NewSolver() *Solver {
	return &Solver{
		gjkSolver: internal.NewGJKSolver(),
		epaSolver: internal.NewEPASolver(),
	}
}

// Intersect reports whether shapeA and shapeB overlap, accounting for their
// skin radii. It returns false immediately if either shape has no points.
func (s *Solver) Intersect(shapeA, shapeB Shape) bool {
	if len(shapeA.Points) == 0 || len(shapeB.Points) == 0 {
		return false
	}

	shape := internal.MinkowskiShape{
		Source: internal.Polygon{
			Rotation:    shapeA.Rotation,
			InvRotation: shapeA.Rotation.Inverse(),
			Points:      shapeA.Points,
		},
		Target: internal.Polygon{
			Rotation:    shapeB.Rotation,
			InvRotation: shapeB.Rotation.Inverse(),
			Points:      shapeB.Points,
		},
		Offset:     dprec.Vec2Diff(shapeB.Position, shapeA.Position),
		SkinRadius: shapeA.SkinRadius + shapeB.SkinRadius,
	}

	s.gjkSolver.Reset(&shape)
	for s.gjkSolver.Next(&shape) {
		if s.gjkSolver.OverlapsOrigin() {
			return true // break early; we don't need the final simplex
		}
	}
	return s.gjkSolver.OverlapsOrigin()
}

// Resolve reports whether shapeA and shapeB overlap and, if so, returns a
// [shape2d.Contact] describing how to separate them. In the resulting
// contact, shapeA acts as the source shape and shapeB as the target shape.
// The boolean result is false when the shapes do not overlap, in which case
// the contact is meaningless.
func (s *Solver) Resolve(shapeA, shapeB Shape) (shape2d.Contact, bool) {
	if len(shapeA.Points) == 0 || len(shapeB.Points) == 0 {
		return shape2d.Contact{}, false
	}

	shape := internal.MinkowskiShape{
		Source: internal.Polygon{
			Rotation:    shapeA.Rotation,
			InvRotation: shapeA.Rotation.Inverse(),
			Points:      shapeA.Points,
		},
		Target: internal.Polygon{
			Rotation:    shapeB.Rotation,
			InvRotation: shapeB.Rotation.Inverse(),
			Points:      shapeB.Points,
		},
		Offset:     dprec.Vec2Diff(shapeB.Position, shapeA.Position),
		SkinRadius: shapeA.SkinRadius + shapeB.SkinRadius,
	}

	// Run GJK to determine whether the shapes overlap and to produce a simplex.
	s.gjkSolver.Reset(&shape)
	for s.gjkSolver.Next(&shape) {
	}
	if !s.gjkSolver.OverlapsOrigin() {
		return shape2d.Contact{}, false
	}

	// Run EPA to compute the contact information.
	s.epaSolver.Reset(&shape, s.gjkSolver.Simplex(), s.gjkSolver.ContainsOrigin())
	for s.epaSolver.Next(&shape) {
	}

	epaSolution := s.epaSolver.Solution()
	vertexA := epaSolution.VertexA
	vertexB := epaSolution.VertexB
	normal := epaSolution.Normal
	lerp := epaSolution.Lerp
	depth := epaSolution.Depth

	var contactPoint dprec.Vec2
	switch {
	case lerp <= 0.0:
		contactPoint = shapeB.WSPosition(vertexA.Refs.TargetIndex)
	case lerp >= 1.0:
		contactPoint = shapeB.WSPosition(vertexB.Refs.TargetIndex)
	default:
		fromPoint := shapeB.WSPosition(vertexA.Refs.TargetIndex)
		toPoint := shapeB.WSPosition(vertexB.Refs.TargetIndex)
		contactPoint = dprec.Vec2Lerp(fromPoint, toPoint, lerp)
	}
	contactPoint = dprec.Vec2Sum(contactPoint, dprec.Vec2Prod(normal, shapeB.SkinRadius))

	return shape2d.Contact{
		Depth:        depth,
		TargetPoint:  contactPoint,
		TargetNormal: normal,
	}, true
}
