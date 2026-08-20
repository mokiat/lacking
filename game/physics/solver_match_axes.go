package physics

import "github.com/mokiat/gomath/dprec"

// MatchAxesSolverConfig holds the parameters with which a
// [MatchAxesSolver] is configured, either through [NewMatchAxesSolver]
// or [MatchAxesSolver.Configure].
type MatchAxesSolverConfig struct {

	// PrimaryBodyAxis is the body-local-space direction, relative to the
	// primary target's rotation, that is driven to become parallel with
	// SecondaryBodyAxis. It need not be unit-length; it is normalized
	// when the solver is configured.
	PrimaryBodyAxis dprec.Vec3

	// SecondaryBodyAxis is the body-local-space direction, relative to
	// the secondary target's rotation, that PrimaryBodyAxis is driven to
	// become parallel with. It need not be unit-length; it is normalized
	// when the solver is configured.
	SecondaryBodyAxis dprec.Vec3
}

// MatchAxesSolver is a [PairConstraintSolver] that rotates its two
// target bodies so that a body-fixed axis on the primary target becomes
// parallel to a body-fixed axis on the secondary target - it aligns the
// two axes as lines, without caring which of the two directions along
// that shared line either axis actually points.
//
// Unlike [CopyRotationSolver], which unconditionally makes the primary
// target track the secondary target's entire rotation, MatchAxesSolver
// only removes the 2 rotational degrees of freedom that would otherwise
// let the two axes drift apart; both targets remain free to spin about
// their now-shared axis, and their relative translation is left
// entirely unconstrained.
//
// A MatchAxesSolver must be configured, either through
// [NewMatchAxesSolver] or [MatchAxesSolver.Configure], before being
// registered with a [Scene] through [PairConstraintView.Create].
type MatchAxesSolver struct {
	primaryBodyAxis   dprec.Vec3
	secondaryBodyAxis dprec.Vec3

	primaryJacobian1   Jacobian
	primaryJacobian2   Jacobian
	secondaryJacobian1 Jacobian
	secondaryJacobian2 Jacobian
	drift1             float64
	drift2             float64
}

var _ PairConstraintSolver = (*MatchAxesSolver)(nil)

