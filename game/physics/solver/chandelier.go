package solver

import (
	"github.com/mokiat/gomath/sprec"
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

	fixture    sprec.Vec3
	bodyAnchor sprec.Vec3
	length     float32
}

// Fixture returns the fixture location for the chandelier hook.
func (c *Chandelier) Fixture() sprec.Vec3 {
	return c.fixture
}

// SetFixture changes the fixture location for the chandelier hook.
func (c *Chandelier) SetFixture(fixture sprec.Vec3) *Chandelier {
	c.fixture = fixture
	return c
}

// BodyAnchor returns the offset from the center of mass of the
// body that it is wired to the chandelier.
func (c *Chandelier) BodyAnchor() sprec.Vec3 {
	return c.bodyAnchor
}

// SetBodyAnchor changes the offset at which the body is attached
// to the chandelier wiring.
func (c *Chandelier) SetBodyAnchor(anchor sprec.Vec3) *Chandelier {
	c.bodyAnchor = anchor
	return c
}

// Length returns the chandelier length.
func (c *Chandelier) Length() float32 {
	return c.length
}

// SetLength changes the chandelier length.
func (c *Chandelier) SetLength(length float32) *Chandelier {
	c.length = length
	return c
}

func (c *Chandelier) calculate(ctx physics.SBSolverContext) (physics.Jacobian, float32) {
	body := ctx.Body
	anchorWS := sprec.Vec3Sum(body.Position(), sprec.QuatVec3Rotation(body.Orientation(), c.bodyAnchor))
	radiusWS := sprec.Vec3Diff(anchorWS, body.Position())
	deltaPosition := sprec.Vec3Diff(anchorWS, c.fixture)
	normal := sprec.BasisXVec3()
	if deltaPosition.SqrLength() > sqrEpsilon {
		normal = sprec.UnitVec3(deltaPosition)
	}
	return physics.Jacobian{
			SlopeVelocity: sprec.NewVec3(
				normal.X,
				normal.Y,
				normal.Z,
			),
			SlopeAngularVelocity: sprec.NewVec3(
				normal.Z*radiusWS.Y-normal.Y*radiusWS.Z,
				normal.X*radiusWS.Z-normal.Z*radiusWS.X,
				normal.Y*radiusWS.X-normal.X*radiusWS.Y,
			),
		},
		deltaPosition.Length() - c.length
}
