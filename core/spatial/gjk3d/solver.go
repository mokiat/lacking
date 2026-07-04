package gjk3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk3d/internal"
	"github.com/mokiat/lacking/core/spatial/shape3d"
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
		Source: internal.Hull{
			Rotation:    shapeA.Rotation,
			InvRotation: shapeA.Rotation.Inverse(),
			Points:      shapeA.Points,
		},
		Target: internal.Hull{
			Rotation:    shapeB.Rotation,
			InvRotation: shapeB.Rotation.Inverse(),
			Points:      shapeB.Points,
		},
		Offset:     dprec.Vec3Diff(shapeB.Position, shapeA.Position),
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
// [shape3d.Contact] describing how to separate them. In the resulting
// contact, shapeA acts as the source shape and shapeB as the target shape.
// The boolean result is false when the shapes do not overlap, in which case
// the contact is meaningless.
func (s *Solver) Resolve(shapeA, shapeB Shape) (shape3d.Contact, bool) {
	if len(shapeA.Points) == 0 || len(shapeB.Points) == 0 {
		return shape3d.Contact{}, false
	}

	shape := internal.MinkowskiShape{
		Source: internal.Hull{
			Rotation:    shapeA.Rotation,
			InvRotation: shapeA.Rotation.Inverse(),
			Points:      shapeA.Points,
		},
		Target: internal.Hull{
			Rotation:    shapeB.Rotation,
			InvRotation: shapeB.Rotation.Inverse(),
			Points:      shapeB.Points,
		},
		Offset:     dprec.Vec3Diff(shapeB.Position, shapeA.Position),
		SkinRadius: shapeA.SkinRadius + shapeB.SkinRadius,
	}

	// Run GJK to determine whether the shapes overlap and to produce a simplex.
	s.gjkSolver.Reset(&shape)
	for s.gjkSolver.Next(&shape) {
	}
	if !s.gjkSolver.OverlapsOrigin() {
		return shape3d.Contact{}, false
	}

	// Run EPA to compute the contact information.
	s.epaSolver.Reset(&shape, s.gjkSolver.Simplex(), s.gjkSolver.ContainsOrigin())
	for s.epaSolver.Next(&shape) {
	}

	epaSolution := s.epaSolver.Solution()
	normal := epaSolution.Normal
	depth := epaSolution.Depth

	// The contact point is the closest feature evaluated on the target shape,
	// interpolated through the barycentric weights of the closest point.
	pointA := shapeB.WSPosition(epaSolution.VertexA.Refs.TargetIndex)
	pointB := shapeB.WSPosition(epaSolution.VertexB.Refs.TargetIndex)
	pointC := shapeB.WSPosition(epaSolution.VertexC.Refs.TargetIndex)
	contactPoint := dprec.Vec3MultiSum(
		dprec.Vec3Prod(pointA, epaSolution.BaryA),
		dprec.Vec3Prod(pointB, epaSolution.BaryB),
		dprec.Vec3Prod(pointC, epaSolution.BaryC),
	)
	contactPoint = dprec.Vec3Sum(contactPoint, dprec.Vec3Prod(normal, shapeB.SkinRadius))

	return shape3d.Contact{
		Depth:        depth,
		TargetPoint:  contactPoint,
		TargetNormal: normal,
	}, true
}