// NewMatchAxesSolver creates a new [MatchAxesSolver] configured according
// to config.
func NewMatchAxesSolver(config MatchAxesSolverConfig) *MatchAxesSolver {
	result := &MatchAxesSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config. PrimaryBodyAxis
// and SecondaryBodyAxis are each normalized to unit length.
//
// Configure must be called before this solver is registered with a
// [Scene] through [PairConstraintView.Create]. Unlike
// [NewMatchAxesSolver], it can be called on an already-allocated solver,
// which allows solvers to be cached (e.g. in a slice) and configured on
// demand.
func (s *MatchAxesSolver) Configure(config MatchAxesSolverConfig) {
	s.primaryBodyAxis = dprec.UnitVec3(config.PrimaryBodyAxis)
	s.secondaryBodyAxis = dprec.UnitVec3(config.SecondaryBodyAxis)
}

// PrimaryBodyAxis returns the body-local-space, unit-length direction,
// relative to the primary target's rotation, that is driven to become
// parallel with [MatchAxesSolver.SecondaryBodyAxis].
func (s *MatchAxesSolver) PrimaryBodyAxis() dprec.Vec3 {
	return s.primaryBodyAxis
}

// SetPrimaryBodyAxis changes the body-local-space direction, relative to
// the primary target's rotation, that is driven to become parallel with
// [MatchAxesSolver.SecondaryBodyAxis]. The provided axis need not be
// unit-length; it is normalized before being stored.
//
// It returns the solver itself, so that calls can be chained.
func (s *MatchAxesSolver) SetPrimaryBodyAxis(axis dprec.Vec3) *MatchAxesSolver {
	s.primaryBodyAxis = dprec.UnitVec3(axis)
	return s
}

// SecondaryBodyAxis returns the body-local-space, unit-length direction,
// relative to the secondary target's rotation, that
// [MatchAxesSolver.PrimaryBodyAxis] is driven to become parallel with.
func (s *MatchAxesSolver) SecondaryBodyAxis() dprec.Vec3 {
	return s.secondaryBodyAxis
}

// SetSecondaryBodyAxis changes the body-local-space direction, relative
// to the secondary target's rotation, that
// [MatchAxesSolver.PrimaryBodyAxis] is driven to become parallel with.
// The provided axis need not be unit-length; it is normalized before
// being stored.
//
// It returns the solver itself, so that calls can be chained.
func (s *MatchAxesSolver) SetSecondaryBodyAxis(axis dprec.Vec3) *MatchAxesSolver {
	s.secondaryBodyAxis = dprec.UnitVec3(axis)
	return s
}

// Reset implements [PairConstraintSolver.Reset].
//
// It recomputes the constraint's Jacobians and current alignment error
// (drift) between the two axes, the same way
// [MatchAxesSolver.recompute] does.
func (s *MatchAxesSolver) Reset(ctx PairConstraintContext) {
	s.recompute(ctx)
}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// It resolves impulses, without restitution, that drive the two
// targets' relative angular velocity toward closing the alignment error
// (drift) computed by [MatchAxesSolver.Reset], rotating the two axes
// back toward each other whenever they have drifted apart.
func (s *MatchAxesSolver) ApplyImpulses(ctx PairConstraintContext) {
	primaryImpulse1, secondaryImpulse1 := ctx.ImpulseSolution(
		s.primaryJacobian1,
		s.secondaryJacobian1,
		s.drift1,
		0.0,
	)
	primaryImpulse2, secondaryImpulse2 := ctx.ImpulseSolution(
		s.primaryJacobian2,
		s.secondaryJacobian2,
		s.drift2,
		0.0,
	)

	primaryImpulse := Impulse{
		Linear:  dprec.Vec3Sum(primaryImpulse1.Linear, primaryImpulse2.Linear),
		Angular: dprec.Vec3Sum(primaryImpulse1.Angular, primaryImpulse2.Angular),
	}
	secondaryImpulse := Impulse{
		Linear:  dprec.Vec3Sum(secondaryImpulse1.Linear, secondaryImpulse2.Linear),
		Angular: dprec.Vec3Sum(secondaryImpulse1.Angular, secondaryImpulse2.Angular),
	}

	ctx.PrimaryTarget.ApplyImpulse(primaryImpulse)
	ctx.SecondaryTarget.ApplyImpulse(secondaryImpulse)
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It first recomputes the constraint's Jacobians and current alignment
// error (drift), the same way [MatchAxesSolver.recompute] does, since a
// preceding nudge - by this solver's own previous iteration, or by
// another constraint acting on either target - may have rotated either
// target since [MatchAxesSolver.Reset] or the last call to this method.
// It then nudges both targets' rotations to reduce any remaining
// alignment error between the two axes.
func (s *MatchAxesSolver) ApplyNudges(ctx PairConstraintContext) {
	s.recompute(ctx)

	primaryNudge1, secondaryNudge1 := ctx.NudgeSolution(
		s.primaryJacobian1,
		s.secondaryJacobian1,
		s.drift1,
	)
	primaryNudge2, secondaryNudge2 := ctx.NudgeSolution(
		s.primaryJacobian2,
		s.secondaryJacobian2,
		s.drift2,
	)

	primaryNudge := Nudge{
		Linear:  dprec.Vec3Sum(primaryNudge1.Linear, primaryNudge2.Linear),
		Angular: dprec.Vec3Sum(primaryNudge1.Angular, primaryNudge2.Angular),
	}
	secondaryNudge := Nudge{
		Linear:  dprec.Vec3Sum(secondaryNudge1.Linear, secondaryNudge2.Linear),
		Angular: dprec.Vec3Sum(secondaryNudge1.Angular, secondaryNudge2.Angular),
	}

	ctx.PrimaryTarget.ApplyNudge(primaryNudge)
	ctx.SecondaryTarget.ApplyNudge(secondaryNudge)
}

// recompute recalculates the constraint's Jacobians, along with the
// world-space direction of each target's axis (derived from
// PrimaryBodyAxis and SecondaryBodyAxis combined with each target's
// current rotation), and the current alignment error (drift) between
// the two axes, based on the targets' current rotations.
//
// The alignment error is measured as the two components of the primary
// axis along secondaryAxisNorm1 and secondaryAxisNorm2, an arbitrary
// orthonormal basis of the plane perpendicular to the secondary axis -
// both components are zero exactly when the primary axis is parallel
// (or antiparallel) to the secondary axis, regardless of which
// orthonormal basis of that plane is chosen.
func (s *MatchAxesSolver) recompute(ctx PairConstraintContext) {
	primaryAxisWS := dprec.QuatVec3Rotation(ctx.PrimaryTarget.Rotation(), s.primaryBodyAxis)
	secondaryAxisWS := dprec.QuatVec3Rotation(ctx.SecondaryTarget.Rotation(), s.secondaryBodyAxis)
	secondaryAxisNorm1 := dprec.NormalVec3(secondaryAxisWS)
	secondaryAxisNorm2 := dprec.Vec3Cross(secondaryAxisWS, secondaryAxisNorm1)

	s.primaryJacobian1 = Jacobian{
		LinearSlope:  dprec.ZeroVec3(),
		AngularSlope: dprec.Vec3Cross(secondaryAxisNorm1, primaryAxisWS),
	}
	s.secondaryJacobian1 = Jacobian{
		LinearSlope:  dprec.ZeroVec3(),
		AngularSlope: dprec.Vec3Cross(primaryAxisWS, secondaryAxisNorm1),
	}

	s.primaryJacobian2 = Jacobian{
		LinearSlope:  dprec.ZeroVec3(),
		AngularSlope: dprec.Vec3Cross(secondaryAxisNorm2, primaryAxisWS),
	}
	s.secondaryJacobian2 = Jacobian{
		LinearSlope:  dprec.ZeroVec3(),
		AngularSlope: dprec.Vec3Cross(primaryAxisWS, secondaryAxisNorm2),
	}

	s.drift1 = dprec.Vec3Dot(primaryAxisWS, secondaryAxisNorm1)
	s.drift2 = dprec.Vec3Dot(primaryAxisWS, secondaryAxisNorm2)
}
