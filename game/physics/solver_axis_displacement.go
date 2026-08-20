package physics

import "github.com/mokiat/gomath/dprec"

// AxisDisplacementSolverConfig holds the parameters with which an
// [AxisDisplacementSolver] is configured, either through
// [NewAxisDisplacementSolver] or [AxisDisplacementSolver.Configure].
type AxisDisplacementSolverConfig struct {

	// PrimaryBodyAnchorOffset is the body-local-space offset, relative to
	// the primary target's center of mass, of the point at which the
	// displacement constraint is anchored on the primary body.
	PrimaryBodyAnchorOffset dprec.Vec3

	// PrimaryBodyAxis is the body-local-space direction, relative to the
	// primary target's rotation, along which the displacement between the
	// two anchor points is measured. It need not be unit-length; it is
	// normalized when the solver is configured.
	PrimaryBodyAxis dprec.Vec3

	// SecondaryBodyAnchorOffset is the body-local-space offset, relative
	// to the secondary target's center of mass, of the point at which
	// the displacement constraint is anchored on the secondary body.
	SecondaryBodyAnchorOffset dprec.Vec3

	// Displacement is the signed distance, measured along
	// PrimaryBodyAxis, at which the two anchor points are held apart.
	// Unlike a plain Euclidean distance, it may be negative.
	Displacement float64
}

// AxisDisplacementSolver is a [PairConstraintSolver] that holds the
// signed distance between an anchor point on each of its two target
// bodies, measured along an axis fixed to the primary body, at a fixed
// [AxisDisplacementSolver.Displacement] - acting like a rigid rod
// between the two, constrained to slide only along that axis.
//
// Unlike [DistanceSolver], which holds the anchor points at a fixed
// distance apart measured along the direction connecting them,
// AxisDisplacementSolver measures and constrains that distance along a
// single axis fixed to the primary body's orientation, leaving any
// separation between the anchor points perpendicular to that axis
// unconstrained.
//
// Unlike [AxisRangeSolver], which only engages once the axis-aligned
// distance leaves a [AxisRangeSolver.MinDisplacement]/
// [AxisRangeSolver.MaxDisplacement] range, AxisDisplacementSolver always
// enforces a single, exact displacement.
//
// An AxisDisplacementSolver must be configured, either through
// [NewAxisDisplacementSolver] or [AxisDisplacementSolver.Configure],
// before being registered with a [Scene] through
// [PairConstraintView.Create].
type AxisDisplacementSolver struct {
	primaryBodyAnchorOffset   dprec.Vec3
	primaryBodyAxis           dprec.Vec3
	secondaryBodyAnchorOffset dprec.Vec3
	displacement              float64

	primaryJacobian   Jacobian
	secondaryJacobian Jacobian
	drift             float64
}

var _ PairConstraintSolver = (*AxisDisplacementSolver)(nil)

// NewAxisDisplacementSolver creates a new [AxisDisplacementSolver]
// configured according to config.
func NewAxisDisplacementSolver(config AxisDisplacementSolverConfig) *AxisDisplacementSolver {
	result := &AxisDisplacementSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config. PrimaryBodyAxis
// is normalized, as with [AxisDisplacementSolver.SetPrimaryBodyAxis].
//
// Configure must be called before this solver is registered with a
// [Scene] through [PairConstraintView.Create]. Unlike
// [NewAxisDisplacementSolver], it can be called on an already-allocated
// solver, which allows solvers to be cached (e.g. in a slice) and
// configured on demand.
func (s *AxisDisplacementSolver) Configure(config AxisDisplacementSolverConfig) {
	s.primaryBodyAnchorOffset = config.PrimaryBodyAnchorOffset
	s.primaryBodyAxis = dprec.UnitVec3(config.PrimaryBodyAxis)
	s.secondaryBodyAnchorOffset = config.SecondaryBodyAnchorOffset
	s.displacement = config.Displacement
}

// PrimaryBodyAnchorOffset returns the body-local-space offset, relative
// to the primary target's center of mass, of the point at which the
// displacement constraint is anchored on the primary body.
func (s *AxisDisplacementSolver) PrimaryBodyAnchorOffset() dprec.Vec3 {
	return s.primaryBodyAnchorOffset
}

// SetPrimaryBodyAnchorOffset changes the body-local-space offset,
// relative to the primary target's center of mass, of the point at
// which the displacement constraint is anchored on the primary body.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisDisplacementSolver) SetPrimaryBodyAnchorOffset(offset dprec.Vec3) *AxisDisplacementSolver {
	s.primaryBodyAnchorOffset = offset
	return s
}

// PrimaryBodyAxis returns the body-local-space, unit-length direction,
// relative to the primary target's rotation, along which the
// displacement between the two anchor points is measured.
func (s *AxisDisplacementSolver) PrimaryBodyAxis() dprec.Vec3 {
	return s.primaryBodyAxis
}

// SetPrimaryBodyAxis changes the body-local-space direction, relative to
// the primary target's rotation, along which the displacement between
// the two anchor points is measured. The provided axis need not be
// unit-length; it is normalized before being stored.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisDisplacementSolver) SetPrimaryBodyAxis(axis dprec.Vec3) *AxisDisplacementSolver {
	s.primaryBodyAxis = dprec.UnitVec3(axis)
	return s
}

