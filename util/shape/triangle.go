package shape

import "github.com/mokiat/gomath/dprec"

// NewStaticTriangle creates a new StaticTriangle shape.
func NewStaticTriangle(a, b, c Point) StaticTriangle {
	result := StaticTriangle{
		a: a,
		b: b,
		c: c,
	}
	center := result.Center()
	result.bsRadius = dprec.Max(
		dprec.Max(
			dprec.Vec3Diff(dprec.Vec3(a), dprec.Vec3(center)).Length(),
			dprec.Vec3Diff(dprec.Vec3(b), dprec.Vec3(center)).Length(),
		),
		dprec.Vec3Diff(dprec.Vec3(c), dprec.Vec3(center)).Length(),
	)
	return result
}

// StaticTriangle represents a tringle in 3D space.
type StaticTriangle struct {
	a        Point
	b        Point
	c        Point
	bsRadius float64
}

func (t StaticTriangle) Center() Point {
	return Point(dprec.Vec3Quot(dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3(t.a), dprec.Vec3(t.b)), dprec.Vec3(t.c)), 3.0))
}

func (t StaticTriangle) BoundingSphereRadius() float64 {
	return t.bsRadius
}

// A returns the first corner of the triangle.
func (t StaticTriangle) A() Point {
	return t.a
}

// WithA returns a new StaticTriangle based on this one but has the specified
// A point.
func (t StaticTriangle) WithA(a Point) StaticTriangle {
	t.a = a
	return t
}

// B returns the second corner of the triangle.
func (t StaticTriangle) B() Point {
	return t.b
}

// WithB returns a new StaticTriangle based on this one but has the specified
// B point.
func (t StaticTriangle) WithB(b Point) StaticTriangle {
	t.b = b
	return t
}

// C returns the third corner of the triangle.
func (t StaticTriangle) C() Point {
	return t.c
}

// WithC returns a new StaticTriangle based on this one but has the specified
// C point.
func (t StaticTriangle) WithC(c Point) StaticTriangle {
	t.c = c
	return t
}

// Area returns the triangle's surface area.
func (t StaticTriangle) Area() float64 {
	vecAB := dprec.Vec3Diff(dprec.Vec3(t.b), dprec.Vec3(t.a))
	vecAC := dprec.Vec3Diff(dprec.Vec3(t.c), dprec.Vec3(t.a))
	return dprec.Vec3Cross(vecAB, vecAC).Length() / 2.0
}

// Normal returns the orientation of the triangle's surface.
func (t StaticTriangle) Normal() dprec.Vec3 {
	vecAB := dprec.Vec3Diff(dprec.Vec3(t.b), dprec.Vec3(t.a))
	vecAC := dprec.Vec3Diff(dprec.Vec3(t.c), dprec.Vec3(t.a))
	return dprec.UnitVec3(dprec.Vec3Cross(vecAB, vecAC))
}

// IsLookingTowards checks whether the orientation of the triangle looks towards
// the same hemisphere as the provided direction.
func (t StaticTriangle) IsLookingTowards(direction dprec.Vec3) bool {
	return dprec.Vec3Dot(t.Normal(), direction) > 0.0
}

// ContainsPoint checks whether the specified Point is inside the triangle.
//
// Beware, currently this method assumes that the point lies somewhere on the
// surface plane of the StaticTriangle.
func (t StaticTriangle) ContainsPoint(point Point) bool {
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

// Transformed returns a new StaticTriangle that is the result of applying
// the specified rotation and translation to the current triangle.
func (t StaticTriangle) Transformed(parent Transform) StaticTriangle {
	return StaticTriangle{
		a:        t.a.Transformed(parent),
		b:        t.b.Transformed(parent),
		c:        t.c.Transformed(parent),
		bsRadius: t.bsRadius,
	}
}
