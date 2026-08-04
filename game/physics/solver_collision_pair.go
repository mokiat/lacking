package physics

import "github.com/mokiat/gomath/dprec"

// PairCollisionSolverConfig holds the parameters with which a
// [PairCollisionSolver] is configured through [PairCollisionSolver.Init].
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
// A PairCollisionSolver must be configured through
// [PairCollisionSolver.Init] before being registered with a [Scene]
// through [PairConstraintView.Create].
type PairCollisionSolver struct {
	primaryContactNormal   dprec.Vec3
	primaryContactPoint    dprec.Vec3
	secondaryContactNormal dprec.Vec3
	secondaryContactPoint  dprec.Vec3
	contactDepth           float64

	frictionCoefficient    float64
	restitutionCoefficient float64

	primaryPointOffsetWS   dprec.Vec3
	secondaryPointOffsetWS dprec.Vec3
	primaryJacobian        Jacobian
	secondaryJacobian      Jacobian
	drift                  float64
}

var _ PairConstraintSolver = (*PairCollisionSolver)(nil)

// Init configures this solver according to config.
//
// The two bodies' friction coefficients are combined into a single
// coefficient through their geometric mean, and their restitution
// coefficients through their maximum.
//
// Init must be called once, before this solver is registered with a
// [Scene] through [PairConstraintView.Create].
func (s *PairCollisionSolver) Init(config PairCollisionSolverConfig) {
	s.primaryContactNormal = config.PrimaryContactNormal
	s.primaryContactPoint = config.PrimaryContactPoint
	s.secondaryContactNormal = config.SecondaryContactNormal
	s.secondaryContactPoint = config.SecondaryContactPoint
	s.contactDepth = config.ContactDepth

	s.frictionCoefficient = dprec.Sqrt(config.PrimaryFrictionCoefficient * config.SecondaryFrictionCoefficient)
	s.restitutionCoefficient = max(config.PrimaryRestitutionCoefficient, config.SecondaryRestitutionCoefficient)
}

// Reset implements [PairConstraintSolver.Reset].
//
// It recomputes the contact's primary and secondary [Jacobian]s, along
// with the world-space offsets from each target's center of mass to its
// respective contact point that they are derived from, based on the
// targets' current positions.
//
// Each jacobian's linear slope is built from the other body's contact
// normal (e.g. the primary jacobian uses SecondaryContactNormal, not
// PrimaryContactNormal) rather than from an explicit negation, relying
// on the two normals being approximately antiparallel. This is what
// allows a single lambda, as computed by [PairConstraintContext], to be
// applied to both targets - see
// [PairConstraintContext.ImpulseLambda] for that requirement.
func (s *PairCollisionSolver) Reset(ctx PairConstraintContext) {
	s.primaryPointOffsetWS = dprec.Vec3Diff(s.primaryContactPoint, ctx.PrimaryTarget.Position())
	s.secondaryPointOffsetWS = dprec.Vec3Diff(s.secondaryContactPoint, ctx.SecondaryTarget.Position())

	s.primaryJacobian = Jacobian{
		LinearSlope:  s.secondaryContactNormal,
		AngularSlope: dprec.Vec3Cross(s.primaryPointOffsetWS, s.secondaryContactNormal),
	}
	s.secondaryJacobian = Jacobian{
		LinearSlope:  s.primaryContactNormal,
		AngularSlope: dprec.Vec3Cross(s.secondaryPointOffsetWS, s.primaryContactNormal),
	}

	s.drift = s.contactDepth
}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// It first resolves the contact's normal impulse pair, combining
// restitution with Baumgarte positional-drift stabilization. If the two
// targets are already moving apart, it returns without applying
// anything, leaving any remaining penetration to
// [PairCollisionSolver.ApplyNudges]. Otherwise, it additionally resolves
// a Coulomb friction impulse pair that opposes the targets' relative
// lateral (tangential) velocity at the contact, clamped to a fraction of
// the normal impulse's restitution-only component.
func (s *PairCollisionSolver) ApplyImpulses(ctx PairConstraintContext) {
	// Bounce solution
	bounceLambda, baumgarteLambda := ctx.ImpulseLambdaComponents(s.primaryJacobian, s.secondaryJacobian, s.drift, s.restitutionCoefficient)
	if bounceLambda < 0.0 {
		return // moving away
	}
	primaryBounceImpulse := s.primaryJacobian.Impulse(bounceLambda + baumgarteLambda)
	secondaryBounceImpulse := s.secondaryJacobian.Impulse(bounceLambda + baumgarteLambda)

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
		maxFrictionLambda := bounceLambda * s.frictionCoefficient
		frictionLambda = min(frictionLambda, maxFrictionLambda)
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
// If the contact is still penetrating, it nudges the two targets apart
// along their contact normals to reduce the penetration.
func (s *PairCollisionSolver) ApplyNudges(ctx PairConstraintContext) {
	if s.drift > 0.0 {
		primaryNudge, secondaryNudge := ctx.NudgeSolution(s.primaryJacobian, s.secondaryJacobian, s.drift)
		ctx.PrimaryTarget.ApplyNudge(primaryNudge)
		ctx.SecondaryTarget.ApplyNudge(secondaryNudge)
	}
}
