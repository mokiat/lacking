package physics

import "github.com/mokiat/gomath/dprec"

// CoiloverSolverConfig holds the parameters with which a [CoiloverSolver]
// is configured, either through [NewCoiloverSolver] or
// [CoiloverSolver.Configure].
type CoiloverSolverConfig struct {

	// PrimaryBodyAnchorOffset is the body-local-space offset, relative to
	// the primary target's center of mass, of the point at which the
	// coilover is anchored on the primary body.
	PrimaryBodyAnchorOffset dprec.Vec3

	// SecondaryBodyAnchorOffset is the body-local-space offset, relative
	// to the secondary target's center of mass, of the point at which
	// the coilover is anchored on the secondary body.
	SecondaryBodyAnchorOffset dprec.Vec3

	// Frequency is the natural frequency, in Hz, of the spring-damper
	// response that drives the two anchor points toward being
	// RelaxedLength apart.
	//
	// A zero (or otherwise degenerate) Frequency disables the spring and
	// damping forces entirely; [CoiloverSolver.ApplyImpulses] then falls
	// back to only cancelling relative velocity along the constraint's
	// axis, without pulling the anchors back toward RelaxedLength.
	Frequency float64

	// Damping is the damping ratio of the spring-damper response, in the
	// [0.0, 1.0] range, where 0.0 leaves it undamped (it oscillates
	// indefinitely around RelaxedLength) and 1.0 is critically damped
	// (it settles as fast as possible without oscillating).
	Damping float64

	// RelaxedLength is the distance between the two anchor points at
	// which the coilover exerts no force.
	RelaxedLength float64
}

// CoiloverSolver is a [PairConstraintSolver] that models a coilover - the
// spring-damper suspension unit found on vehicles - acting between an
// anchor point on each of its two target bodies.
//
// Unlike a rigid constraint such as [FixedDistanceSolver], it does not
// hold its two anchor points at a fixed distance apart. Instead, it
// drives them toward CoiloverSolverConfig.RelaxedLength through a soft,
// tunable spring-damper response (see CoiloverSolverConfig.Frequency and
// CoiloverSolverConfig.Damping) applied entirely through
// [CoiloverSolver.ApplyImpulses]; [CoiloverSolver.ApplyNudges] is a
// no-op, since the softness of that response already accounts for any
// positional drift over successive steps.
//
// A CoiloverSolver must be configured, either through
// [NewCoiloverSolver] or [CoiloverSolver.Configure], before being
// registered with a [Scene] through [PairConstraintView.Create].
type CoiloverSolver struct {
	primaryBodyAnchorOffset   dprec.Vec3
	secondaryBodyAnchorOffset dprec.Vec3
	frequency                 float64
	damping                   float64
	relaxedLength             float64

	appliedLambda     float64
	primaryJacobian   Jacobian
	secondaryJacobian Jacobian
	drift             float64
}

var _ PairConstraintSolver = (*CoiloverSolver)(nil)

