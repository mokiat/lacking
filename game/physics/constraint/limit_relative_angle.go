package constraint

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

func NewLimitRelativeAngle() *LimitRelativeAngle {
	return &LimitRelativeAngle{}
}

var _ solver.PairConstraint = (*LimitRelativeAngle)(nil)

type LimitRelativeAngle struct {
	primaryDirection   dprec.Vec3
	secondaryDirection dprec.Vec3
	axis               dprec.Vec3
	minAngle           dprec.Angle
	maxAngle           dprec.Angle

	jacobian solver.PairJacobian
	drift    float64
}

func (c *LimitRelativeAngle) PrimaryDirection() dprec.Vec3 {
	return c.primaryDirection
}

func (c *LimitRelativeAngle) SetPrimaryDirection(direction dprec.Vec3) *LimitRelativeAngle {
	c.primaryDirection = dprec.UnitVec3(direction)
	return c
}

func (c *LimitRelativeAngle) SecondaryDirection() dprec.Vec3 {
	return c.secondaryDirection
}

func (c *LimitRelativeAngle) SetSecondaryDirection(direction dprec.Vec3) *LimitRelativeAngle {
	c.secondaryDirection = dprec.UnitVec3(direction)
	return c
}

func (c *LimitRelativeAngle) Axis() dprec.Vec3 {
	return c.axis
}

func (c *LimitRelativeAngle) SetAxis(axis dprec.Vec3) *LimitRelativeAngle {
	c.axis = dprec.UnitVec3(axis)
	return c
}

func (c *LimitRelativeAngle) MinAngle() dprec.Angle {
	return c.minAngle
}

func (c *LimitRelativeAngle) SetMinAngle(angle dprec.Angle) *LimitRelativeAngle {
	c.minAngle = angle
	return c
}

func (c *LimitRelativeAngle) MaxAngle() dprec.Angle {
	return c.maxAngle
}

func (c *LimitRelativeAngle) SetMaxAngle(angle dprec.Angle) *LimitRelativeAngle {
	c.maxAngle = angle
	return c
}

func (c *LimitRelativeAngle) Reset(ctx solver.PairContext) {
	axisWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), c.axis)
	primaryDirectionWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), c.primaryDirection)
	secondaryDirectionWS := dprec.QuatVec3Rotation(ctx.Source.Rotation(), c.secondaryDirection)

	if dprec.Abs(dprec.Vec3Dot(axisWS, secondaryDirectionWS)) > 0.99 {
		return // secondary direction is parallel to axis
	}
	angle := dprec.Vec3ProjectionAngle(primaryDirectionWS, secondaryDirectionWS, axisWS)

	switch {
	case angle > c.maxAngle:
		c.drift = (angle - c.maxAngle).Radians()
		c.jacobian = solver.PairJacobian{
			Target: solver.Jacobian{
				AngularSlope: dprec.InverseVec3(axisWS),
			},
			Source: solver.Jacobian{
				AngularSlope: axisWS,
			},
		}
	case angle < c.minAngle:
		c.drift = (c.minAngle - angle).Radians()
		c.jacobian = solver.PairJacobian{
			Target: solver.Jacobian{
				AngularSlope: axisWS,
			},
			Source: solver.Jacobian{
				AngularSlope: dprec.InverseVec3(axisWS),
			},
		}
	default:
		c.drift = 0.0
		c.jacobian = solver.PairJacobian{}
	}
}

func (c *LimitRelativeAngle) ApplyImpulses(ctx solver.PairContext) {
	if lambda := ctx.JacobianImpulseLambda(c.jacobian, 0.0, 0.0); lambda >= 0.0 {
		return // moving away
	}
	solution := ctx.JacobianImpulseSolution(c.jacobian, c.drift, 0.0)
	ctx.Target.ApplyImpulse(solution.Target)
	ctx.Source.ApplyImpulse(solution.Source)
}

func (c *LimitRelativeAngle) ApplyNudges(ctx solver.PairContext) {
}
