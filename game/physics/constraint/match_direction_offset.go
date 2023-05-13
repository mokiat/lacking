package constraint

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

// NewMatchDirectionOffset creates a new MatchDirectionOffset constraint solver.
func NewMatchDirectionOffset() *MatchDirectionOffset {
	return &MatchDirectionOffset{
		primaryRadius:   dprec.ZeroVec3(),
		secondaryRadius: dprec.ZeroVec3(),
		direction:       dprec.BasisYVec3(),
		offset:          0.0,
	}
}

var _ solver.PairConstraint = (*MatchDirectionOffset)(nil)

// MatchDirectionOffset represents the solution for a constraint which ensures that
// the second body is at an exact distance away from the first body along
// some direction of the first body.
type MatchDirectionOffset struct {
	primaryRadius   dprec.Vec3
	secondaryRadius dprec.Vec3
	direction       dprec.Vec3
	offset          float64

	jacobian solver.PairJacobian
	drift    float64
}

// PrimaryRadius returns the radius vector of the contact point
// on the primary object.
//
// The vector is in the object's local space.
func (s *MatchDirectionOffset) PrimaryRadius() dprec.Vec3 {
	return s.primaryRadius
}

// SetPrimaryRadius changes the attachment point of the link
// on the primary body.
func (s *MatchDirectionOffset) SetPrimaryRadius(radius dprec.Vec3) *MatchDirectionOffset {
	s.primaryRadius = radius
	return s
}

// SecondaryRadius returns the radius vector of the contact point
// on the secondary object.
//
// The vector is in the object's local space.
func (s *MatchDirectionOffset) SecondaryRadius() dprec.Vec3 {
	return s.secondaryRadius
}

// SetSecondaryRadius changes the radius vector of the contact point
// on the secondary object.
//
// The vector is in the object's local space.
func (s *MatchDirectionOffset) SetSecondaryRadius(radius dprec.Vec3) *MatchDirectionOffset {
	s.secondaryRadius = radius
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

func (s *MatchDirectionOffset) Reset(ctx solver.PairContext) {
	dirWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.direction)
	primaryRadiusWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.primaryRadius)
	secondaryRadiusWS := dprec.QuatVec3Rotation(ctx.Source.Rotation(), s.secondaryRadius)
	s.jacobian = solver.PairJacobian{
		Target: solver.Jacobian{
			LinearSlope:  dprec.InverseVec3(dirWS),
			AngularSlope: dprec.Vec3Cross(dirWS, primaryRadiusWS),
		},
		Source: solver.Jacobian{
			LinearSlope:  dirWS,
			AngularSlope: dprec.Vec3Cross(secondaryRadiusWS, dirWS),
		},
	}
	deltaPosition := dprec.Vec3Diff(
		dprec.Vec3Sum(ctx.Source.Position(), secondaryRadiusWS),
		dprec.Vec3Sum(ctx.Target.Position(), primaryRadiusWS),
	)
	s.drift = dprec.Vec3Dot(dirWS, deltaPosition)
}

func (s *MatchDirectionOffset) ApplyImpulses(ctx solver.PairContext) {
	solution := ctx.JacobianImpulseSolution(s.jacobian, s.drift, 0.0)
	ctx.Target.ApplyImpulse(solution.Target)
	ctx.Source.ApplyImpulse(solution.Source)
}

func (s *MatchDirectionOffset) ApplyNudges(ctx solver.PairContext) {
	solution := ctx.JacobianNudgeSolution(s.jacobian, s.drift)
	ctx.Target.ApplyNudge(solution.Target)
	ctx.Source.ApplyNudge(solution.Source)
}
