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
	sourcePosition := s.Source.WSPosition(refs.SourceIndex)
	targetPosition := s.Target.WSPosition(refs.TargetIndex)
	return MinkowskiVertex{
		Position: dprec.Vec2Sum(s.Offset, dprec.Vec2Diff(targetPosition, sourcePosition)),
		Refs:     refs,
	}
}

// VertexNormal returns an outward unit normal of the Minkowski difference at
// the given boundary vertex. It is used when the origin coincides with the
// vertex, where the plain origin-to-vertex direction is undefined.
//
// The vertex lies on the boundary of the convex difference, so an outward
// normal n is any direction along which the vertex is a supporting point,
// that is dot(n, p - vertex) <= 0 for every other difference vertex p. Such a
// normal is perpendicular to one of the hull edges incident to the vertex, so
// we take the right-hand normal of the edge toward each other vertex and keep
// the one under which the difference protrudes the least. For a genuine hull
// edge nothing protrudes past it, so that protrusion is (near) zero, which
// makes the search robust even when several vertices coincide with the origin
// or lie collinear along the same edge. It returns false only when every
// other vertex coincides with the given one (a fully degenerate difference).
func (s *MinkowskiShape) VertexNormal(vertex MinkowskiVertex) dprec.Vec2 {
	var (
		bestNormal   = dprec.BasisXVec2()
		bestProtrude = math.MaxFloat64
	)
	for i := range s.Source.Points {
		for j := range s.Target.Points {
			refs := RefPair{SourceIndex: i, TargetIndex: j}
			if refs == vertex.Refs {
				continue
			}
			edge := dprec.Vec2Diff(s.Vertex(refs).Position, vertex.Position)
			if edge.SqrLength() < 1e-12 {
				continue // coincident vertex yields no usable edge
			}
			normal := dprec.UnitVec2(transposeVec2(edge))
			if protrude := s.maxProtrusion(vertex, normal); protrude < bestProtrude {
				bestProtrude = protrude
				bestNormal = normal
			}
		}
	}
	return bestNormal
}

// maxProtrusion returns the greatest signed distance by which any vertex of
// the difference extends past the supporting line through vertex along normal.
// A value at or near zero means normal is an outward normal at vertex.
func (s *MinkowskiShape) maxProtrusion(vertex MinkowskiVertex, normal dprec.Vec2) float64 {
	protrusion := 0.0 // the vertex itself lies on the line
	for i := range s.Source.Points {
		for j := range s.Target.Points {
			point := s.Vertex(RefPair{SourceIndex: i, TargetIndex: j}).Position
			protrusion = max(protrusion, dprec.Vec2Dot(normal, dprec.Vec2Diff(point, vertex.Position)))
		}
	}
	return protrusion
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
