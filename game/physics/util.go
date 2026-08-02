package physics

import "github.com/mokiat/gomath/dprec"

func QuatFromVector(vector dprec.Vec3) dprec.Quat {
	radians := vector.Length()

	const angularEpsilon = float64(0.00001)
	if dprec.Abs(radians) < angularEpsilon {
		return dprec.IdentityQuat()
	}

	return dprec.RotationQuat(dprec.Radians(radians), vector)
}
