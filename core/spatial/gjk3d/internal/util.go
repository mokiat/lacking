package internal

import (
	"math"

	"github.com/mokiat/gomath/dprec"
)

// perpendicularVec3 returns an arbitrary vector that is perpendicular to v.
// The basis axis least aligned with v is used for the cross product, which
// guarantees a non-degenerate result for any non-zero v.
func perpendicularVec3(v dprec.Vec3) dprec.Vec3 {
	absX := math.Abs(v.X)
	absY := math.Abs(v.Y)
	absZ := math.Abs(v.Z)
	switch {
	case absX <= absY && absX <= absZ:
		return dprec.Vec3Cross(v, dprec.BasisXVec3())
	case absY <= absZ:
		return dprec.Vec3Cross(v, dprec.BasisYVec3())
	default:
		return dprec.Vec3Cross(v, dprec.BasisZVec3())
	}
}
