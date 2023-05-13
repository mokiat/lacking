package solver

import "github.com/mokiat/gomath/dprec"

type Impulse struct {
	Linear  dprec.Vec3
	Angular dprec.Vec3
}

type PairImpulse struct {
	Target Impulse
	Source Impulse
}

type Nudge struct {
	Linear  dprec.Vec3
	Angular dprec.Vec3
}

type PairNudge struct {
	Target Nudge
	Source Nudge
}
