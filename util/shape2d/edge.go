package shape2d

import "github.com/mokiat/gomath/dprec"

// NewEdge creates a new Edge from points a to b.
func NewEdge(a, b dprec.Vec2) Edge {
	return Edge{
		A: a,
		B: b,
	}
}

// TransformedEdge returns a new Edge that is the result of applying the given
// transform to the provided edge.
func TransformedEdge(edge Edge, transform Transform) Edge {
	return Edge{
		A: transform.Apply(edge.A),
		B: transform.Apply(edge.B),
	}
}

// Edge represents an edge in 2D space.
type Edge struct {

	// A is the starting point of the edge.
	A dprec.Vec2

	// B is the ending point of the edge.
	B dprec.Vec2
}

// Length computes and returns the length of the edge.
func (e Edge) Length() float64 {
	return dprec.Vec2Diff(e.B, e.A).Length()
}

// Normal computes and returns the normal vector of the edge.
//
// It assumes a counter-clockwise winding order.
func (e Edge) Normal() dprec.Vec2 {
	return dprec.NormalVec2(dprec.Vec2Diff(e.A, e.B))
}

// Center returns the center point of the segment.
func (e Edge) Center() dprec.Vec2 {
	return dprec.Vec2Prod(dprec.Vec2Sum(e.A, e.B), 0.5)
}

// BoundingCircle returns a Circle that encompasses this segment.
func (e Edge) BoundingCircle() Circle {
	center := e.Center()
	length := e.Length()
	return NewCircle(center, length*0.5)
}