// NewCoiloverSolver creates a new [CoiloverSolver] configured according
// to config.
func NewCoiloverSolver(config CoiloverSolverConfig) *CoiloverSolver {
	result := &CoiloverSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config.
//
// Configure must be called before this solver is registered with a
// [Scene] through [PairConstraintView.Create]. Unlike
// [NewCoiloverSolver], it can be called on an already-allocated solver,
// which allows solvers to be cached (e.g. in a slice) and configured on
// demand.
func (s *CoiloverSolver) Configure(config CoiloverSolverConfig) {
	s.primaryBodyAnchorOffset = config.PrimaryBodyAnchorOffset
	s.secondaryBodyAnchorOffset = config.SecondaryBodyAnchorOffset
	s.frequency = config.Frequency
	s.damping = config.Damping
	s.relaxedLength = config.RelaxedLength
}

// PrimaryBodyAnchorOffset returns the body-local-space offset, relative
// to the primary target's center of mass, of the point at which the
// coilover is anchored on the primary body.
func (s *CoiloverSolver) PrimaryBodyAnchorOffset() dprec.Vec3 {
	return s.primaryBodyAnchorOffset
}

// SetPrimaryBodyAnchorOffset changes the body-local-space offset,
// relative to the primary target's center of mass, of the point at
// which the coilover is anchored on the primary body.
//
// It returns the solver itself, so that calls can be chained.
func (s *CoiloverSolver) SetPrimaryBodyAnchorOffset(offset dprec.Vec3) *CoiloverSolver {
	s.primaryBodyAnchorOffset = offset
	return s
}

// SecondaryBodyAnchorOffset returns the body-local-space offset,
// relative to the secondary target's center of mass, of the point at
// which the coilover is anchored on the secondary body.
func (s *CoiloverSolver) SecondaryBodyAnchorOffset() dprec.Vec3 {
	return s.secondaryBodyAnchorOffset
}

// SetSecondaryBodyAnchorOffset changes the body-local-space offset,
// relative to the secondary target's center of mass, of the point at
// which the coilover is anchored on the secondary body.
//
// It returns the solver itself, so that calls can be chained.
func (s *CoiloverSolver) SetSecondaryBodyAnchorOffset(offset dprec.Vec3) *CoiloverSolver {
	s.secondaryBodyAnchorOffset = offset
	return s
}

// Frequency returns the natural frequency, in Hz, of the spring-damper
// response that drives the two anchor points toward being RelaxedLength
// apart.
func (s *CoiloverSolver) Frequency() float64 {
	return s.frequency
}

// SetFrequency changes the natural frequency, in Hz, of the
// spring-damper response that drives the two anchor points toward being
// RelaxedLength apart.
//
// It returns the solver itself, so that calls can be chained.
func (s *CoiloverSolver) SetFrequency(frequency float64) *CoiloverSolver {
	s.frequency = frequency
	return s
}

// Damping returns the damping ratio of the spring-damper response, in
// the [0.0, 1.0] range.
func (s *CoiloverSolver) Damping() float64 {
	return s.damping
}

// SetDamping changes the damping ratio of the spring-damper response, in
// the [0.0, 1.0] range.
//
// It returns the solver itself, so that calls can be chained.
func (s *CoiloverSolver) SetDamping(damping float64) *CoiloverSolver {
	s.damping = damping
	return s
}

// RelaxedLength returns the distance between the two anchor points at
// which the coilover exerts no force.
func (s *CoiloverSolver) RelaxedLength() float64 {
	return s.relaxedLength
}

// SetRelaxedLength changes the distance between the two anchor points at
// which the coilover exerts no force.
//
// It returns the solver itself, so that calls can be chained.
func (s *CoiloverSolver) SetRelaxedLength(length float64) *CoiloverSolver {
	s.relaxedLength = length
	return s
}

// Reset implements [PairConstraintSolver.Reset].
//
// It recomputes the constraint's primary and secondary [Jacobian]s,
// along with the world-space offset from each target's center of mass to
// its respective anchor point (derived from
// CoiloverSolverConfig.PrimaryBodyAnchorOffset and
// CoiloverSolverConfig.SecondaryBodyAnchorOffset combined with each
// target's current rotation), and the current length error (drift)
// between the distance separating the two anchor points and
// CoiloverSolverConfig.RelaxedLength, based on the targets' current
// positions and rotations. It also clears any impulse accumulated by a
// prior [CoiloverSolver.ApplyImpulses] call, so that the spring-damper
// response starts fresh for the new step.
//
// If the two anchor points currently coincide, the constraint's axis is
// undefined; the world Y axis is used as a fallback in that degenerate
// case.
func (s *CoiloverSolver) Reset(ctx PairConstraintContext) {
	primaryAnchorOffsetWS := dprec.QuatVec3Rotation(ctx.PrimaryTarget.Rotation(), s.primaryBodyAnchorOffset)
	primaryAnchorWS := dprec.Vec3Sum(ctx.PrimaryTarget.Position(), primaryAnchorOffsetWS)

	secondaryAnchorOffsetWS := dprec.QuatVec3Rotation(ctx.SecondaryTarget.Rotation(), s.secondaryBodyAnchorOffset)
	secondaryAnchorWS := dprec.Vec3Sum(ctx.SecondaryTarget.Position(), secondaryAnchorOffsetWS)

	delta := dprec.Vec3Diff(secondaryAnchorWS, primaryAnchorWS)
	actualDistance := delta.Length()

	normal := dprec.BasisYVec3()
	if actualDistance > Epsilon {
		normal = dprec.Vec3Quot(delta, actualDistance)
	}

	s.appliedLambda = 0.0 // reset applied impulse
	s.primaryJacobian = Jacobian{
		LinearSlope:  dprec.InverseVec3(normal),
		AngularSlope: dprec.Vec3Cross(normal, primaryAnchorOffsetWS),
	}
	s.secondaryJacobian = Jacobian{
		LinearSlope:  normal,
		AngularSlope: dprec.Vec3Cross(secondaryAnchorOffsetWS, normal),
	}
	s.drift = s.relaxedLength - actualDistance
}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// It resolves a pair of impulses that drive the anchor points' relative
// velocity along the constraint's axis according to a soft
// spring-damper response tuned by CoiloverSolverConfig.Frequency and
// CoiloverSolverConfig.Damping, pulling the two anchors back toward
// being CoiloverSolverConfig.RelaxedLength apart. The impulse is
// accumulated internally across successive calls within the same step
// and fed back into the response, as is standard practice for soft
// constraints.
//
// If Frequency is zero (or otherwise produces a degenerate spring
// constant), this falls back to only cancelling relative velocity along
// the constraint's axis, without pulling the anchors toward
// RelaxedLength.
//
// ApplyImpulses does nothing if both targets have infinite effective
// mass along the constraint's axis (e.g. both are static, or otherwise
// immovable along it).
func (s *CoiloverSolver) ApplyImpulses(ctx PairConstraintContext) {
	invEffectiveMass := s.primaryJacobian.InverseEffectiveMass(ctx.PrimaryTarget) + s.secondaryJacobian.InverseEffectiveMass(ctx.SecondaryTarget)
	if invEffectiveMass < Epsilon {
		return // infinite effective mass
	}

	w := 2.0 * dprec.Pi * s.frequency
	dc := 2.0 * s.damping * w / invEffectiveMass
	k := w * w / invEffectiveMass

	var gamma, beta float64
	if k > Epsilon {
		gamma = 1.0 / (ctx.DeltaSeconds * (dc + ctx.DeltaSeconds*k))
		beta = ctx.DeltaSeconds * k * gamma
	}

	effVelocity := s.primaryJacobian.EffectiveVelocity(ctx.PrimaryTarget) + s.secondaryJacobian.EffectiveVelocity(ctx.SecondaryTarget)
	lambda := -(effVelocity - beta*s.drift + gamma*s.appliedLambda) / (invEffectiveMass + gamma)
	s.appliedLambda += lambda

	primaryImpulse := s.primaryJacobian.Impulse(lambda)
	ctx.PrimaryTarget.ApplyImpulse(primaryImpulse)

	secondaryImpulse := s.secondaryJacobian.Impulse(lambda)
	ctx.SecondaryTarget.ApplyImpulse(secondaryImpulse)
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It does nothing. Unlike a rigid constraint, a coilover's softness (see
// CoiloverSolverConfig.Frequency and CoiloverSolverConfig.Damping)
// already accounts for positional drift through
// [CoiloverSolver.ApplyImpulses], so no separate position-level
// correction is needed.
func (s *CoiloverSolver) ApplyNudges(ctx PairConstraintContext) {}
