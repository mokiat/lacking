package internal

import (
	"math"

	"github.com/mokiat/gomath/dprec"
)

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

// MaxIterations returns an iteration budget that is sufficient for the GJK
// solver to converge on this shape.
//
// The Minkowski difference of two convex polygons has at most m+n distinct
// support results, so an iteration budget of m+n would suffice if every
// iteration discovered a new vertex. However, when the solver downgrades
// the simplex back to a single point, a previously discarded vertex may be
// revisited, costing an extra iteration. Doubling the budget covers such
// revisits.
func (s *MinkowskiShape) MaxIterations() int {
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

func (s *MinkowskiShape) Vertex(refs RefPair) MinkowskiVertex {
	sourcePosition := s.Source.Points[refs.SourceIndex]
	targetPosition := s.Target.Points[refs.TargetIndex]
	return MinkowskiVertex{
		Position: dprec.Vec2Sum(s.Offset, dprec.Vec2Diff(targetPosition, sourcePosition)),
		Refs:     refs,
	}
}

func (s *MinkowskiShape) FurthestVertex() MinkowskiVertex {
	maxDistance := -math.MaxFloat64
	var furthestVertex MinkowskiVertex
	for i := range s.Source.Points {
		for j := range s.Target.Points {
			vertex := s.Vertex(RefPair{
				SourceIndex: i,
				TargetIndex: j,
			})
			distance := vertex.Position.SqrLength()
			if distance > maxDistance {
				maxDistance = distance
				furthestVertex = vertex
			}
		}
	}
	return furthestVertex
}

func (s *MinkowskiShape) VertexCount() int {
	return len(s.Source.Points) * len(s.Target.Points)
}

// MinkowskiVertex is a point on the boundary of the Minkowski difference,
// together with the source and target vertices that produced it.
type MinkowskiVertex struct {
	// Position is the location of the vertex within the Minkowski difference.
	Position dprec.Vec2
	// Refs identifies the source and target polygon vertices that produced
	// this vertex.
	Refs RefPair
}

// RefPair identifies the pair of polygon vertices that produced a Minkowski
// vertex. Two Minkowski vertices are the same exactly when their ref pairs
// match, which allows cheap and float-safe identity comparisons.
type RefPair struct {
	SourceIndex int
	TargetIndex int
}
