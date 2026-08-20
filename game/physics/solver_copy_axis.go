package physics

import "github.com/mokiat/gomath/dprec"

// CopyAxisSolverConfig holds the parameters with which a
// [CopyAxisSolver] is configured, either through [NewCopyAxisSolver] or
// [CopyAxisSolver.Configure].
type CopyAxisSolverConfig struct {

	// PrimaryBodyAxis is the body-local-space direction, relative to the
	// primary target's rotation, that is driven to point in the same
	// direction as SecondaryBodyAxis. It need not be unit-length; it is
	// normalized when the solver is configured, so it must not be the
	// zero vector.
	PrimaryBodyAxis dprec.Vec3

	// SecondaryBodyAxis is the body-local-space direction, relative to
	// the secondary target's rotation, that PrimaryBodyAxis is driven to
	// point in the same direction as. It need not be unit-length; it is
	// normalized when the solver is configured, so it must not be the
	// zero vector.
	SecondaryBodyAxis dprec.Vec3
}

// CopyAxisSolver is a [PairConstraintSolver] that makes a body-fixed
// axis on the primary target point in the same direction as a
// body-fixed axis on the secondary target, regardless of any forces or
// torques acting on the primary target.
//
// It is a kinematic constraint, in the same vein as
// [CopyRotationSolver], except that it copies the direction of a single
// axis instead of the secondary target's entire rotation. The primary
// target is left free to spin about that shared axis, keeping whatever
// spin it already had about it; only the 2 rotational degrees of
// freedom that would otherwise let the two axes drift apart are taken
// away. The primary target's position and linear velocity are left
// untouched, and the secondary target is never modified at all - see
// [CopyPositionSolver] for the positional counterpart.
//
// Unlike [MatchAxesSolver], which aligns the two axes as lines and
// rotates both targets in proportion to their inertia, CopyAxisSolver
// aligns them as directions - the primary axis ends up pointing the
// same way as the secondary axis, not merely along the same line - and
// achieves that by driving the primary target alone. Consequently, two
// exactly antiparallel axes are not an equilibrium for this solver, but
// they are an ambiguous configuration: the primary target is turned
// halfway around an arbitrary axis perpendicular to it, so callers that
// can reach that state should not rely on which way the primary target
// swings through it.
//
// A CopyAxisSolver must be configured, either through
// [NewCopyAxisSolver] or [CopyAxisSolver.Configure], before being
// registered with a [Scene] through [PairConstraintView.Create].
type CopyAxisSolver struct {
	primaryBodyAxis   dprec.Vec3
	secondaryBodyAxis dprec.Vec3
}

var _ PairConstraintSolver = (*CopyAxisSolver)(nil)

