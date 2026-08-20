package physics

import "github.com/mokiat/gomath/dprec"

// DistanceSolverConfig holds the parameters with which a [DistanceSolver]
// is configured, either through [NewDistanceSolver] or
// [DistanceSolver.Configure].
type DistanceSolverConfig struct {

	// PrimaryBodyAnchorOffset is the body-local-space offset, relative to
	// the primary target's center of mass, of the point at which the
	// distance constraint is anchored on the primary body.
	PrimaryBodyAnchorOffset dprec.Vec3

	// SecondaryBodyAnchorOffset is the body-local-space offset, relative
	// to the secondary target's center of mass, of the point at which
	// the distance constraint is anchored on the secondary body.
	SecondaryBodyAnchorOffset dprec.Vec3

	// Distance is the distance at which the two anchor points are held
	// apart.
	Distance float64
}

// DistanceSolver is a [PairConstraintSolver] that holds an anchor point on
// each of its two target bodies at a fixed distance apart, acting like a
// rigid rod between the two - it resists the anchor points moving both
// closer together and farther apart.
//
// Unlike [FixedDistanceSolver], which anchors a single target to a fixed
// point in world space, DistanceSolver acts between two independently
// moving targets.
//
// A DistanceSolver must be configured, either through [NewDistanceSolver]
// or [DistanceSolver.Configure], before being registered with a [Scene]
// through [PairConstraintView.Create].
type DistanceSolver struct {
	primaryBodyAnchorOffset   dprec.Vec3
	secondaryBodyAnchorOffset dprec.Vec3
	distance                  float64

	primaryJacobian   Jacobian
	secondaryJacobian Jacobian
	drift             float64
}

var _ PairConstraintSolver = (*DistanceSolver)(nil)

// NewDistanceSolver creates a new [DistanceSolver] configured according to
// config.
func NewDistanceSolver(config DistanceSolverConfig) *DistanceSolver {
	result := &DistanceSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config. Negative
// Distance values are clamped to zero, as with [DistanceSolver.SetDistance].
//
// Configure must be called before this solver is registered with a
// [Scene] through [PairConstraintView.Create]. Unlike [NewDistanceSolver],
// it can be called on an already-allocated solver, which allows solvers to
// be cached (e.g. in a slice) and configured on demand.
func (s *DistanceSolver) Configure(config DistanceSolverConfig) {
	s.primaryBodyAnchorOffset = config.PrimaryBodyAnchorOffset
	s.secondaryBodyAnchorOffset = config.SecondaryBodyAnchorOffset
	s.distance = max(0.0, config.Distance)
}

// PrimaryBodyAnchorOffset returns the body-local-space offset, relative
// to the primary target's center of mass, of the point at which the
// distance constraint is anchored on the primary body.
func (s *DistanceSolver) PrimaryBodyAnchorOffset() dprec.Vec3 {
	return s.primaryBodyAnchorOffset
}

// SetPrimaryBodyAnchorOffset changes the body-local-space offset,
// relative to the primary target's center of mass, of the point at
// which the distance constraint is anchored on the primary body.
//
// It returns the solver itself, so that calls can be chained.
func (s *DistanceSolver) SetPrimaryBodyAnchorOffset(offset dprec.Vec3) *DistanceSolver {
	s.primaryBodyAnchorOffset = offset
	return s
}

// SecondaryBodyAnchorOffset returns the body-local-space offset,
// relative to the secondary target's center of mass, of the point at
// which the distance constraint is anchored on the secondary body.
func (s *DistanceSolver) SecondaryBodyAnchorOffset() dprec.Vec3 {
	return s.secondaryBodyAnchorOffset
}

// SetSecondaryBodyAnchorOffset changes the body-local-space offset,
// relative to the secondary target's center of mass, of the point at
// which the distance constraint is anchored on the secondary body.
//
// It returns the solver itself, so that calls can be chained.
func (s *DistanceSolver) SetSecondaryBodyAnchorOffset(offset dprec.Vec3) *DistanceSolver {
	s.secondaryBodyAnchorOffset = offset
	return s
}

// Distance returns the distance at which the two anchor points are held
// apart.
func (s *DistanceSolver) Distance() float64 {
	return s.distance
}

// SetDistance changes the distance at which the two anchor points are
// held apart. Negative values are clamped to zero.
//
// It returns the solver itself, so that calls can be chained.
func (s *DistanceSolver) SetDistance(distance float64) *DistanceSolver {
	s.distance = max(0.0, distance)
	return s
}

// Reset implements [PairConstraintSolver.Reset].
//
// It recomputes the constraint's primary and secondary [Jacobian]s and
// current distance error (drift), the same way
// [DistanceSolver.recompute] does.
func (s *DistanceSolver) Reset(ctx PairConstraintContext) {
	s.recompute(ctx)
}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// It resolves a pair of impulses, without restitution, that drive the
// anchor points' relative velocity toward closing the distance error
// (drift) computed by [DistanceSolver.Reset], pulling the anchor points
// together when they are too far apart and pushing them apart when they
// are too close.
func (s *DistanceSolver) ApplyImpulses(ctx PairConstraintContext) {
	primaryImpulse, secondaryImpulse := ctx.ImpulseSolution(
		s.primaryJacobian, s.secondaryJacobian, s.drift, 0.0,
	)
	ctx.PrimaryTarget.ApplyImpulse(primaryImpulse)
	ctx.SecondaryTarget.ApplyImpulse(secondaryImpulse)
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It first recomputes the constraint's primary and secondary [Jacobian]s
// and current distance error (drift), the same way
// [DistanceSolver.recompute] does, since a preceding nudge - by this
// solver's own previous iteration, or by another constraint acting on
// either target - may have moved either target since
// [DistanceSolver.Reset] or the last call to this method. It then nudges
// both targets' positions and rotations to reduce any remaining distance
// error between their anchor points.
func (s *DistanceSolver) ApplyNudges(ctx PairConstraintContext) {
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
// each target's current rotation), and the current distance error
// (drift) between the distance separating the two anchor points and
// Distance, based on the targets' current positions and rotations.
//
// If the two anchor points currently coincide, the constraint's axis is
// undefined; the world Y axis is used as a fallback in that degenerate
// case.
func (s *DistanceSolver) recompute(ctx PairConstraintContext) {
	primaryAnchorOffsetWS := dprec.QuatVec3Rotation(ctx.PrimaryTarget.Rotation(), s.primaryBodyAnchorOffset)
	primaryAnchorWS := dprec.Vec3Sum(ctx.PrimaryTarget.Position(), primaryAnchorOffsetWS)

	secondaryAnchorOffsetWS := dprec.QuatVec3Rotation(ctx.SecondaryTarget.Rotation(), s.secondaryBodyAnchorOffset)
	secondaryAnchorWS := dprec.Vec3Sum(ctx.SecondaryTarget.Position(), secondaryAnchorOffsetWS)

	delta := dprec.Vec3Diff(secondaryAnchorWS, primaryAnchorWS)
	actualDistance := delta.Length()

	normal := dprec.BasisYVec3()
	if actualDistance > Epsilon {
		normal = dprec.UnitVec3(delta)
	}

	s.primaryJacobian = Jacobian{
		LinearSlope:  dprec.InverseVec3(normal),
		AngularSlope: dprec.Vec3Cross(normal, primaryAnchorOffsetWS),
	}
	s.secondaryJacobian = Jacobian{
		LinearSlope:  normal,
		AngularSlope: dprec.Vec3Cross(secondaryAnchorOffsetWS, normal),
	}
	s.drift = s.distance - actualDistance
}
