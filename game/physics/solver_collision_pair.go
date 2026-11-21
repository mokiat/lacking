package physics

import "github.com/mokiat/gomath/dprec"

// PairCollisionSolverConfig holds the parameters with which a
// [PairCollisionSolver] is configured, either through
// [NewPairCollisionSolver] or [PairCollisionSolver.Configure].
//
// It describes a single contact between two dynamic bodies - the contact
// geometry (normals, points, penetration depth), each as seen from its
// own body's point of view, as well as the per-surface material
// properties (friction, restitution) that are combined into the
// solver's effective coefficients.
type PairCollisionSolverConfig struct {

	// PrimaryFrictionCoefficient is the friction coefficient of the
	// primary body's surface at the contact point.
	PrimaryFrictionCoefficient float64

	// PrimaryRestitutionCoefficient is the restitution (bounciness)
	// coefficient of the primary body's surface at the contact point.
	PrimaryRestitutionCoefficient float64

	// PrimaryContactNormal is the unit-length surface normal of the
	// primary body at the contact point, expressed in world space and
	// pointing away from the primary body (i.e. towards the secondary
	// body). It is expected to be approximately the negation of
	// SecondaryContactNormal.
	PrimaryContactNormal dprec.Vec3

	// PrimaryContactPoint is the position, in world space, of the point
	// on the primary body's surface where the contact occurs.
	PrimaryContactPoint dprec.Vec3

	// SecondaryFrictionCoefficient is the friction coefficient of the
	// secondary body's surface at the contact point.
	SecondaryFrictionCoefficient float64

	// SecondaryRestitutionCoefficient is the restitution (bounciness)
	// coefficient of the secondary body's surface at the contact point.
	SecondaryRestitutionCoefficient float64

	// SecondaryContactNormal is the unit-length surface normal of the
	// secondary body at the contact point, expressed in world space and
	// pointing away from the secondary body (i.e. towards the primary
	// body). It is expected to be approximately the negation of
	// PrimaryContactNormal.
	SecondaryContactNormal dprec.Vec3

	// SecondaryContactPoint is the position, in world space, of the
	// point on the secondary body's surface where the contact occurs.
	SecondaryContactPoint dprec.Vec3

	// ContactDepth is the penetration depth between the primary and
	// secondary bodies, as measured along the contact normals at the
	// moment the contact was detected. It is expected to be positive
	// while the two are overlapping.
	ContactDepth float64
}

// PairCollisionSolver is a [PairConstraintSolver] that resolves a single
// contact between two dynamic bodies.
//
// Through [PairCollisionSolver.ApplyImpulses] it applies a pair of
// normal ("bounce") impulses - which prevent the bodies from moving
// further into one another and account for restitution and
// positional-drift stabilization - together with a pair of Coulomb
// friction impulses bounded by that normal impulse. Through
// [PairCollisionSolver.ApplyNudges] it separately corrects any
// remaining penetration at the position level.
//
// A PairCollisionSolver must be configured, either through
// [NewPairCollisionSolver] or [PairCollisionSolver.Configure], before
// being registered with a [Scene] through [PairConstraintView.Create].
type PairCollisionSolver struct {
	primaryContactNormal   dprec.Vec3
	primaryContactPoint    dprec.Vec3
	secondaryContactNormal dprec.Vec3
	secondaryContactPoint  dprec.Vec3
	contactDepth           float64

	frictionCoefficient    float64
	restitutionCoefficient float64

	primaryPointOffsetLS   dprec.Vec3
	secondaryPointOffsetLS dprec.Vec3
	primaryPointOffsetWS   dprec.Vec3
	secondaryPointOffsetWS dprec.Vec3
	primaryJacobian        Jacobian
	secondaryJacobian      Jacobian
	drift                  float64

	targetVelocity float64
	bounceLambda   float64
	frictionLambda float64
}

var _ PairConstraintSolver = (*PairCollisionSolver)(nil)

