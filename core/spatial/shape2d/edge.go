package shape2d

import "github.com/mokiat/gomath/dprec"

// Edge represents a directed line segment from A to B in 2D space.
//
// Unlike a [Segment], an Edge is oriented and exposes an [Edge.Normal] that
// points to the right of the A-to-B direction. This makes it suitable for
// describing the boundary edges of a counter-clockwise-wound polygon, whose
// outward normals point away from the interior.
type Edge struct {
	// A is the start of the edge.
	A dprec.Vec2
	// B is the end of the edge.
	B dprec.Vec2
}

// NewEdge creates an [Edge] with the given start and end points.
func NewEdge(a, b dprec.Vec2) Edge {
	return Edge{
		A: a,
		B: b,
	}
}

// TransformedEdge returns a new [Edge] that is the result of applying the
// specified transform to both endpoints of the given edge.
func TransformedEdge(edge Edge, transform Transform) Edge {
	return Edge{
		A: transform.Apply(edge.A),
		B: transform.Apply(edge.B),
	}
}

// Midpoint returns the point halfway between the start and end of the edge.
func (e Edge) Midpoint() dprec.Vec2 {
	return dprec.Vec2Prod(dprec.Vec2Sum(e.A, e.B), 0.5)
}

// Length returns the length of the edge.
func (e Edge) Length() float64 {
	return dprec.Vec2Diff(e.B, e.A).Length()
}

// Normal returns the unit vector perpendicular to the edge, pointing to the
// right of the direction from A to B. For an edge of a counter-clockwise-wound
// polygon this is the outward-facing normal.
func (e Edge) Normal() dprec.Vec2 {
	return dprec.NormalVec2(dprec.Vec2Diff(e.B, e.A))
}

// BoundingCircle returns the smallest [Circle] that fully encompasses the edge.
func (e Edge) BoundingCircle() Circle {
	center := e.Midpoint()
	radius := e.Length() * 0.5
	return Circle{
		Center: center,
		Radius: radius,
	}
}
