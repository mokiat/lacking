package shape2d

import "github.com/mokiat/gomath/dprec"

// NewTriangle creates a new Triangle instance given three vertices a, b, and c.
func NewTriangle(a, b, c dprec.Vec2) Triangle {
	return Triangle{
		A: a,
		B: b,
		C: c,
	}
}

// Triangle represents a triangle in 2D space defined by three vertices.
//
// The ordering of the vertices is significant and should follow a
// counter-clockwise direction to ensure correct orientation.
type Triangle struct {

	// A is the first vertex of the triangle.
	A dprec.Vec2

	// B is the second vertex of the triangle.
	B dprec.Vec2

	// C is the third vertex of the triangle.
	C dprec.Vec2
}

// Centroid computes and returns the centroid of the triangle.
func (t Triangle) Centroid() dprec.Vec2 {
	return dprec.Vec2{
		X: (t.A.X + t.B.X + t.C.X) / 3.0,
		Y: (t.A.Y + t.B.Y + t.C.Y) / 3.0,
	}
}

// LengthAB computes and returns the length of the edge between vertices
// A and B.
func (t Triangle) LengthAB() float64 {
	return dprec.Vec2Diff(t.B, t.A).Length()
}

// LengthBC computes and returns the length of the edge between vertices
// B and C.
func (t Triangle) LengthBC() float64 {
	return dprec.Vec2Diff(t.C, t.B).Length()
}

// LengthCA computes and returns the length of the edge between vertices
// C and A.
func (t Triangle) LengthCA() float64 {
	return dprec.Vec2Diff(t.A, t.C).Length()
}

// IsCCW checks if the vertices of the triangle are ordered in a
// counter-clockwise manner.
func (t Triangle) IsCCW() bool {
	vecAB := dprec.Vec2Diff(t.B, t.A)
	vecAC := dprec.Vec2Diff(t.C, t.A)
	cross := dprec.Vec2Cross(vecAB, vecAC)
	return cross > 0.0
}

// ContainsPoint checks if the given point p lies within the triangle t.
func (t Triangle) ContainsPoint(p dprec.Vec2) bool {
	vecAB := dprec.Vec2Diff(t.B, t.A)
	vecAC := dprec.Vec2Diff(t.C, t.A)

	det := dprec.Vec2Cross(vecAB, vecAC)
	if det < 0.00001 {
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
