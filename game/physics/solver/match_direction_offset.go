package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

// NewMatchDirectionOffset creates a new MatchDirectionOffset constraint solver.
func NewMatchDirectionOffset() *MatchDirectionOffset {
	return &MatchDirectionOffset{
		primaryAnchor:   dprec.ZeroVec3(),
		secondaryAnchor: dprec.ZeroVec3(),
		direction:       dprec.BasisYVec3(),
		offset:          0.0,
	}
}

var _ physics.DBConstraintSolver = (*MatchDirectionOffset)(nil)

// MatchDirectionOffset represents the solution for a constraint which ensures that
// the second body is at an exact distance away from the first body along
// some direction of the first body.
type MatchDirectionOffset struct {
	primaryAnchor   dprec.Vec3
	secondaryAnchor dprec.Vec3
	direction       dprec.Vec3
	offset          float64

	jacobian physics.PairJacobian
	drift    float64
}

// PrimaryAnchor returns the attachment point on the primary body to which the
// secondary will match.
func (s *MatchDirectionOffset) PrimaryAnchor() dprec.Vec3 {
	return s.primaryAnchor
}

// SetPrimaryAnchor changes the attachment point on the primary body.
func (s *MatchDirectionOffset) SetPrimaryAnchor(anchor dprec.Vec3) *MatchDirectionOffset {
	s.primaryAnchor = anchor
	return s
}

// SecondaryAnchor returns the attachment point on the primary body to which the
// primary will match.
func (s *MatchDirectionOffset) SecondaryAnchor() dprec.Vec3 {
	return s.secondaryAnchor
}

// SetSecondaryAnchor changes the attachment point on the secondary body.
func (s *MatchDirectionOffset) SetSecondaryAnchor(anchor dprec.Vec3) *MatchDirectionOffset {
	s.secondaryAnchor = anchor
	return s
}

// Direction returns the constraint direction, which is in local space of
// the first body.
func (s *MatchDirectionOffset) Direction() dprec.Vec3 {
	return s.direction
}

// SetDirection changes the constraint direction, which must be in local space
// of the first body.
func (s *MatchDirectionOffset) SetDirection(direction dprec.Vec3) *MatchDirectionOffset {
	s.direction = dprec.UnitVec3(direction)
	return s
}

// Offset returns the directional offset.
func (s *MatchDirectionOffset) Offset() float64 {
	return s.offset
}

// SetOffset changes the directional offset.
func (s *MatchDirectionOffset) SetOffset(offset float64) *MatchDirectionOffset {
	s.offset = offset
	return s
}

func (s *MatchDirectionOffset) Reset(ctx physics.DBSolverContext) {
	s.updateJacobian(ctx)
}

func (s *MatchDirectionOffset) ApplyImpulses(ctx physics.DBSolverContext) {
	ctx.ApplyImpulseSolution(ctx.JacobianImpulseSolution(s.jacobian, s.drift, 0.0))
}

func (s *MatchDirectionOffset) ApplyNudges(ctx physics.DBSolverContext) {}

func (s *MatchDirectionOffset) updateJacobian(ctx physics.DBSolverContext) {
	dirWS := dprec.QuatVec3Rotation(ctx.Primary.Orientation(), s.direction)
	primaryRadiusWS := dprec.QuatVec3Rotation(ctx.Primary.Orientation(), s.primaryAnchor)
	secondaryRadiusWS := dprec.QuatVec3Rotation(ctx.Secondary.Orientation(), s.secondaryAnchor)
	s.jacobian = physics.PairJacobian{
		Primary: physics.Jacobian{
			SlopeVelocity:        dprec.InverseVec3(dirWS),
			SlopeAngularVelocity: dprec.Vec3Cross(dirWS, primaryRadiusWS),
		},
		Secondary: physics.Jacobian{
			SlopeVelocity:        dirWS,
			SlopeAngularVelocity: dprec.Vec3Cross(secondaryRadiusWS, dirWS),
		},
	}
	deltaPosition := dprec.Vec3Diff(
		dprec.Vec3Sum(ctx.Secondary.Position(), secondaryRadiusWS),
		dprec.Vec3Sum(ctx.Primary.Position(), primaryRadiusWS),
	)
	s.drift = dprec.Vec3Dot(dirWS, deltaPosition)
}
