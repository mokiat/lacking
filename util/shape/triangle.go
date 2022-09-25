package shape

import "github.com/mokiat/gomath/dprec"

// NewStaticTriangle creates a new StaticTriangle.
func NewStaticTriangle(a, b, c dprec.Vec3) StaticTriangle {
	return StaticTriangle{
		a: a,
		b: b,
		c: c,
	}
}

// StaticTriangle represents a tringle in 3D space.
type StaticTriangle struct {
	a dprec.Vec3
	b dprec.Vec3
	c dprec.Vec3
}

// Transformed returns a new StaticTriangle that is the result of applying
// the specified rotation and translation to the current triangle.
func (t StaticTriangle) Transformed(translation dprec.Vec3, rotation dprec.Quat) StaticTriangle {
	return StaticTriangle{
		a: dprec.Vec3Sum(translation, dprec.QuatVec3Rotation(rotation, t.a)),
		b: dprec.Vec3Sum(translation, dprec.QuatVec3Rotation(rotation, t.b)),
		c: dprec.Vec3Sum(translation, dprec.QuatVec3Rotation(rotation, t.c)),
	}
}

// A returns the first corner of the triangle.
func (t StaticTriangle) A() dprec.Vec3 {
	return t.a
}

// B returns the second corner of the triangle.
func (t StaticTriangle) B() dprec.Vec3 {
	return t.b
}

// C returns the third corner of the triangle.
func (t StaticTriangle) C() dprec.Vec3 {
	return t.c
}

// Normal returns the orientation of the triangle's surface.
func (t StaticTriangle) Normal() dprec.Vec3 {
	vecAB := dprec.Vec3Diff(t.b, t.a)
	vecAC := dprec.Vec3Diff(t.c, t.a)
	return dprec.UnitVec3(dprec.Vec3Cross(vecAB, vecAC))
}

// Area returns the triangle's surface area.
func (t StaticTriangle) Area() float64 {
	vecAB := dprec.Vec3Diff(t.b, t.a)
	vecAC := dprec.Vec3Diff(t.c, t.a)
	return dprec.Vec3Cross(vecAB, vecAC).Length() / 2.0
}

// IsLookingTowards checks whether the orientation of the triangle looks towards
// the same hemisphere as the provided direction.
func (t StaticTriangle) IsLookingTowards(direction dprec.Vec3) bool {
	return dprec.Vec3Dot(t.Normal(), direction) > 0.0
}

// ContainsPoint checks whether the specified point is is inside the triangle.
//
// Beware, currently this method assumes that the point lies somewhere on the
// surface plane of the triangle.
func (t StaticTriangle) ContainsPoint(point dprec.Vec3) bool {
	normal := t.Normal()
	if triangleABP := NewStaticTriangle(t.a, t.b, point); !triangleABP.IsLookingTowards(normal) {
		return false
	}
	if triangleBCP := NewStaticTriangle(t.b, t.c, point); !triangleBCP.IsLookingTowards(normal) {
		return false
	}
	if triangleCAP := NewStaticTriangle(t.c, t.a, point); !triangleCAP.IsLookingTowards(normal) {
		return false
	}
	return true
}
