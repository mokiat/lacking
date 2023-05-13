package collision

import "github.com/mokiat/gomath/dprec"

// NewTriangle creates a new Triangle shape.
func NewTriangle(a, b, c dprec.Vec3) Triangle {
	return Triangle{
		a: a,
		b: b,
		c: c,
	}
}

// Triangle represents a tringle in 3D space.
type Triangle struct {
	a dprec.Vec3
	b dprec.Vec3
	c dprec.Vec3
}

// Replace replaces this shape with the template one after the specified
// transformation has been applied to it.
func (t *Triangle) Replace(template Triangle, transform Transform) {
	t.a = transform.Vector(template.a)
	t.b = transform.Vector(template.b)
	t.c = transform.Vector(template.c)
}

// A returns the first corner of the triangle.
func (t *Triangle) A() dprec.Vec3 {
	return t.a
}

// B returns the second corner of the triangle.
func (t *Triangle) B() dprec.Vec3 {
	return t.b
}

// C returns the third corner of the triangle.
func (t *Triangle) C() dprec.Vec3 {
	return t.c
}

// Center returns the center of mass of the triangle.
func (t *Triangle) Center() dprec.Vec3 {
	return dprec.Vec3{
		X: (t.a.X + t.b.X + t.c.X) / 3.0,
		Y: (t.a.Y + t.b.Y + t.c.Y) / 3.0,
		Z: (t.a.Z + t.b.Z + t.c.Z) / 3.0,
	}
}

// Normal returns the orientation of the triangle's surface.
func (t *Triangle) Normal() dprec.Vec3 {
	vecAB := dprec.Vec3Diff(t.b, t.a)
	vecAC := dprec.Vec3Diff(t.c, t.a)
	return dprec.UnitVec3(dprec.Vec3Cross(vecAB, vecAC))
}

// Area returns the triangle's surface area.
func (t *Triangle) Area() float64 {
	vecAB := dprec.Vec3Diff(t.b, t.a)
	vecAC := dprec.Vec3Diff(t.c, t.a)
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
	if triangleABP := NewTriangle(t.a, t.b, point); !triangleABP.IsLookingTowards(normal) {
		return false
	}
	if triangleBCP := NewTriangle(t.b, t.c, point); !triangleBCP.IsLookingTowards(normal) {
		return false
	}
	if triangleCAP := NewTriangle(t.c, t.a, point); !triangleCAP.IsLookingTowards(normal) {
		return false
	}
	return true
}

// BoundingSphere returns a Sphere shape that encompases this triangle.
func (t *Triangle) BoundingSphere() Sphere {
	center := t.Center()
	radius := 0.0
	if lng := dprec.Vec3Diff(t.a, center).Length(); lng > radius {
		radius = lng
	}
	if lng := dprec.Vec3Diff(t.b, center).Length(); lng > radius {
		radius = lng
	}
	if lng := dprec.Vec3Diff(t.c, center).Length(); lng > radius {
		radius = lng
	}
	return NewSphere(center, radius)
}
