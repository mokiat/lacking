package solver

import (
	"github.com/mokiat/gomath/dprec"
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

	fixture dprec.Vec3
}

// Fixture returns the location that the body will be
// tied to.
func (t *FixedTranslation) Fixture() dprec.Vec3 {
	return t.fixture
}

// SetFixture changes the location to which the body
// will be constrained.
func (t *FixedTranslation) SetFixture(fixture dprec.Vec3) *FixedTranslation {
	t.fixture = fixture
	return t
}

func (t *FixedTranslation) calculate(ctx physics.SBSolverContext) (physics.Jacobian, float64) {
	deltaPosition := dprec.Vec3Diff(ctx.Body.Position(), t.fixture)
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
			SlopeAngularVelocity: dprec.ZeroVec3(),
		},
		deltaPosition.Length()
}
