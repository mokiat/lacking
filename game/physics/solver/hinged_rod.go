package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

// NewHingedRod creates a new HingedRod constraint solver.
func NewHingedRod() *HingedRod {
	return &HingedRod{
		primaryRadius:   dprec.ZeroVec3(),
		secondaryRadius: dprec.ZeroVec3(),
		length:          1.0,
	}
}

var _ physics.DBConstraintSolver = (*HingedRod)(nil)

// HingedRod represents the solution for a constraint that keeps two bodies
// tied together with a hard link of specific length.
type HingedRod struct {
	primaryRadius   dprec.Vec3
	secondaryRadius dprec.Vec3
	length          float64

	jacobian physics.PairJacobian
	drift    float64
}

// PrimaryRadius returns the radius vector of the contact point
// on the primary object.
//
// The vector is in the object's local space.
func (s *HingedRod) PrimaryRadius() dprec.Vec3 {
	return s.primaryRadius
}

// SetPrimaryRadius changes the attachment point of the link
// on the primary body.
func (s *HingedRod) SetPrimaryRadius(radius dprec.Vec3) *HingedRod {
	s.primaryRadius = radius
	return s
}

// SecondaryRadius returns the radius vector of the contact point
// on the secondary object.
//
// The vector is in the object's local space.
func (s *HingedRod) SecondaryRadius() dprec.Vec3 {
	return s.secondaryRadius
}

// SetSecondaryRadius changes the radius vector of the contact point
// on the secondary object.
//
// The vector is in the object's local space.
func (s *HingedRod) SetSecondaryRadius(radius dprec.Vec3) *HingedRod {
	s.secondaryRadius = radius
	return s
}

// Length returns the link length.
func (s *HingedRod) Length() float64 {
	return s.length
}

// SetLength changes the link length.
func (s *HingedRod) SetLength(length float64) *HingedRod {
	s.length = length
	return s
}

// Reset re-evaluates the constraint.
func (s *HingedRod) Reset(ctx physics.DBSolverContext) {
	primaryRadiusWS := dprec.QuatVec3Rotation(ctx.Primary.Orientation(), s.primaryRadius)
	primaryAnchorWS := dprec.Vec3Sum(ctx.Primary.Position(), primaryRadiusWS)
	secondaryRadiusWS := dprec.QuatVec3Rotation(ctx.Secondary.Orientation(), s.secondaryRadius)
	secondaryAnchorWS := dprec.Vec3Sum(ctx.Secondary.Position(), secondaryRadiusWS)
	deltaPosition := dprec.Vec3Diff(secondaryAnchorWS, primaryAnchorWS)
	normal := SafeNormal(deltaPosition, dprec.BasisYVec3())
	s.jacobian = physics.PairJacobian{
		Primary: physics.Jacobian{
			SlopeVelocity:        dprec.InverseVec3(normal),
			SlopeAngularVelocity: dprec.Vec3Cross(normal, primaryRadiusWS),
		},
		Secondary: physics.Jacobian{
			SlopeVelocity:        normal,
			SlopeAngularVelocity: dprec.Vec3Cross(secondaryRadiusWS, normal),
		},
	}
	s.drift = deltaPosition.Length() - s.length
}

// ApplyImpulses applies impulses in order to keep the velocity part of
// the constraint satisfied.
func (s *HingedRod) ApplyImpulses(ctx physics.DBSolverContext) {
	solution := ctx.JacobianImpulseSolution(s.jacobian, s.drift, 0.0)
	ctx.ApplyImpulseSolution(solution)
}

// ApplyNudges applies nudges in order to keep the positional part of the
// constraint satisfied.
func (s *HingedRod) ApplyNudges(ctx physics.DBSolverContext) {
	solution := ctx.JacobianNudgeSolution(s.jacobian, s.drift)
	ctx.ApplyNudgeSolution(solution)
}
