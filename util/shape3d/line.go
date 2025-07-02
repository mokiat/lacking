package shape3d

import "github.com/mokiat/gomath/dprec"

func L(a, b dprec.Vec3) Line {
	return Line{
		A: a,
		B: b,
	}
}

type Line struct {
	A dprec.Vec3
	B dprec.Vec3
}
