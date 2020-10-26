package physics

import "github.com/mokiat/gomath/sprec"

type MatchRotationConstraint struct {
	NilConstraint
	FirstBody  *Body
	SecondBody *Body
}

func (c MatchRotationConstraint) ApplyImpulse(ctx Context) {
	c.yConstraint().ApplyImpulse(ctx)
	c.zConstraint().ApplyImpulse(ctx)
}

func (c MatchRotationConstraint) ApplyNudge(ctx Context) {
	c.yConstraint().ApplyNudge(ctx)
	c.zConstraint().ApplyNudge(ctx)
}

func (c MatchRotationConstraint) yConstraint() MatchAxisConstraint {
	return MatchAxisConstraint{
		FirstBody:      c.FirstBody,
		FirstBodyAxis:  sprec.BasisYVec3(),
		SecondBody:     c.SecondBody,
		SecondBodyAxis: sprec.BasisYVec3(),
	}
}

func (c MatchRotationConstraint) zConstraint() MatchAxisConstraint {
	return MatchAxisConstraint{
		FirstBody:      c.FirstBody,
		FirstBodyAxis:  sprec.BasisZVec3(),
		SecondBody:     c.SecondBody,
		SecondBodyAxis: sprec.BasisZVec3(),
	}
}
