package shape

import "github.com/mokiat/gomath/sprec"

func NewStaticLine(a, b sprec.Vec3) StaticLine {
	return StaticLine{
		a: a,
		b: b,
	}
}

type StaticLine struct {
	a sprec.Vec3
	b sprec.Vec3
}

func (l StaticLine) A() sprec.Vec3 {
	return l.a
}

func (l StaticLine) B() sprec.Vec3 {
	return l.b
}

func (l StaticLine) SqrLength() float32 {
	return sprec.Vec3Diff(l.b, l.a).SqrLength()
}

func (l StaticLine) Length() float32 {
	return sprec.Vec3Diff(l.b, l.a).Length()
}
