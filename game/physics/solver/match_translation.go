package solver

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.DBConstraintSolver = (*MatchTranslation)(nil)

// NewMatchTranslation creates a new MatchTranslation constraint solver.
func NewMatchTranslation() *MatchTranslation {
	result := &MatchTranslation{}
	result.DBJacobianConstraintSolver = physics.NewDBJacobianConstraintSolver(result.calculate)
	return result
}

// MatchTranslation represents the solution for a constraint
// that keeps a secondary body attached to an offset of the primary body.
// It is possible to disable the constraint for certain axis of the
// primary body.
type MatchTranslation struct {
	*physics.DBJacobianConstraintSolver

	primaryAnchor sprec.Vec3
	ignoreX       bool
	ignoreY       bool
	ignoreZ       bool
}

// PrimaryAnchor returns the attachment point on the primary
// body to which the secondary will match.
func (t *MatchTranslation) PrimaryAnchor() sprec.Vec3 {
	return t.primaryAnchor
}

// SetPrimaryAnchor changes the attachment point on the primary
// body.
func (t *MatchTranslation) SetPrimaryAnchor(anchor sprec.Vec3) *MatchTranslation {
	t.primaryAnchor = anchor
	return t
}

// IgnoreX returns whether the X dimension, relative to the
// primary body, will be matched.
func (t *MatchTranslation) IgnoreX() bool {
	return t.ignoreX
}

// SetIgnoreX changes whether the X dimension, relative to the
// primary body, will be considered.
func (t *MatchTranslation) SetIgnoreX(ignore bool) *MatchTranslation {
	t.ignoreX = ignore
	return t
}

// IgnoreY returns whether the Y dimension, relative to the
// primary body, will be matched.
func (t *MatchTranslation) IgnoreY() bool {
	return t.ignoreY
}

// SetIgnoreY changes whether the Y dimension, relative to the
// primary body, will be considered.
func (t *MatchTranslation) SetIgnoreY(ignore bool) *MatchTranslation {
	t.ignoreY = ignore
	return t
}

// IgnoreZ returns whether the Z dimension, relative to the
// primary body, will be matched.
func (t *MatchTranslation) IgnoreZ() bool {
	return t.ignoreZ
}

// SetIgnoreZ changes whether the Z dimension, relative to the
// primary body, will be considered.
func (t *MatchTranslation) SetIgnoreZ(ignore bool) *MatchTranslation {
	t.ignoreZ = ignore
	return t
}

func (t *MatchTranslation) calculate(ctx physics.DBSolverContext) (physics.PairJacobian, float32) {
	firstRadiusWS := sprec.QuatVec3Rotation(ctx.Primary.Orientation(), t.primaryAnchor)
	firstAnchorWS := sprec.Vec3Sum(ctx.Primary.Position(), firstRadiusWS)
	deltaPosition := sprec.Vec3Diff(ctx.Secondary.Position(), firstAnchorWS)
	if t.ignoreX {
		deltaPosition = sprec.Vec3Diff(deltaPosition, sprec.Vec3Prod(ctx.Primary.Orientation().OrientationX(), sprec.Vec3Dot(deltaPosition, ctx.Primary.Orientation().OrientationX())))
	}
	if t.ignoreY {
		deltaPosition = sprec.Vec3Diff(deltaPosition, sprec.Vec3Prod(ctx.Primary.Orientation().OrientationY(), sprec.Vec3Dot(deltaPosition, ctx.Primary.Orientation().OrientationY())))
	}
	if t.ignoreZ {
		deltaPosition = sprec.Vec3Diff(deltaPosition, sprec.Vec3Prod(ctx.Primary.Orientation().OrientationZ(), sprec.Vec3Dot(deltaPosition, ctx.Primary.Orientation().OrientationZ())))
	}
	normal := sprec.BasisXVec3()
	if deltaPosition.SqrLength() > sqrEpsilon {
		normal = sprec.UnitVec3(deltaPosition)
	}
	return physics.PairJacobian{
			Primary: physics.Jacobian{
				SlopeVelocity: sprec.NewVec3(
					-normal.X,
					-normal.Y,
					-normal.Z,
				),
				SlopeAngularVelocity: sprec.NewVec3(
					-(normal.Z*firstRadiusWS.Y - normal.Y*firstRadiusWS.Z),
					-(normal.X*firstRadiusWS.Z - normal.Z*firstRadiusWS.X),
					-(normal.Y*firstRadiusWS.X - normal.X*firstRadiusWS.Y),
				),
			},
			Secondary: physics.Jacobian{
				SlopeVelocity: sprec.NewVec3(
					normal.X,
					normal.Y,
					normal.Z,
				),
				SlopeAngularVelocity: sprec.ZeroVec3(),
			},
		},
		deltaPosition.Length()
}
