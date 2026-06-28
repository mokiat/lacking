package shape3d

import "github.com/mokiat/gomath/dprec"

// Triangle represents a triangle in 3D space. Its three vertices are expected
// to be specified in counter-clockwise order when viewed from the side the
// normal faces.
type Triangle struct {
	// A is the first vertex of the triangle.
	A dprec.Vec3
	// B is the second vertex of the triangle.
	B dprec.Vec3
	// C is the third vertex of the triangle.
	C dprec.Vec3
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
func (t Triangle) Centroid() dprec.Vec3 {
	return dprec.Vec3{
		X: (t.A.X + t.B.X + t.C.X) / 3.0,
		Y: (t.A.Y + t.B.Y + t.C.Y) / 3.0,
		Z: (t.A.Z + t.B.Z + t.C.Z) / 3.0,
	}
}

// Normal returns the unit vector that is perpendicular to the triangle's
// surface. It points to the side from which the vertices appear in
// counter-clockwise order.
func (t Triangle) Normal() dprec.Vec3 {
	vecAB := dprec.Vec3Diff(t.B, t.A)
	vecAC := dprec.Vec3Diff(t.C, t.A)
	return dprec.UnitVec3(dprec.Vec3Cross(vecAB, vecAC))
}

// Area returns the surface area of the triangle.
func (t Triangle) Area() float64 {
	vecAB := dprec.Vec3Diff(t.B, t.A)
	vecAC := dprec.Vec3Diff(t.C, t.A)
	return dprec.Vec3Cross(vecAB, vecAC).Length() / 2.0
}

// FacesTowards returns whether the triangle's normal points into the same
// hemisphere as the specified direction, that is whether the angle between the
// normal and the direction is strictly less than 90 degrees.
func (t Triangle) FacesTowards(direction dprec.Vec3) bool {
	return dprec.Vec3Dot(t.Normal(), direction) > 0.0
}

// BoundingSphere returns the smallest [Sphere] that is centered at the
// triangle's centroid and fully encompasses the triangle.
func (t Triangle) BoundingSphere() Sphere {
	center := t.Centroid()
	radius := max(
		dprec.Vec3Diff(t.A, center).Length(),
		dprec.Vec3Diff(t.B, center).Length(),
		dprec.Vec3Diff(t.C, center).Length(),
	)
	return Sphere{
		Center: center,
		Radius: radius,
	}
}
