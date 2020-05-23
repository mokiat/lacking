package shape

import "github.com/mokiat/gomath/sprec"

type Placement struct {
	Position    sprec.Vec3
	Orientation sprec.Quat
	Shape       interface{}
}

func (p Placement) Transformed(translation sprec.Vec3, rotation sprec.Quat) Placement {
	p.Position = sprec.Vec3Sum(translation, sprec.QuatVec3Rotation(rotation, p.Position))
	p.Orientation = sprec.QuatProd(rotation, p.Orientation)
	return p
}
