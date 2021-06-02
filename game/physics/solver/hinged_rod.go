package solver

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.DBConstraintSolver = (*HingedRod)(nil)

// NewHingedRod creates a new HingedRod constraint solution.
func NewHingedRod() *HingedRod {
	result := &HingedRod{
		length: 1.0,
	}
	result.DBJacobianConstraintSolver = physics.NewDBJacobianConstraintSolver(result.calculate)
	return result
}

// HingedRod represents the solution for a constraint
// that keeps two bodies tied together with a hard link
// of specific length.
type HingedRod struct {
	*physics.DBJacobianConstraintSolver

	primaryAnchor   sprec.Vec3
	secondaryAnchor sprec.Vec3
	length          float32
}

// PrimaryAnchor returns the attachment point of the link
// on the primary body.
func (r *HingedRod) PrimaryAnchor() sprec.Vec3 {
	return r.primaryAnchor
}

// SetPrimaryAnchor changes the attachment point of the link
// on the primary body.
func (r *HingedRod) SetPrimaryAnchor(anchor sprec.Vec3) *HingedRod {
	r.primaryAnchor = anchor
	return r
}

// SecondaryAnchor returns the attachment point of the link
// on the secondary body.
func (r *HingedRod) SecondaryAnchor() sprec.Vec3 {
	return r.secondaryAnchor
}

// SetSecondaryAnchor changes the attachment point of the link
// on the secondary body.
func (r *HingedRod) SetSecondaryAnchor(anchor sprec.Vec3) *HingedRod {
	r.secondaryAnchor = anchor
	return r
}

// Length returns the link length.
func (r *HingedRod) Length() float32 {
	return r.length
}

// SetLength changes the link length.
func (r *HingedRod) SetLength(length float32) *HingedRod {
	r.length = length
	return r
}

func (r *HingedRod) calculate(ctx physics.DBSolverContext) (physics.PairJacobian, float32) {
	firstRadiusWS := sprec.QuatVec3Rotation(ctx.Primary.Orientation(), r.primaryAnchor)
	secondRadiusWS := sprec.QuatVec3Rotation(ctx.Secondary.Orientation(), r.secondaryAnchor)
	firstAnchorWS := sprec.Vec3Sum(ctx.Primary.Position(), firstRadiusWS)
	secondAnchorWS := sprec.Vec3Sum(ctx.Secondary.Position(), secondRadiusWS)
	deltaPosition := sprec.Vec3Diff(secondAnchorWS, firstAnchorWS)
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
				SlopeAngularVelocity: sprec.NewVec3(
					normal.Z*secondRadiusWS.Y-normal.Y*secondRadiusWS.Z,
					normal.X*secondRadiusWS.Z-normal.Z*secondRadiusWS.X,
					normal.Y*secondRadiusWS.X-normal.X*secondRadiusWS.Y,
				),
			},
		},
		deltaPosition.Length() - r.length
}
