package shape3d

import "github.com/mokiat/gomath/dprec"

// NewTriangle creates a new Triangle shape.
func NewTriangle(a, b, c dprec.Vec3) Triangle {
	return Triangle{
		A: a,
		B: b,
		C: c,
	}
}

// Triangle represents a tringle in 3D space.
type Triangle struct {

	// A is the first vertex of the triangle.
	A dprec.Vec3

	// B is the second vertex of the triangle.
	B dprec.Vec3

	// C is the third vertex of the triangle.
	C dprec.Vec3
}

// Center returns the center of mass of the triangle.
func (t *Triangle) Center() dprec.Vec3 {
	return dprec.Vec3{
		X: (t.A.X + t.B.X + t.C.X) / 3.0,
		Y: (t.A.Y + t.B.Y + t.C.Y) / 3.0,
		Z: (t.A.Z + t.B.Z + t.C.Z) / 3.0,
	}
}

// Normal returns the orientation of the triangle's surface.
func (t *Triangle) Normal() dprec.Vec3 {
	vecAB := dprec.Vec3Diff(t.B, t.A)
	vecAC := dprec.Vec3Diff(t.C, t.A)
	return dprec.UnitVec3(dprec.Vec3Cross(vecAB, vecAC))
}

// Area returns the triangle's surface area.
func (t *Triangle) Area() float64 {
	vecAB := dprec.Vec3Diff(t.B, t.A)
	vecAC := dprec.Vec3Diff(t.C, t.A)
	return dprec.Vec3Cross(vecAB, vecAC).Length() / 2.0
}

// IsLookingTowards checks whether the orientation of the triangle looks
// towards the same hemisphere as the provided direction.
func (t *Triangle) IsLookingTowards(direction dprec.Vec3) bool {
	return dprec.Vec3Dot(t.Normal(), direction) > 0.0
}

// ContainsPoint checks whether the specified Point is inside the triangle.
//
// Beware, currently this method assumes that the point lies somewhere on the
// surface plane of the triangle.
func (t *Triangle) ContainsPoint(point dprec.Vec3) bool {
	normal := t.Normal()
	if triangleABP := NewTriangle(t.A, t.B, point); !triangleABP.IsLookingTowards(normal) {
		return false
	}
	if triangleBCP := NewTriangle(t.B, t.C, point); !triangleBCP.IsLookingTowards(normal) {
		return false
	}
	if triangleCAP := NewTriangle(t.C, t.A, point); !triangleCAP.IsLookingTowards(normal) {
		return false
	}
	return true
}

// BoundingSphere returns a Sphere shape that encompases this triangle.
func (t *Triangle) BoundingSphere() Sphere {
	center := t.Center()
	radius := 0.0
	if lng := dprec.Vec3Diff(t.A, center).Length(); lng > radius {
		radius = lng
	}
	if lng := dprec.Vec3Diff(t.B, center).Length(); lng > radius {
		radius = lng
	}
	if lng := dprec.Vec3Diff(t.C, center).Length(); lng > radius {
		radius = lng
	}
	return NewSphere(center, radius)
}
