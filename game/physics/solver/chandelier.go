package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

// NewChandelier creates a new Chandelier constraint solver.
func NewChandelier() *Chandelier {
	return &Chandelier{
		fixture:    dprec.ZeroVec3(),
		bodyAnchor: dprec.ZeroVec3(),
		length:     1.0,
	}
}

var _ physics.ExplicitSBConstraintSolver = (*Chandelier)(nil)

// Chandelier represents the solution for a constraint
// that keeps a body hanging off of a fixture location similar
// to a chandelier.
type Chandelier struct {
	physics.NilSBConstraintSolver // TODO: Remove

	fixture    dprec.Vec3
	bodyAnchor dprec.Vec3
	length     float64

	jacobian physics.Jacobian
	drift    float64
}

// Fixture returns the fixture location for the chandelier hook.
func (c *Chandelier) Fixture() dprec.Vec3 {
	return c.fixture
}

// SetFixture changes the fixture location for the chandelier hook.
func (c *Chandelier) SetFixture(fixture dprec.Vec3) *Chandelier {
	c.fixture = fixture
	return c
}

// BodyAnchor returns the offset from the center of mass of the
// body that it is wired to the chandelier.
func (c *Chandelier) BodyAnchor() dprec.Vec3 {
	return c.bodyAnchor
}

// SetBodyAnchor changes the offset at which the body is attached
// to the chandelier wiring.
func (c *Chandelier) SetBodyAnchor(anchor dprec.Vec3) *Chandelier {
	c.bodyAnchor = anchor
	return c
}

// Length returns the chandelier length.
func (c *Chandelier) Length() float64 {
	return c.length
}

// SetLength changes the chandelier length.
func (c *Chandelier) SetLength(length float64) *Chandelier {
	c.length = length
	return c
}

func (c *Chandelier) Reset(ctx physics.SBSolverContext) {
	c.updateJacobian(ctx)
}

func (c *Chandelier) ApplyImpulses(ctx physics.SBSolverContext) {
	ctx.ApplyImpulse(c.jacobian)
}

func (c *Chandelier) ApplyNudges(ctx physics.SBSolverContext) {
	c.updateJacobian(ctx)
	ctx.ApplyNudge(c.jacobian, c.drift)
}

func (c *Chandelier) updateJacobian(ctx physics.SBSolverContext) {
	radiusWS := dprec.QuatVec3Rotation(ctx.Body.Orientation(), c.bodyAnchor)
	anchorWS := dprec.Vec3Sum(ctx.Body.Position(), radiusWS)
	deltaPositionWS := dprec.Vec3Diff(anchorWS, c.fixture)
	normalWS := SafeNormal(deltaPositionWS, dprec.BasisXVec3())
	c.jacobian = physics.Jacobian{
		SlopeVelocity:        normalWS,
		SlopeAngularVelocity: dprec.Vec3Cross(radiusWS, normalWS),
	}
	c.drift = deltaPositionWS.Length() - c.length
}
