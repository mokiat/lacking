package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

// NewHingedRod creates a new HingedRod constraint solver.
func NewHingedRod() *HingedRod {
	return &HingedRod{
		primaryAnchor:   dprec.ZeroVec3(),
		secondaryAnchor: dprec.ZeroVec3(),
		length:          1.0,
	}
}

var _ physics.DBConstraintSolver = (*HingedRod)(nil)

// HingedRod represents the solution for a constraint that keeps two bodies
// tied together with a hard link of specific length.
type HingedRod struct {
	primaryAnchor   dprec.Vec3
	secondaryAnchor dprec.Vec3
	length          float64

	jacobian physics.PairJacobian
	drift    float64
}

// PrimaryAnchor returns the attachment point of the link
// on the primary body.
func (r *HingedRod) PrimaryAnchor() dprec.Vec3 {
	return r.primaryAnchor
}

// SetPrimaryAnchor changes the attachment point of the link
// on the primary body.
func (r *HingedRod) SetPrimaryAnchor(anchor dprec.Vec3) *HingedRod {
	r.primaryAnchor = anchor
	return r
}

// SecondaryAnchor returns the attachment point of the link
// on the secondary body.
func (r *HingedRod) SecondaryAnchor() dprec.Vec3 {
	return r.secondaryAnchor
}

// SetSecondaryAnchor changes the attachment point of the link
// on the secondary body.
func (r *HingedRod) SetSecondaryAnchor(anchor dprec.Vec3) *HingedRod {
	r.secondaryAnchor = anchor
	return r
}

// Length returns the link length.
func (r *HingedRod) Length() float64 {
	return r.length
}

// SetLength changes the link length.
func (r *HingedRod) SetLength(length float64) *HingedRod {
	r.length = length
	return r
}

func (r *HingedRod) Reset(ctx physics.DBSolverContext) {
	r.updateJacobian(ctx)
}

func (r *HingedRod) ApplyImpulses(ctx physics.DBSolverContext) {
	ctx.ApplyImpulse(r.jacobian)
}

func (r *HingedRod) ApplyNudges(ctx physics.DBSolverContext) {
	r.updateJacobian(ctx)
	ctx.ApplyNudge(r.jacobian, r.drift)
}

func (r *HingedRod) updateJacobian(ctx physics.DBSolverContext) {
	primaryRadiusWS := dprec.QuatVec3Rotation(ctx.Primary.Orientation(), r.primaryAnchor)
	primaryAnchorWS := dprec.Vec3Sum(ctx.Primary.Position(), primaryRadiusWS)
	secondaryRadiusWS := dprec.QuatVec3Rotation(ctx.Secondary.Orientation(), r.secondaryAnchor)
	secondaryAnchorWS := dprec.Vec3Sum(ctx.Secondary.Position(), secondaryRadiusWS)
	deltaPosition := dprec.Vec3Diff(secondaryAnchorWS, primaryAnchorWS)
	normal := SafeNormal(deltaPosition, dprec.BasisYVec3())
	r.jacobian = physics.PairJacobian{
		Primary: physics.Jacobian{
			SlopeVelocity:        dprec.InverseVec3(normal),
			SlopeAngularVelocity: dprec.Vec3Cross(normal, primaryRadiusWS),
		},
		Secondary: physics.Jacobian{
			SlopeVelocity:        normal,
			SlopeAngularVelocity: dprec.Vec3Cross(secondaryRadiusWS, normal),
		},
	}
	r.drift = deltaPosition.Length() - r.length
}
