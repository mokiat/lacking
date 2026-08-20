package physics

import "github.com/mokiat/gomath/dprec"

// AxisRangeSolverConfig holds the parameters with which an
// [AxisRangeSolver] is configured, either through [NewAxisRangeSolver] or
// [AxisRangeSolver.Configure].
type AxisRangeSolverConfig struct {

	// PrimaryBodyAnchorOffset is the body-local-space offset, relative to
	// the primary target's center of mass, of the point at which the
	// distance constraint is anchored on the primary body.
	PrimaryBodyAnchorOffset dprec.Vec3

	// PrimaryBodyAxis is the body-local-space direction, relative to the
	// primary target's rotation, along which the distance between the two
	// anchor points is measured. It need not be unit-length; it is
	// normalized when the solver is configured.
	PrimaryBodyAxis dprec.Vec3

	// SecondaryBodyAnchorOffset is the body-local-space offset, relative
	// to the secondary target's center of mass, of the point at which
	// the distance constraint is anchored on the secondary body.
	SecondaryBodyAnchorOffset dprec.Vec3

	// MinDisplacement is the lowest permitted signed distance, measured along
	// PrimaryBodyAxis, between the two anchor points.
	MinDisplacement float64

	// MaxDisplacement is the highest permitted signed distance, measured
	// along PrimaryBodyAxis, between the two anchor points.
	MaxDisplacement float64

	// RestitutionCoefficient is the bounciness applied when the
	// axis-aligned distance reaches MinDisplacement or MaxDisplacement.
	// Negative values are clamped to zero.
	RestitutionCoefficient float64
}

// AxisRangeSolver is a [PairConstraintSolver] that keeps the signed
// distance between an anchor point on each of its two target bodies,
// measured along an axis fixed to the primary body, within a
// [AxisRangeSolver.MinDisplacement] and [AxisRangeSolver.MaxDisplacement]
// range - acting like a rigid rod between the two only once one of the range's
// limits is reached, and applying no force while the axis-aligned distance
// is within range.
//
// Unlike [DistanceSolver], which holds the anchor points at a fixed
// distance apart measured along the direction connecting them,
// AxisRangeSolver measures the distance along a single axis fixed to the
// primary body's orientation and only constrains that distance to stay
// within a range, rather than at a fixed value.
//
// An AxisRangeSolver must be configured, either through
// [NewAxisRangeSolver] or [AxisRangeSolver.Configure], before being
// registered with a [Scene] through [PairConstraintView.Create].
type AxisRangeSolver struct {
	primaryBodyAnchorOffset   dprec.Vec3
	primaryBodyAxis           dprec.Vec3
	secondaryBodyAnchorOffset dprec.Vec3
	minDisplacement           float64
	maxDisplacement           float64
	restitutionCoefficient    float64

	primaryJacobian   Jacobian
	secondaryJacobian Jacobian
	drift             float64
}

var _ PairConstraintSolver = (*AxisRangeSolver)(nil)

// NewAxisRangeSolver creates a new [AxisRangeSolver] configured according
// to config.
func NewAxisRangeSolver(config AxisRangeSolverConfig) *AxisRangeSolver {
	result := &AxisRangeSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config. PrimaryBodyAxis
// is normalized, as with [AxisRangeSolver.SetPrimaryBodyAxis], and
// negative RestitutionCoefficient values are clamped to zero, as with
// [AxisRangeSolver.SetRestitutionCoefficient].
//
// Configure must be called before this solver is registered with a
// [Scene] through [PairConstraintView.Create]. Unlike
// [NewAxisRangeSolver], it can be called on an already-allocated solver,
// which allows solvers to be cached (e.g. in a slice) and configured on
// demand.
func (s *AxisRangeSolver) Configure(config AxisRangeSolverConfig) {
	s.primaryBodyAnchorOffset = config.PrimaryBodyAnchorOffset
	s.primaryBodyAxis = dprec.UnitVec3(config.PrimaryBodyAxis)
	s.secondaryBodyAnchorOffset = config.SecondaryBodyAnchorOffset
	s.minDisplacement = config.MinDisplacement
	s.maxDisplacement = config.MaxDisplacement
	s.restitutionCoefficient = max(0.0, config.RestitutionCoefficient)
}

// PrimaryBodyAnchorOffset returns the body-local-space offset, relative
// to the primary target's center of mass, of the point at which the
// distance constraint is anchored on the primary body.
func (s *AxisRangeSolver) PrimaryBodyAnchorOffset() dprec.Vec3 {
	return s.primaryBodyAnchorOffset
}

// SetPrimaryBodyAnchorOffset changes the body-local-space offset,
// relative to the primary target's center of mass, of the point at
// which the distance constraint is anchored on the primary body.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisRangeSolver) SetPrimaryBodyAnchorOffset(offset dprec.Vec3) *AxisRangeSolver {
	s.primaryBodyAnchorOffset = offset
	return s
}

