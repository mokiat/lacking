package physics

import "github.com/mokiat/gomath/dprec"

// SoloCollisionSolverConfig holds the parameters with which a
// [SoloCollisionSolver] is configured, either through
// [NewSoloCollisionSolver] or [SoloCollisionSolver.Configure].
//
// It describes a single contact between a body and static terrain -
// the contact geometry (normals, point, penetration depth) as well as
// the per-surface material properties (friction, restitution) that are
// combined into the solver's effective coefficients.
type SoloCollisionSolverConfig struct {

	// TerrainFrictionCoefficient is the friction coefficient of the
	// terrain surface at the contact point.
	TerrainFrictionCoefficient float64

	// TerrainRestitutionCoefficient is the restitution (bounciness)
	// coefficient of the terrain surface at the contact point.
	TerrainRestitutionCoefficient float64

	// TerrainContactNormal is the unit-length surface normal of the
	// terrain at the contact point, expressed in world space and
	// pointing away from the terrain (i.e. towards the body).
	TerrainContactNormal dprec.Vec3

	// BodyFrictionCoefficient is the friction coefficient of the body's
	// surface at the contact point.
	BodyFrictionCoefficient float64

	// BodyRestitutionCoefficient is the restitution (bounciness)
	// coefficient of the body's surface at the contact point.
	BodyRestitutionCoefficient float64

	// BodyContactPoint is the position, in world space, of the point on
	// the body's surface where the contact occurs.
	BodyContactPoint dprec.Vec3

	// ContactDepth is the penetration depth between the body and the
	// terrain, as measured along TerrainContactNormal at the moment the
	// contact was detected. It is expected to be positive while the two
	// are overlapping.
	ContactDepth float64
}

// SoloCollisionSolver is a [SoloConstraintSolver] that resolves a single
// contact between a body and static terrain.
//
// Through [SoloCollisionSolver.ApplyImpulses] it applies a normal
// ("bounce") impulse - which prevents the body from moving further into
// the terrain and accounts for restitution and positional-drift
// stabilization - together with a Coulomb friction impulse bounded by
// that normal impulse. Through [SoloCollisionSolver.ApplyNudges] it
// separately corrects any remaining penetration at the position level.
//
// A SoloCollisionSolver must be configured, either through
// [NewSoloCollisionSolver] or [SoloCollisionSolver.Configure], before
// being registered with a [Scene] through [SoloConstraintView.Create].
type SoloCollisionSolver struct {
	terrainContactNormal dprec.Vec3
	bodyContactPoint     dprec.Vec3
	contactDepth         float64

	frictionCoefficient    float64
	restitutionCoefficient float64

	pointOffsetLS dprec.Vec3
	pointOffsetWS dprec.Vec3
	jacobian      Jacobian
	drift         float64

	targetVelocity float64
	bounceLambda   float64
	frictionLambda float64
}

var _ SoloConstraintSolver = (*SoloCollisionSolver)(nil)

