package shape3d

import "github.com/mokiat/gomath/dprec"

type Box struct {
	Position   dprec.Vec3
	Rotation   dprec.Quat
	HalfWidth  float64
	HalfHeight float64
	HalfLength float64
}
