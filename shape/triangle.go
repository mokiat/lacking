package shape

import "github.com/mokiat/gomath/sprec"

func NewStaticTriangle(a, b, c sprec.Vec3) StaticTriangle {
	return StaticTriangle{
		a: a,
		b: b,
		c: c,
	}
}

type StaticTriangle struct {
	a sprec.Vec3
	b sprec.Vec3
	c sprec.Vec3
}

func (t StaticTriangle) Transformed(translation sprec.Vec3, rotation sprec.Quat) StaticTriangle {
	return StaticTriangle{
		a: sprec.Vec3Sum(translation, sprec.QuatVec3Rotation(rotation, t.a)),
		b: sprec.Vec3Sum(translation, sprec.QuatVec3Rotation(rotation, t.b)),
		c: sprec.Vec3Sum(translation, sprec.QuatVec3Rotation(rotation, t.c)),
	}
}

func (t StaticTriangle) A() sprec.Vec3 {
	return t.a
}

func (t StaticTriangle) B() sprec.Vec3 {
	return t.b
}

func (t StaticTriangle) C() sprec.Vec3 {
	return t.c
}

func (t StaticTriangle) Normal() sprec.Vec3 {
	vecAB := sprec.Vec3Diff(t.b, t.a)
	vecAC := sprec.Vec3Diff(t.c, t.a)
	return sprec.UnitVec3(sprec.Vec3Cross(vecAB, vecAC))
}

func (t StaticTriangle) Area() float32 {
	vecAB := sprec.Vec3Diff(t.b, t.a)
	vecAC := sprec.Vec3Diff(t.c, t.a)
	return sprec.Vec3Cross(vecAB, vecAC).Length() / 2.0
}

func (t StaticTriangle) IsLookingTowards(direction sprec.Vec3) bool {
	return sprec.Vec3Dot(t.Normal(), direction) > 0.0
}

func (t StaticTriangle) ContainsPoint(point sprec.Vec3) bool {
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
