package hierarchy

import "github.com/mokiat/gomath/dprec"

type nodeState struct {
	translation dprec.Vec3
	rotation    dprec.Quat
	scale       dprec.Vec3
	absMatrix   dprec.Mat4
}

func (s *nodeState) initialize() {
	s.translation = dprec.ZeroVec3()
	s.rotation = dprec.IdentityQuat()
	s.scale = dprec.NewVec3(1.0, 1.0, 1.0)
	s.absMatrix = dprec.IdentityMat4()
}
