package constraint

import (
	"math"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

// NewCopyDirection creates a new CopyDirection constraint solver.
func NewCopyDirection() *CopyDirection {
	return &CopyDirection{
		primaryDirection:   dprec.BasisYVec3(),
		secondaryDirection: dprec.BasisYVec3(),
	}
}

var _ solver.PairConstraint = (*CopyDirection)(nil)

// CopyDirection ensures that the second body has the same direction as
// the first one.
// This solver is immediate - it does not use impulses or nudges.
type CopyDirection struct {
	primaryDirection   dprec.Vec3
	secondaryDirection dprec.Vec3
}

// PrimaryDirection returns the direction of the primary body.
func (s *CopyDirection) PrimaryDirection() dprec.Vec3 {
	return s.primaryDirection
}

// SetPrimaryDirection changes the direction of the primary body.
func (s *CopyDirection) SetPrimaryDirection(direction dprec.Vec3) *CopyDirection {
	s.primaryDirection = dprec.UnitVec3(direction)
	return s
}

// SecondaryDirection returns the direction of the secondary body.
func (s *CopyDirection) SecondaryDirection() dprec.Vec3 {
	return s.secondaryDirection
}

// SetSecondaryDirection changes the direction of the secondary body.
func (s *CopyDirection) SetSecondaryDirection(direction dprec.Vec3) *CopyDirection {
	s.secondaryDirection = dprec.UnitVec3(direction)
	return s
}

func (s *CopyDirection) Reset(ctx solver.PairContext) {}

func (s *CopyDirection) ApplyImpulses(ctx solver.PairContext) {
	// The secondary body will have its direction aligned with the primary body's
	// direction. As such, we need to ensure that the secondary's body angular
	// velocity is only aligned with the primary body's direction (i.e. there is
	// no rotation component that tries to move it away).

	primaryDirWS := dprec.QuatVec3Rotation(ctx.Source.Rotation(), s.primaryDirection)
	angularVelocityAmount := dprec.Vec3Dot(primaryDirWS, ctx.Target.AngularVelocity())
	ctx.Target.SetAngularVelocity(dprec.Vec3Prod(primaryDirWS, angularVelocityAmount))
}

func (s *CopyDirection) ApplyNudges(ctx solver.PairContext) {
	primaryDirWS := dprec.QuatVec3Rotation(ctx.Source.Rotation(), s.primaryDirection)
	secondaryDirWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.secondaryDirection)

	rotationAxis := dprec.Vec3Cross(secondaryDirWS, primaryDirWS)
	cos := dprec.Vec3Dot(secondaryDirWS, primaryDirWS)
	sin := rotationAxis.Length()

	angle := dprec.Abs(dprec.Radians(math.Atan2(sin, cos)))
	if angle > dprec.Angle(solver.Epsilon) {
		rotation := dprec.RotationQuat(angle, dprec.UnitVec3(rotationAxis))
		ctx.Target.SetRotation(dprec.UnitQuat(dprec.QuatProd(
			rotation,
			ctx.Target.Rotation(),
		)))
	}
}
