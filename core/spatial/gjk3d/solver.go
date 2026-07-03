package gjk3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk3d/internal"
)

// Solver performs intersection queries between two [Shape] values. It holds
// reusable internal state to avoid allocations, hence it is not safe for
// concurrent use; each goroutine should use its own instance.
//
// The zero value is not usable; construct instances through [NewSolver].
type Solver struct {
	gjkSolver *internal.GJKSolver
}

// NewSolver creates a new [Solver] instance.
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
