package solver

import "github.com/mokiat/gomath/dprec"

const epsilon = float64(0.00001)

// SafeNormal tries to return the specified vector in unit length. If the
// vector is too small or zero then the fallback vector is returned as is.
func SafeNormal(vector, fallback dprec.Vec3) dprec.Vec3 {
	if vector.Length() < epsilon {
		return fallback
	}
	return dprec.UnitVec3(vector)
}