// SecondaryBodyAnchorOffset returns the body-local-space offset,
// relative to the secondary target's center of mass, of the point at
// which the displacement constraint is anchored on the secondary body.
func (s *AxisDisplacementSolver) SecondaryBodyAnchorOffset() dprec.Vec3 {
	return s.secondaryBodyAnchorOffset
}

// SetSecondaryBodyAnchorOffset changes the body-local-space offset,
// relative to the secondary target's center of mass, of the point at
// which the displacement constraint is anchored on the secondary body.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisDisplacementSolver) SetSecondaryBodyAnchorOffset(offset dprec.Vec3) *AxisDisplacementSolver {
	s.secondaryBodyAnchorOffset = offset
	return s
}

// Displacement returns the signed distance, measured along
// [AxisDisplacementSolver.PrimaryBodyAxis], at which the two anchor
// points are held apart.
func (s *AxisDisplacementSolver) Displacement() float64 {
	return s.displacement
}

// SetDisplacement changes the signed distance, measured along
// [AxisDisplacementSolver.PrimaryBodyAxis], at which the two anchor
// points are held apart. Unlike a plain Euclidean distance, it may be
// negative.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisDisplacementSolver) SetDisplacement(displacement float64) *AxisDisplacementSolver {
	s.displacement = displacement
	return s
}

// Reset implements [PairConstraintSolver.Reset].
//
// It recomputes the constraint's primary and secondary [Jacobian]s and
// current displacement error (drift), the same way
// [AxisDisplacementSolver.recompute] does.
func (s *AxisDisplacementSolver) Reset(ctx PairConstraintContext) {
	s.recompute(ctx)
}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// It resolves a pair of impulses, without restitution, that drive the
// anchor points' relative velocity toward closing the axis-aligned
// displacement error (drift) computed by [AxisDisplacementSolver.Reset],
// pulling the anchor points together along the axis when the actual
// displacement is too large and pushing them apart along the axis when
// it is too small.
func (s *AxisDisplacementSolver) ApplyImpulses(ctx PairConstraintContext) {
	primaryImpulse, secondaryImpulse := ctx.ImpulseSolution(
		s.primaryJacobian, s.secondaryJacobian, s.drift, 0.0,
	)
	ctx.PrimaryTarget.ApplyImpulse(primaryImpulse)
	ctx.SecondaryTarget.ApplyImpulse(secondaryImpulse)
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It first recomputes the constraint's primary and secondary [Jacobian]s
// and current axis-aligned displacement error (drift), the same way
// [AxisDisplacementSolver.recompute] does, since a preceding nudge - by
// this solver's own previous iteration, or by another constraint acting
// on either target - may have moved either target since
// [AxisDisplacementSolver.Reset] or the last call to this method. It
// then nudges both targets' positions and rotations to reduce any
// remaining displacement error between their anchor points.
func (s *AxisDisplacementSolver) ApplyNudges(ctx PairConstraintContext) {
	s.recompute(ctx)

	primaryNudge, secondaryNudge := ctx.NudgeSolution(
		s.primaryJacobian, s.secondaryJacobian, s.drift,
	)
	ctx.PrimaryTarget.ApplyNudge(primaryNudge)
	ctx.SecondaryTarget.ApplyNudge(secondaryNudge)
}

// recompute recalculates the constraint's primary and secondary
// [Jacobian]s, along with the world-space offset from each target's
// center of mass to its respective anchor point (derived from
// PrimaryBodyAnchorOffset and SecondaryBodyAnchorOffset combined with
// each target's current rotation), and the current displacement error
// (drift) between the actual axis-aligned displacement and
// Displacement, based on the targets' current positions and rotations.
//
// The axis-aligned displacement is obtained by projecting the offset
// between the two anchor points onto PrimaryBodyAxis, transformed into
// world space through the primary target's current rotation.
func (s *AxisDisplacementSolver) recompute(ctx PairConstraintContext) {
	primaryAnchorOffsetWS := dprec.QuatVec3Rotation(ctx.PrimaryTarget.Rotation(), s.primaryBodyAnchorOffset)
	primaryAnchorWS := dprec.Vec3Sum(ctx.PrimaryTarget.Position(), primaryAnchorOffsetWS)

	secondaryAnchorOffsetWS := dprec.QuatVec3Rotation(ctx.SecondaryTarget.Rotation(), s.secondaryBodyAnchorOffset)
	secondaryAnchorWS := dprec.Vec3Sum(ctx.SecondaryTarget.Position(), secondaryAnchorOffsetWS)

	axisWS := dprec.QuatVec3Rotation(ctx.PrimaryTarget.Rotation(), s.primaryBodyAxis)

	delta := dprec.Vec3Diff(secondaryAnchorWS, primaryAnchorWS)
	actualDisplacement := dprec.Vec3Dot(axisWS, delta)

	s.primaryJacobian = Jacobian{
		LinearSlope:  dprec.InverseVec3(axisWS),
		AngularSlope: dprec.Vec3Cross(axisWS, primaryAnchorOffsetWS),
	}
	s.secondaryJacobian = Jacobian{
		LinearSlope:  axisWS,
		AngularSlope: dprec.Vec3Cross(secondaryAnchorOffsetWS, axisWS),
	}
	s.drift = s.displacement - actualDisplacement
}