// NewSoloCollisionSolver creates a new [SoloCollisionSolver] configured
// according to config. See [SoloCollisionSolver.Configure] for details.
func NewSoloCollisionSolver(config SoloCollisionSolverConfig) *SoloCollisionSolver {
	result := &SoloCollisionSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config.
//
// The body's and terrain's friction coefficients are combined into a
// single coefficient through their geometric mean, and their restitution
// coefficients through their maximum.
//
// Configure must be called before this solver is registered with a
// [Scene] through [SoloConstraintView.Create]. Unlike
// [NewSoloCollisionSolver], it can be called on an already-allocated
// solver, which allows solvers to be cached (e.g. in a slice) and
// configured on demand as new contacts are detected.
func (s *SoloCollisionSolver) Configure(config SoloCollisionSolverConfig) {
	s.terrainContactNormal = config.TerrainContactNormal
	s.bodyContactPoint = config.BodyContactPoint
	s.contactDepth = config.ContactDepth

	s.frictionCoefficient = CombinedFrictionCoefficient(
		config.BodyFrictionCoefficient,
		config.TerrainFrictionCoefficient,
	)
	s.restitutionCoefficient = CombinedRestitutionCoefficient(
		config.BodyRestitutionCoefficient,
		config.TerrainRestitutionCoefficient,
	)
}

// Reset implements [SoloConstraintSolver.Reset].
//
// It recomputes the contact's [Jacobian] from the target's current
// position and determines, from the velocity at which the target
// approaches the terrain, the velocity that the contact is to be resolved
// to, combining restitution with Baumgarte positional-drift
// stabilization.
func (s *SoloCollisionSolver) Reset(ctx SoloConstraintContext) {
	// The contact point is attached to the target, so that it can be
	// followed as the target gets moved over the course of the step.
	s.pointOffsetLS = dprec.QuatVec3Rotation(
		dprec.ConjugateQuat(ctx.Target.Rotation()),
		dprec.Vec3Diff(s.bodyContactPoint, ctx.Target.Position()),
	)
	s.recompute(ctx)

	// The velocity that the contact should end up with is determined once,
	// from the velocity that the target approaches the terrain with, so
	// that the individual iterations converge towards it instead of each
	// bouncing off of the outcome of the previous one.
	approachVelocity := min(0.0, s.jacobian.EffectiveVelocity(ctx.Target))
	bounceVelocity := -s.restitutionCoefficient * RestitutionClamp(approachVelocity) * approachVelocity
	driftVelocity := ctx.ImpulseBeta * s.drift / ctx.DeltaSeconds
	s.targetVelocity = max(bounceVelocity, driftVelocity)

	s.bounceLambda = 0.0
	s.frictionLambda = 0.0
}

// ApplyImpulses implements [SoloConstraintSolver.ApplyImpulses].
//
// It first resolves the contact's normal impulse, which drives the
// contact towards the velocity that was determined by
// [SoloCollisionSolver.Reset]. The impulse only ever pushes the target
// away from the terrain when summed up across the iterations of a step,
// though an individual iteration may take back part of what previous
// ones have applied. It additionally resolves a Coulomb friction impulse
// that opposes the target's lateral (tangential) velocity at the contact
// point, which is bounded, across the iterations of a step, by a fraction
// of the normal impulse.
func (s *SoloCollisionSolver) ApplyImpulses(ctx SoloConstraintContext) {
	// Bounce solution
	invEffectiveMass := s.jacobian.InverseEffectiveMass(ctx.Target)
	if invEffectiveMass < Epsilon {
		return
	}
	// The impulse that has been applied so far, across all iterations, is
	// what must never pull the target into the terrain. An individual
	// iteration is free to take back part of it, in case other constraints
	// have pushed the target away further than necessary in the meantime.
	bounceLambda := max(0.0, s.bounceLambda+(s.targetVelocity-s.jacobian.EffectiveVelocity(ctx.Target))/invEffectiveMass)
	deltaBounceLambda := bounceLambda - s.bounceLambda
	s.bounceLambda = bounceLambda
	bounceImpulse := s.jacobian.Impulse(deltaBounceLambda)

	// Friction solution
	pointVelocity := dprec.Vec3Sum(ctx.Target.LinearVelocity(), dprec.Vec3Cross(ctx.Target.AngularVelocity(), s.pointOffsetWS))
	pointLateralVelocity := dprec.Vec3Projection(pointVelocity, s.terrainContactNormal)
	var frictionImpulse Impulse
	if lng := pointLateralVelocity.Length(); lng > Epsilon {
		velocityLateralDirection := dprec.UnitVec3(pointLateralVelocity)
		frictionJacobian := Jacobian{
			LinearSlope:  dprec.InverseVec3(velocityLateralDirection),
			AngularSlope: dprec.Vec3Cross(velocityLateralDirection, s.pointOffsetWS),
		}
		frictionLambda := ctx.ImpulseLambda(frictionJacobian, 0.0, 0.0)
		maxFrictionLambda := max(0.0, s.bounceLambda*s.frictionCoefficient-s.frictionLambda)
		frictionLambda = min(frictionLambda, maxFrictionLambda)
		s.frictionLambda += frictionLambda
		frictionImpulse = frictionJacobian.Impulse(frictionLambda)
	}

	// Note: Make sure to apply these as late as possible, otherwise you are
	// introducing noise that is picked up by friction calculations.
	ctx.Target.ApplyImpulse(bounceImpulse)
	ctx.Target.ApplyImpulse(frictionImpulse)
}

// ApplyNudges implements [SoloConstraintSolver.ApplyNudges].
//
// It first recomputes the contact's [Jacobian] and remaining penetration
// from the target's current position, since the target's motion or a
// preceding nudge - by this solver's own previous iteration, or by
// another constraint acting on the same target - may have moved the
// target since [SoloCollisionSolver.Reset] or the last call to this
// method. If the contact is still penetrating, it then nudges the target
// along the terrain's contact normal to reduce the penetration.
func (s *SoloCollisionSolver) ApplyNudges(ctx SoloConstraintContext) {
	s.recompute(ctx)
	if s.drift > 0.0 {
		nudge := ctx.NudgeSolution(s.jacobian, s.drift)
		ctx.Target.ApplyNudge(nudge)
	}
}

// recompute recalculates the contact's [Jacobian] and remaining
// penetration, along with the world-space offset from the target's
// center of mass to the contact point that they are derived from, based
// on the target's current position and orientation.
func (s *SoloCollisionSolver) recompute(ctx SoloConstraintContext) {
	s.pointOffsetWS = dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.pointOffsetLS)

	s.jacobian = Jacobian{
		LinearSlope:  s.terrainContactNormal,
		AngularSlope: dprec.Vec3Cross(s.pointOffsetWS, s.terrainContactNormal),
	}

	// The penetration that remains is the one that was detected, reduced
	// by however much the contact point has moved away from the terrain
	// since.
	point := dprec.Vec3Sum(ctx.Target.Position(), s.pointOffsetWS)
	separation := dprec.Vec3Diff(point, s.bodyContactPoint)
	s.drift = s.contactDepth - dprec.Vec3Dot(separation, s.terrainContactNormal)
}
