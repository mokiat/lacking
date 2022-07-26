package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.SBConstraintSolver = (*Chandelier)(nil)

// NewChandelier creates a new Chandelier constraint solver.
func NewChandelier() *Chandelier {
	result := &Chandelier{
		length: 1.0,
	}
	result.SBJacobianConstraintSolver = physics.NewSBJacobianConstraintSolver(result.calculate)
	return result
}

// Chandelier represents the solution for a constraint
// that keeps a body hanging off of a fixture location similar
// to a chandelier.
type Chandelier struct {
	*physics.SBJacobianConstraintSolver

	fixture    dprec.Vec3
	bodyAnchor dprec.Vec3
	length     float64
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

func (c *Chandelier) calculate(ctx physics.SBSolverContext) (physics.Jacobian, float64) {
	anchorWS := dprec.Vec3Sum(ctx.Body.Position(), dprec.QuatVec3Rotation(ctx.Body.Orientation(), c.bodyAnchor))
	radiusWS := dprec.Vec3Diff(anchorWS, ctx.Body.Position())
	deltaPosition := dprec.Vec3Diff(anchorWS, c.fixture)
	normal := dprec.BasisXVec3()
	if deltaPosition.SqrLength() > sqrEpsilon {
		normal = dprec.UnitVec3(deltaPosition)
	}
	return physics.Jacobian{
			SlopeVelocity: dprec.NewVec3(
				normal.X,
				normal.Y,
				normal.Z,
			),
			SlopeAngularVelocity: dprec.NewVec3(
				normal.Z*radiusWS.Y-normal.Y*radiusWS.Z,
				normal.X*radiusWS.Z-normal.Z*radiusWS.X,
				normal.Y*radiusWS.X-normal.X*radiusWS.Y,
			),
		},
		deltaPosition.Length() - c.length
}
