package internal

import "github.com/mokiat/gomath/dprec"

// MinkowskiShape represents the Minkowski difference (Target - Source) of
// two convex polygons. The two shapes overlap exactly when the difference
// contains the origin (or, with skin radii, when the origin is within
// SkinRadius distance of the difference).
type MinkowskiShape struct {
	// Source is the polygon that is subtracted in the difference.
	Source Polygon
	// Target is the polygon that is added in the difference.
	Target Polygon
	// Offset is the position of Target relative to Source in world space.
	Offset dprec.Vec2
	// SkinRadius is the combined skin radius of the two shapes.
	SkinRadius float64
}

func (s *MinkowskiShape) MaxIterations() int {
	// The Minkowski difference of two convex polygons has at most m+n
	// distinct support results, so an iteration budget of m+n would suffice
	// if every iteration discovered a new vertex. However, when the solver
	// collapses the simplex back to a single point (see the vertex-region
	// case in appendToEdge), a previously discarded vertex may be revisited,
	// costing an extra iteration. Doubling the budget covers such revisits.
	return 2 * (len(s.Source.Points) + len(s.Target.Points))
}

// Support returns the vertex of the Minkowski difference that is furthest
// along dir. The direction does not need to be normalized.
func (s *MinkowskiShape) Support(dir dprec.Vec2) MinkowskiVertex {
	sourcePosition, sourceIndex := s.Source.Support(dprec.InverseVec2(dir))
	targetPosition, targetIndex := s.Target.Support(dir)
	return MinkowskiVertex{
		Position: dprec.Vec2Sum(s.Offset, dprec.Vec2Diff(targetPosition, sourcePosition)),
		Refs: RefPair{
			SourceIndex: sourceIndex,
			TargetIndex: targetIndex,
		},
	}
}

// MinkowskiVertex is a point on the boundary of the Minkowski difference,
// together with the source and target vertices that produced it.
type MinkowskiVertex struct {
	Position dprec.Vec2
	Refs     RefPair
}

// RefPair identifies the pair of polygon vertices that produced a Minkowski
// vertex. Two Minkowski vertices are the same exactly when their ref pairs
// match, which allows cheap and float-safe identity comparisons.
type RefPair struct {
	SourceIndex int
	TargetIndex int
}
