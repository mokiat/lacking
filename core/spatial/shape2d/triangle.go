package shape2d

import "github.com/mokiat/gomath/dprec"

// Triangle represents a triangle in 2D space, defined by its three vertices.
//
// The winding of the vertices (counter-clockwise or clockwise) is not fixed and
// can be inspected with [Triangle.SignedArea] and [Triangle.IsCCW]. Some
// methods, such as [Triangle.ContainsPoint], expect a counter-clockwise winding.
type Triangle struct {
	// A is the first vertex of the triangle.
	A dprec.Vec2
	// B is the second vertex of the triangle.
	B dprec.Vec2
	// C is the third vertex of the triangle.
	C dprec.Vec2
}

// NewTriangle creates a [Triangle] from the given vertices.
func NewTriangle(a, b, c dprec.Vec2) Triangle {
	return Triangle{
		A: a,
		B: b,
		C: c,
	}
}

// TransformedTriangle returns a new [Triangle] that is the result of applying
// the specified transform to each of the vertices of the given triangle.
func TransformedTriangle(triangle Triangle, transform Transform) Triangle {
	return Triangle{
		A: transform.Apply(triangle.A),
		B: transform.Apply(triangle.B),
		C: transform.Apply(triangle.C),
	}
}

// Centroid returns the centroid of the triangle, which is the average of its
// three vertices.
func (t Triangle) Centroid() dprec.Vec2 {
	return dprec.Vec2{
		X: (t.A.X + t.B.X + t.C.X) / 3.0,
		Y: (t.A.Y + t.B.Y + t.C.Y) / 3.0,
	}
}

// LengthAB returns the length of the edge from vertex A to vertex B.
func (t Triangle) LengthAB() float64 {
	return dprec.Vec2Diff(t.B, t.A).Length()
}

// LengthBC returns the length of the edge from vertex B to vertex C.
func (t Triangle) LengthBC() float64 {
	return dprec.Vec2Diff(t.C, t.B).Length()
}

// LengthCA returns the length of the edge from vertex C to vertex A.
func (t Triangle) LengthCA() float64 {
	return dprec.Vec2Diff(t.A, t.C).Length()
}

// SignedArea returns the signed area of the triangle. It is positive when the
// vertices are wound counter-clockwise, negative when they are wound clockwise,
// and zero when the vertices are collinear.
func (t Triangle) SignedArea() float64 {
	vecAB := dprec.Vec2Diff(t.B, t.A)
	vecAC := dprec.Vec2Diff(t.C, t.A)
	cross := dprec.Vec2Cross(vecAB, vecAC)
	return cross / 2.0
}

// Area returns the unsigned surface area of the triangle.
func (t Triangle) Area() float64 {
	return dprec.Abs(t.SignedArea())
}

// IsCCW returns whether the triangle's vertices are wound counter-clockwise,
// that is whether its signed area is positive.
func (t Triangle) IsCCW() bool {
	return t.SignedArea() > 0.0
}

// ContainsPoint returns whether the given point lies within the triangle,
// including on its edges and vertices.
//
// The triangle is expected to be wound counter-clockwise. For a clockwise-wound
// triangle (one whose [Triangle.IsCCW] reports false) this always returns false.
func (t Triangle) ContainsPoint(p dprec.Vec2) bool {
	vecAB := dprec.Vec2Diff(t.B, t.A)
	vecAC := dprec.Vec2Diff(t.C, t.A)

	det := dprec.Vec2Cross(vecAB, vecAC)
	if det < 0.0 {
		return false
	}

	offset := dprec.Vec2Diff(p, t.A)

	scaledU := dprec.Vec2Cross(offset, vecAC)
	if scaledU < 0.0 || scaledU > det {
		return false
	}

	scaledV := dprec.Vec2Cross(vecAB, offset)
	if scaledV < 0.0 || (scaledU+scaledV) > det {
		return false
	}

	return true
}

// BoundingCircle returns the smallest [Circle] that is centered at the
// triangle's centroid and fully encompasses the triangle.
func (t Triangle) BoundingCircle() Circle {
	center := t.Centroid()
	radius := dprec.Sqrt(max(
		dprec.Vec2Diff(t.A, center).SqrLength(),
		dprec.Vec2Diff(t.B, center).SqrLength(),
		dprec.Vec2Diff(t.C, center).SqrLength(),
	))
	return Circle{
		Center: center,
		Radius: radius,
	}
}