// NewCopyAxisSolver creates a new [CopyAxisSolver] configured according
// to config.
func NewCopyAxisSolver(config CopyAxisSolverConfig) *CopyAxisSolver {
	result := &CopyAxisSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config. PrimaryBodyAxis
// and SecondaryBodyAxis are each normalized to unit length.
//
// Configure must be called before this solver is registered with a
// [Scene] through [PairConstraintView.Create]. Unlike
// [NewCopyAxisSolver], it can be called on an already-allocated solver,
// which allows solvers to be cached (e.g. in a slice) and configured on
// demand.
func (s *CopyAxisSolver) Configure(config CopyAxisSolverConfig) {
	s.primaryBodyAxis = dprec.UnitVec3(config.PrimaryBodyAxis)
	s.secondaryBodyAxis = dprec.UnitVec3(config.SecondaryBodyAxis)
}

// PrimaryBodyAxis returns the body-local-space, unit-length direction,
// relative to the primary target's rotation, that is driven to point in
// the same direction as [CopyAxisSolver.SecondaryBodyAxis].
func (s *CopyAxisSolver) PrimaryBodyAxis() dprec.Vec3 {
	return s.primaryBodyAxis
}

// SetPrimaryBodyAxis changes the body-local-space direction, relative to
// the primary target's rotation, that is driven to point in the same
// direction as [CopyAxisSolver.SecondaryBodyAxis]. The provided axis
// need not be unit-length; it is normalized before being stored.
//
// It returns the solver itself, so that calls can be chained.
func (s *CopyAxisSolver) SetPrimaryBodyAxis(axis dprec.Vec3) *CopyAxisSolver {
	s.primaryBodyAxis = dprec.UnitVec3(axis)
	return s
}

// SecondaryBodyAxis returns the body-local-space, unit-length direction,
// relative to the secondary target's rotation, that
// [CopyAxisSolver.PrimaryBodyAxis] is driven to point in the same
// direction as.
func (s *CopyAxisSolver) SecondaryBodyAxis() dprec.Vec3 {
	return s.secondaryBodyAxis
}

// SetSecondaryBodyAxis changes the body-local-space direction, relative
// to the secondary target's rotation, that
// [CopyAxisSolver.PrimaryBodyAxis] is driven to point in the same
// direction as. The provided axis need not be unit-length; it is
// normalized before being stored.
//
// It returns the solver itself, so that calls can be chained.
func (s *CopyAxisSolver) SetSecondaryBodyAxis(axis dprec.Vec3) *CopyAxisSolver {
	s.secondaryBodyAxis = dprec.UnitVec3(axis)
	return s
}

// Reset implements [PairConstraintSolver.Reset].
//
// It is a no-op, since this solver holds no per-step state that needs to
// be derived from the targets' current rotations or velocities.
func (s *CopyAxisSolver) Reset(ctx PairConstraintContext) {}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// It unconditionally overwrites the primary target's angular velocity
// with the one that keeps the two axes from drifting apart: the
// secondary target's angular velocity about all directions
// perpendicular to the axis, combined with the primary target's own,
// preserved rotational speed about the axis itself. The secondary
// target's rotational speed about the axis is deliberately not copied,
// since spinning about an axis does not move that axis.
func (s *CopyAxisSolver) ApplyImpulses(ctx PairConstraintContext) {
	// The primary body has its axis aligned with the secondary body's axis.
	// The axis is dragged along by whatever angular velocity the two bodies
	// do not have in common, so to keep the two aligned it is the relative
	// angular velocity - not the primary body's own - that has to be
	// collinear with the axis (i.e. there is no rotational component that
	// tries to move the axes apart).
	//
	// That means the primary body has to take over the secondary body's
	// angular velocity perpendicular to the axis, while keeping its own
	// rotational speed about the axis, since spinning about an axis does
	// not move that axis.
	//
	// Note that dprec.Vec3Projection flattens the vector onto the plane
	// described by the normal, so it yields the perpendicular component.

	secondaryAxisWS := dprec.QuatVec3Rotation(ctx.SecondaryTarget.Rotation(), s.secondaryBodyAxis)
	secondaryOrthogonalAngularVelocity := dprec.Vec3Projection(ctx.SecondaryTarget.AngularVelocity(), secondaryAxisWS)

	result := dprec.Vec3Prod(secondaryAxisWS, dprec.Vec3Dot(ctx.PrimaryTarget.AngularVelocity(), secondaryAxisWS))
	result = dprec.Vec3Sum(result, secondaryOrthogonalAngularVelocity)
	ctx.PrimaryTarget.SetAngularVelocity(result)
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It unconditionally rotates the primary target by the smallest
// rotation that brings its axis onto the secondary target's, which
// leaves the primary target's spin about the resulting shared axis
// unchanged. Since the correction is applied in full, rather than being
// scaled by [PairConstraintContext.NudgeBeta], a single call is enough
// to align the two axes exactly; any subsequent call within the same
// step merely counteracts whatever misalignment other constraints have
// introduced in the meantime.
//
// If the two axes happen to be exactly antiparallel, the smallest
// rotation is not unique - the correction is then a half turn about an
// arbitrary axis perpendicular to the primary axis.
func (s *CopyAxisSolver) ApplyNudges(ctx PairConstraintContext) {
	primaryAxisWS := dprec.QuatVec3Rotation(ctx.PrimaryTarget.Rotation(), s.primaryBodyAxis)
	secondaryAxisWS := dprec.QuatVec3Rotation(ctx.SecondaryTarget.Rotation(), s.secondaryBodyAxis)

	rotationAxis := dprec.Vec3Cross(primaryAxisWS, secondaryAxisWS)
	if rotationAxis.SqrLength() < Epsilon*Epsilon {
		rotationAxis = dprec.NormalVec3(primaryAxisWS)
	}
	rotationAngle := dprec.Vec3Angle(primaryAxisWS, secondaryAxisWS)

	rotation := dprec.RotationQuat(rotationAngle, rotationAxis)
	ctx.PrimaryTarget.SetRotation(dprec.QuatProd(
		rotation,
		ctx.PrimaryTarget.Rotation(),
	))
}
