package asset

import "github.com/mokiat/gomath/dprec"

const UnspecifiedNodeIndex = int32(-1)

type Node struct {
	Name        string
	ParentIndex int32
	Translation dprec.Vec3
	Rotation    dprec.Quat
	Scale       dprec.Vec3
}
