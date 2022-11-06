package solver

import (
	"math"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.DBConstraintSolver = (*CopyRotation)(nil)

// NewCopyRotation creates a new CopyRotation constraint solver.
func NewCopyRotation() *CopyRotation {
	return &CopyRotation{}
}

type CopyRotation struct {
	physics.NilDBConstraintSolver

	AxisX bool
	AxisY bool
	AxisZ bool
}

func (s *CopyRotation) CalculateNudges(ctx physics.DBSolverContext) physics.DBNudgeSolution {
	// TODO: Can we run this at the end?
	if s.AxisX {
		primaryX := ctx.Primary.Orientation().OrientationX()
		secondaryX := ctx.Secondary.Orientation().OrientationX()
		s.matchDirections(ctx, primaryX, secondaryX)
	}
	if s.AxisY {
		primaryY := ctx.Primary.Orientation().OrientationY()
		secondaryY := ctx.Secondary.Orientation().OrientationY()
		s.matchDirections(ctx, primaryY, secondaryY)
	}
	if s.AxisZ {
		primaryZ := ctx.Primary.Orientation().OrientationZ()
		secondaryZ := ctx.Secondary.Orientation().OrientationZ()
		s.matchDirections(ctx, primaryZ, secondaryZ)
	}

	return physics.DBNudgeSolution{}
}

func (s *CopyRotation) matchDirections(ctx physics.DBSolverContext, primaryDir, secondaryDir dprec.Vec3) {
	axis := dprec.Vec3Cross(primaryDir, secondaryDir)
	sin := axis.Length()
	cos := dprec.Vec3Dot(primaryDir, secondaryDir)

	angle := dprec.Radians(math.Atan2(sin, cos))
	if dprec.Abs(angle) > 0.00001 {
		rotation := dprec.RotationQuat(-dprec.Abs(angle), dprec.UnitVec3(axis))
		ctx.Secondary.SetOrientation(dprec.UnitQuat(dprec.QuatProd(
			rotation,
			ctx.Secondary.Orientation(),
		)))
	}
}