// NewPairCollisionSolver creates a new [PairCollisionSolver] configured
// according to config. See [PairCollisionSolver.Configure] for details.
func NewPairCollisionSolver(config PairCollisionSolverConfig) *PairCollisionSolver {
	result := &PairCollisionSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config.
//
// The two bodies' friction coefficients are combined into a single
// coefficient through their geometric mean, and their restitution
// coefficients through their maximum.
//
// Configure must be called before this solver is registered with a
// [Scene] through [PairConstraintView.Create]. Unlike
// [NewPairCollisionSolver], it can be called on an already-allocated
// solver, which allows solvers to be cached (e.g. in a slice) and
// configured on demand as new contacts are detected.
func (s *PairCollisionSolver) Configure(config PairCollisionSolverConfig) {
	s.primaryContactNormal = config.PrimaryContactNormal
	s.primaryContactPoint = config.PrimaryContactPoint
	s.secondaryContactNormal = config.SecondaryContactNormal
	s.secondaryContactPoint = config.SecondaryContactPoint
	s.contactDepth = config.ContactDepth

	s.frictionCoefficient = CombinedFrictionCoefficient(
		config.PrimaryFrictionCoefficient,
		config.SecondaryFrictionCoefficient,
	)
	s.restitutionCoefficient = CombinedRestitutionCoefficient(
		config.PrimaryRestitutionCoefficient,
		config.SecondaryRestitutionCoefficient,
	)
}

// Reset implements [PairConstraintSolver.Reset].
//
// It recomputes the contact's primary and secondary [Jacobian]s from the
// targets' current positions and determines, from the velocity at which
// the targets approach each other, the velocity that the contact is to be
// resolved to, combining restitution with Baumgarte positional-drift
// stabilization.
func (s *PairCollisionSolver) Reset(ctx PairConstraintContext) {
	// The contact points are attached to their targets, so that they can
	// be followed as the targets get moved over the course of the step.
	s.primaryPointOffsetLS = dprec.QuatVec3Rotation(
		dprec.ConjugateQuat(ctx.PrimaryTarget.Rotation()),
		dprec.Vec3Diff(s.primaryContactPoint, ctx.PrimaryTarget.Position()),
	)
	s.secondaryPointOffsetLS = dprec.QuatVec3Rotation(
		dprec.ConjugateQuat(ctx.SecondaryTarget.Rotation()),
		dprec.Vec3Diff(s.secondaryContactPoint, ctx.SecondaryTarget.Position()),
	)
	s.recompute(ctx)

	// The velocity that the contact should end up with is determined once,
	// from the velocity that the targets approach each other with, so that
	// the individual iterations converge towards it instead of each
	// bouncing off of the outcome of the previous one.
	approachVelocity := min(0.0, s.effectiveVelocity(ctx))
	bounceVelocity := -s.restitutionCoefficient * RestitutionClamp(approachVelocity) * approachVelocity
	driftVelocity := ctx.ImpulseBeta * s.drift / ctx.DeltaSeconds
	s.targetVelocity = max(bounceVelocity, driftVelocity)

	s.bounceLambda = 0.0
	s.frictionLambda = 0.0
}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// It first resolves the contact's normal impulse pair, which drives the
// contact towards the velocity that was determined by
// [PairCollisionSolver.Reset]. The impulse pair only ever pushes the
// targets apart when summed up across the iterations of a step, though an
// individual iteration may take back part of what previous ones have
// applied. It additionally resolves a Coulomb friction impulse pair that
// opposes the targets' relative lateral (tangential) velocity at the
// contact, which is bounded, across the iterations of a step, by a
// fraction of the normal impulse.
func (s *PairCollisionSolver) ApplyImpulses(ctx PairConstraintContext) {
	// Bounce solution
	invEffectiveMass := s.primaryJacobian.InverseEffectiveMass(ctx.PrimaryTarget) + s.secondaryJacobian.InverseEffectiveMass(ctx.SecondaryTarget)
	if invEffectiveMass < Epsilon {
		return
	}
	// The impulse that has been applied so far, across all iterations, is
	// what must never pull the targets together. An individual iteration is
	// free to take back part of it, in case other constraints have pushed
	// the targets apart further than necessary in the meantime.
	bounceLambda := max(0.0, s.bounceLambda+(s.targetVelocity-s.effectiveVelocity(ctx))/invEffectiveMass)
	deltaBounceLambda := bounceLambda - s.bounceLambda
	s.bounceLambda = bounceLambda
	primaryBounceImpulse := s.primaryJacobian.Impulse(deltaBounceLambda)
	secondaryBounceImpulse := s.secondaryJacobian.Impulse(deltaBounceLambda)

	// Friction solution
	primaryPointVelocity := dprec.Vec3Sum(ctx.PrimaryTarget.LinearVelocity(), dprec.Vec3Cross(ctx.PrimaryTarget.AngularVelocity(), s.primaryPointOffsetWS))
	secondaryPointVelocity := dprec.Vec3Sum(ctx.SecondaryTarget.LinearVelocity(), dprec.Vec3Cross(ctx.SecondaryTarget.AngularVelocity(), s.secondaryPointOffsetWS))
	deltaPointVelocity := dprec.Vec3Diff(primaryPointVelocity, secondaryPointVelocity)
	pointsLateralVelocity := dprec.Vec3Projection(deltaPointVelocity, s.secondaryContactNormal)
	var primaryFrictionImpulse, secondaryFrictionImpulse Impulse
	if lng := pointsLateralVelocity.Length(); lng > Epsilon {
		velocityLateralDirection := dprec.UnitVec3(pointsLateralVelocity)
		primaryFrictionJacobian := Jacobian{
			LinearSlope:  dprec.InverseVec3(velocityLateralDirection),
			AngularSlope: dprec.Vec3Cross(velocityLateralDirection, s.primaryPointOffsetWS),
		}
		secondaryFrictionJacobian := Jacobian{
			LinearSlope:  velocityLateralDirection,
			AngularSlope: dprec.Vec3Cross(s.secondaryPointOffsetWS, velocityLateralDirection),
		}
		frictionLambda := ctx.ImpulseLambda(primaryFrictionJacobian, secondaryFrictionJacobian, 0.0, 0.0)
		maxFrictionLambda := max(0.0, s.bounceLambda*s.frictionCoefficient-s.frictionLambda)
		frictionLambda = min(frictionLambda, maxFrictionLambda)
		s.frictionLambda += frictionLambda
		primaryFrictionImpulse = primaryFrictionJacobian.Impulse(frictionLambda)
		secondaryFrictionImpulse = secondaryFrictionJacobian.Impulse(frictionLambda)
	}

	// Note: Make sure to apply these as late as possible, otherwise you are
	// introducing noise that is picked up by friction calculations.
	ctx.PrimaryTarget.ApplyImpulse(primaryBounceImpulse)
	ctx.SecondaryTarget.ApplyImpulse(secondaryBounceImpulse)
	ctx.PrimaryTarget.ApplyImpulse(primaryFrictionImpulse)
	ctx.SecondaryTarget.ApplyImpulse(secondaryFrictionImpulse)
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It first recomputes the contact's primary and secondary [Jacobian]s
// and remaining penetration from the targets' current positions, since
// the targets' motion or a preceding nudge - by this solver's own
// previous iteration, or by another constraint acting on either
// target - may have moved a target since
// [PairCollisionSolver.Reset] or the last call to this method. If the
// contact is still penetrating, it then nudges the two targets apart
// along their contact normals to reduce the penetration.
func (s *PairCollisionSolver) ApplyNudges(ctx PairConstraintContext) {
	s.recompute(ctx)
	if s.drift > 0.0 {
		primaryNudge, secondaryNudge := ctx.NudgeSolution(s.primaryJacobian, s.secondaryJacobian, s.drift)
		ctx.PrimaryTarget.ApplyNudge(primaryNudge)
		ctx.SecondaryTarget.ApplyNudge(secondaryNudge)
	}
}

// recompute recalculates the contact's primary and secondary
// [Jacobian]s and remaining penetration, along with the world-space
// offsets from each target's center of mass to its respective contact
// point that they are derived from, based on the targets' current
// positions and orientations.
//
// Each jacobian's linear slope is built from the other body's contact
// normal (e.g. the primary jacobian uses SecondaryContactNormal, not
// PrimaryContactNormal) rather than from an explicit negation, relying
// on the two normals being approximately antiparallel. This is what
// allows a single lambda, as computed by [PairConstraintContext], to be
// applied to both targets - see [PairConstraintContext.ImpulseLambda]
// for that requirement.
func (s *PairCollisionSolver) recompute(ctx PairConstraintContext) {
	s.primaryPointOffsetWS = dprec.QuatVec3Rotation(ctx.PrimaryTarget.Rotation(), s.primaryPointOffsetLS)
	s.secondaryPointOffsetWS = dprec.QuatVec3Rotation(ctx.SecondaryTarget.Rotation(), s.secondaryPointOffsetLS)

	s.primaryJacobian = Jacobian{
		LinearSlope:  s.secondaryContactNormal,
		AngularSlope: dprec.Vec3Cross(s.primaryPointOffsetWS, s.secondaryContactNormal),
	}
	s.secondaryJacobian = Jacobian{
		LinearSlope:  s.primaryContactNormal,
		AngularSlope: dprec.Vec3Cross(s.secondaryPointOffsetWS, s.primaryContactNormal),
	}

	// The penetration that remains is the one that was detected, reduced
	// by however much the two contact points have moved apart since.
	primaryPoint := dprec.Vec3Sum(ctx.PrimaryTarget.Position(), s.primaryPointOffsetWS)
	secondaryPoint := dprec.Vec3Sum(ctx.SecondaryTarget.Position(), s.secondaryPointOffsetWS)
	separation := dprec.Vec3Diff(
		dprec.Vec3Diff(primaryPoint, secondaryPoint),
		dprec.Vec3Diff(s.primaryContactPoint, s.secondaryContactPoint),
	)
	s.drift = s.contactDepth - dprec.Vec3Dot(separation, s.secondaryContactNormal)
}

// effectiveVelocity returns the velocity at which the two targets are
// moving apart at the contact, which is negative while they approach each
// other.
func (s *PairCollisionSolver) effectiveVelocity(ctx PairConstraintContext) float64 {
	return s.primaryJacobian.EffectiveVelocity(ctx.PrimaryTarget) + s.secondaryJacobian.EffectiveVelocity(ctx.SecondaryTarget)
}
