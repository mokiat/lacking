package solver

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.SBConstraintSolver = (*FixedTranslation)(nil)

// NewFixedTranslation creates a new FixedTranslation
// constraint solver.
func NewFixedTranslation() *FixedTranslation {
	result := &FixedTranslation{}
	result.SBJacobianConstraintSolver = physics.NewSBJacobianConstraintSolver(result.calculate)
	return result
}

// FixedTranslation represents the solution for a constraint
// that keeps a body positioned at the specified fixture location.
type FixedTranslation struct {
	*physics.SBJacobianConstraintSolver

	fixture sprec.Vec3
}

// Fixture returns the location that the body will be
// tied to.
func (t *FixedTranslation) Fixture() sprec.Vec3 {
	return t.fixture
}

// SetFixture changes the location to which the body
// will be constrained.
func (t *FixedTranslation) SetFixture(fixture sprec.Vec3) *FixedTranslation {
	t.fixture = fixture
	return t
}

func (t *FixedTranslation) calculate(ctx physics.SBSolverContext) (physics.Jacobian, float32) {
	body := ctx.Body
	deltaPosition := sprec.Vec3Diff(body.Position(), t.fixture)
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
			SlopeAngularVelocity: sprec.ZeroVec3(),
		},
		deltaPosition.Length()
}
