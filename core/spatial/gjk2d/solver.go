package gjk2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk2d/internal"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

type Solver struct {
	gjkSolver *internal.GJKSolver
}

func NewSolver() *Solver {
	return &Solver{
		gjkSolver: internal.NewGJKSolver(),
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
// [shape2d.Contact] describing how to separate them. The contact is expressed
// relative to shapeB as the target shape. The boolean result is false when the
// shapes do not overlap, in which case the contact is meaningless.
func (s *Solver) Resolve(shapeA, shapeB Shape) (shape2d.Contact, bool) {
	// TODO: Add proper implementation using EPA algorithm.
	return shape2d.Contact{
		Depth:        5.0,
		TargetPoint:  dprec.NewVec2(200.0, 200.0),
		TargetNormal: dprec.BasisXVec2(),
	}, s.Intersect(shapeA, shapeB)
}