// PrimaryBodyAxis returns the body-local-space, unit-length direction,
// relative to the primary target's rotation, along which the distance
// between the two anchor points is measured.
func (s *AxisRangeSolver) PrimaryBodyAxis() dprec.Vec3 {
	return s.primaryBodyAxis
}

// SetPrimaryBodyAxis changes the body-local-space direction, relative to
// the primary target's rotation, along which the distance between the
// two anchor points is measured. The provided axis need not be
// unit-length; it is normalized before being stored.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisRangeSolver) SetPrimaryBodyAxis(axis dprec.Vec3) *AxisRangeSolver {
	s.primaryBodyAxis = dprec.UnitVec3(axis)
	return s
}

// SecondaryBodyAnchorOffset returns the body-local-space offset,
// relative to the secondary target's center of mass, of the point at
// which the distance constraint is anchored on the secondary body.
func (s *AxisRangeSolver) SecondaryBodyAnchorOffset() dprec.Vec3 {
	return s.secondaryBodyAnchorOffset
}

// SetSecondaryBodyAnchorOffset changes the body-local-space offset,
// relative to the secondary target's center of mass, of the point at
// which the distance constraint is anchored on the secondary body.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisRangeSolver) SetSecondaryBodyAnchorOffset(offset dprec.Vec3) *AxisRangeSolver {
	s.secondaryBodyAnchorOffset = offset
	return s
}

// RestitutionCoefficient returns the bounciness applied when the
// axis-aligned distance reaches [AxisRangeSolver.MinDisplacement] or
// [AxisRangeSolver.MaxDisplacement].
func (s *AxisRangeSolver) RestitutionCoefficient() float64 {
	return s.restitutionCoefficient
}

// SetRestitutionCoefficient changes the bounciness applied when the
// axis-aligned distance reaches [AxisRangeSolver.MinDisplacement] or
// [AxisRangeSolver.MaxDisplacement]. Negative values are clamped to zero.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisRangeSolver) SetRestitutionCoefficient(coefficient float64) *AxisRangeSolver {
	s.restitutionCoefficient = max(0.0, coefficient)
	return s
}

// MinDisplacement returns the lowest permitted signed distance, measured
// along [AxisRangeSolver.PrimaryBodyAxis], between the two anchor
// points.
func (s *AxisRangeSolver) MinDisplacement() float64 {
	return s.minDisplacement
}

// SetMinDisplacement changes the lowest permitted signed distance,
// measured along [AxisRangeSolver.PrimaryBodyAxis], between the two
// anchor points.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisRangeSolver) SetMinDisplacement(min float64) *AxisRangeSolver {
	s.minDisplacement = min
	return s
}

// MaxDisplacement returns the highest permitted signed distance, measured
// along [AxisRangeSolver.PrimaryBodyAxis], between the two anchor
// points.
func (s *AxisRangeSolver) MaxDisplacement() float64 {
	return s.maxDisplacement
}

// SetMaxDisplacement changes the highest permitted signed distance,
// measured along [AxisRangeSolver.PrimaryBodyAxis], between the two
// anchor points.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisRangeSolver) SetMaxDisplacement(max float64) *AxisRangeSolver {
	s.maxDisplacement = max
	return s
}

