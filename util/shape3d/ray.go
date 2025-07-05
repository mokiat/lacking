package shape3d

import "github.com/mokiat/gomath/dprec"

func NewRay(origin, direction dprec.Vec3) Ray {
	return Ray{
		Origin:    origin,
		Direction: direction,
	}
}

type Ray struct {
	Origin    dprec.Vec3
	Direction dprec.Vec3
}
