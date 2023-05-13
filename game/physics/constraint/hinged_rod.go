package constraint

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

// NewHingedRod creates a new HingedRod constraint solver.
func NewHingedRod() *HingedRod {
	return &HingedRod{
		primaryRadius:   dprec.ZeroVec3(),
		secondaryRadius: dprec.ZeroVec3(),
		length:          1.0,
	}
}

var _ solver.PairConstraint = (*HingedRod)(nil)

// HingedRod represents the solution for a constraint that keeps two bodies
// tied together with a hard link of specific length.
type HingedRod struct {
	primaryRadius   dprec.Vec3
	secondaryRadius dprec.Vec3
	length          float64

	jacobian solver.PairJacobian
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
func (s *HingedRod) Reset(ctx solver.PairContext) {
	primaryRadiusWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.primaryRadius)
	primaryAnchorWS := dprec.Vec3Sum(ctx.Target.Position(), primaryRadiusWS)
	secondaryRadiusWS := dprec.QuatVec3Rotation(ctx.Source.Rotation(), s.secondaryRadius)
	secondaryAnchorWS := dprec.Vec3Sum(ctx.Source.Position(), secondaryRadiusWS)
	deltaPosition := dprec.Vec3Diff(secondaryAnchorWS, primaryAnchorWS)
	if lng := deltaPosition.Length(); lng > solver.Epsilon {
		normal := dprec.Vec3Quot(deltaPosition, lng)
		s.jacobian = solver.PairJacobian{
			Target: solver.Jacobian{
				LinearSlope:  dprec.InverseVec3(normal),
				AngularSlope: dprec.Vec3Cross(normal, primaryRadiusWS),
			},
			Source: solver.Jacobian{
				LinearSlope:  normal,
				AngularSlope: dprec.Vec3Cross(secondaryRadiusWS, normal),
			},
		}
		s.drift = lng - s.length
	} else {
		s.jacobian = solver.PairJacobian{}
		s.drift = 0.0
	}
}

// ApplyImpulses applies impulses in order to keep the velocity part of
// the constraint satisfied.
func (s *HingedRod) ApplyImpulses(ctx solver.PairContext) {
	solution := ctx.JacobianImpulseSolution(s.jacobian, s.drift, 0.0)
	ctx.Target.ApplyImpulse(solution.Target)
	ctx.Source.ApplyImpulse(solution.Source)
}

// ApplyNudges applies nudges in order to keep the positional part of the
// constraint satisfied.
func (s *HingedRod) ApplyNudges(ctx solver.PairContext) {
	solution := ctx.JacobianNudgeSolution(s.jacobian, s.drift)
	ctx.Target.ApplyNudge(solution.Target)
	ctx.Source.ApplyNudge(solution.Source)
}