// Reset implements [PairConstraintSolver.Reset].
//
// It recomputes the constraint's primary and secondary [Jacobian]s and
// current range violation (drift), the same way
// [AxisRangeSolver.recompute] does.
func (s *AxisRangeSolver) Reset(ctx PairConstraintContext) {
	s.recompute(ctx)
}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// If the axis-aligned distance between the two anchor points, as of the
// last call to [AxisRangeSolver.Reset] or
// [AxisRangeSolver.ApplyNudges], is within the
// [AxisRangeSolver.MinDisplacement]/[AxisRangeSolver.MaxDisplacement]
// range, it does nothing. Otherwise, it resolves a pair of impulses, combining
// restitution with Baumgarte positional-drift stabilization, that drive
// the anchor points' relative velocity toward bringing the axis-aligned
// distance back within range. If the two targets are already moving
// apart (back towards the permitted range), it returns without applying
// anything, leaving any remaining violation to
// [AxisRangeSolver.ApplyNudges].
func (s *AxisRangeSolver) ApplyImpulses(ctx PairConstraintContext) {
	if s.drift == 0.0 {
		return // no constraint violation
	}

	bounceLambda, baumgarteLambda := ctx.ImpulseLambdaComponents(s.primaryJacobian, s.secondaryJacobian, s.drift, s.restitutionCoefficient)
	if bounceLambda < 0.0 {
		return // moving away
	}

	lambda := bounceLambda + baumgarteLambda
	primaryImpulse := s.primaryJacobian.Impulse(lambda)
	secondaryImpulse := s.secondaryJacobian.Impulse(lambda)

	ctx.PrimaryTarget.ApplyImpulse(primaryImpulse)
	ctx.SecondaryTarget.ApplyImpulse(secondaryImpulse)
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It first recomputes the constraint's primary and secondary [Jacobian]s
// and current range violation (drift), the same way
// [AxisRangeSolver.recompute] does, since a preceding nudge - by this
// solver's own previous iteration, or by another constraint acting on
// either target - may have moved either target since
// [AxisRangeSolver.Reset] or the last call to this method. If the
// axis-aligned distance is within range, it does nothing; otherwise, it
// nudges both targets' positions and rotations to bring the axis-aligned
// distance back within the
// [AxisRangeSolver.MinDisplacement]/[AxisRangeSolver.MaxDisplacement]
// range.
func (s *AxisRangeSolver) ApplyNudges(ctx PairConstraintContext) {
	s.recompute(ctx)

	if s.drift > 0.0 {
		primaryNudge, secondaryNudge := ctx.NudgeSolution(
			s.primaryJacobian, s.secondaryJacobian, s.drift,
		)
		ctx.PrimaryTarget.ApplyNudge(primaryNudge)
		ctx.SecondaryTarget.ApplyNudge(secondaryNudge)
	}
}

// recompute recalculates the constraint's primary and secondary
// [Jacobian]s, along with the current range violation (drift), based on
// the targets' current positions and rotations.
//
// It projects the offset between the two anchor points onto
// PrimaryBodyAxis, transformed into world space through the primary
// target's current rotation, to obtain the axis-aligned displacement. If
// that displacement is below MinDisplacement, the Jacobians and drift
// are set up to push the anchor points apart along the axis; if it is
// above MaxDisplacement, they are set up to pull the anchor points
// together along the axis; otherwise both Jacobians and the drift are
// reset to zero, so that [AxisRangeSolver.ApplyImpulses] and
// [AxisRangeSolver.ApplyNudges] apply no correction.
func (s *AxisRangeSolver) recompute(ctx PairConstraintContext) {
	primaryAnchorOffsetWS := dprec.QuatVec3Rotation(ctx.PrimaryTarget.Rotation(), s.primaryBodyAnchorOffset)
	primaryAnchorWS := dprec.Vec3Sum(ctx.PrimaryTarget.Position(), primaryAnchorOffsetWS)

	secondaryAnchorOffsetWS := dprec.QuatVec3Rotation(ctx.SecondaryTarget.Rotation(), s.secondaryBodyAnchorOffset)
	secondaryAnchorWS := dprec.Vec3Sum(ctx.SecondaryTarget.Position(), secondaryAnchorOffsetWS)

	axisWS := dprec.QuatVec3Rotation(ctx.PrimaryTarget.Rotation(), s.primaryBodyAxis)

	delta := dprec.Vec3Diff(secondaryAnchorWS, primaryAnchorWS)
	actualDisplacement := dprec.Vec3Dot(axisWS, delta)

	switch {
	case actualDisplacement < s.minDisplacement:
		s.primaryJacobian = Jacobian{
			LinearSlope:  dprec.InverseVec3(axisWS),
			AngularSlope: dprec.Vec3Cross(axisWS, primaryAnchorOffsetWS),
		}
		s.secondaryJacobian = Jacobian{
			LinearSlope:  axisWS,
			AngularSlope: dprec.Vec3Cross(secondaryAnchorOffsetWS, axisWS),
		}
		s.drift = s.minDisplacement - actualDisplacement

	case actualDisplacement > s.maxDisplacement:
		s.primaryJacobian = Jacobian{
			LinearSlope:  axisWS,
			AngularSlope: dprec.Vec3Cross(primaryAnchorOffsetWS, axisWS),
		}
		s.secondaryJacobian = Jacobian{
			LinearSlope:  dprec.InverseVec3(axisWS),
			AngularSlope: dprec.Vec3Cross(axisWS, secondaryAnchorOffsetWS),
		}
		s.drift = actualDisplacement - s.maxDisplacement

	default:
		s.primaryJacobian = Jacobian{}
		s.secondaryJacobian = Jacobian{}
		s.drift = 0.0
	}
}
