package shape3d

import "github.com/mokiat/gomath/dprec"

func NewAABB(lower, higher dprec.Vec3) AABB {
	return AABB{
		MinX: lower.X,
		MinY: lower.Y,
		MinZ: lower.Z,
		MaxX: higher.X,
		MaxY: higher.Y,
		MaxZ: higher.Z,
	}
}

type AABB struct {
	MinX float64
	MinY float64
	MinZ float64
	MaxX float64
	MaxY float64
	MaxZ float64
}
