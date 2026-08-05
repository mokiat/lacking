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

	pointOffsetWS dprec.Vec3
	jacobian      Jacobian
	drift         float64
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

	s.frictionCoefficient = dprec.Sqrt(config.BodyFrictionCoefficient * config.TerrainFrictionCoefficient)
	s.restitutionCoefficient = max(config.BodyRestitutionCoefficient, config.TerrainRestitutionCoefficient)
}

// Reset implements [SoloConstraintSolver.Reset].
//
// It recomputes the contact's [Jacobian], along with the world-space
// offset from the target's center of mass to the contact point that it
// is derived from, based on the target's current position.
func (s *SoloCollisionSolver) Reset(ctx SoloConstraintContext) {
	s.pointOffsetWS = dprec.Vec3Diff(s.bodyContactPoint, ctx.Target.Position())

	s.jacobian = Jacobian{
		LinearSlope:  s.terrainContactNormal,
		AngularSlope: dprec.Vec3Cross(s.pointOffsetWS, s.terrainContactNormal),
	}

	s.drift = s.contactDepth
}

// ApplyImpulses implements [SoloConstraintSolver.ApplyImpulses].
//
// It first resolves the contact's normal impulse, combining restitution
// with Baumgarte positional-drift stabilization. If the target is
// already moving away from the terrain, it returns without applying
// anything, leaving any remaining penetration to
// [SoloCollisionSolver.ApplyNudges]. Otherwise, it additionally resolves
// a Coulomb friction impulse that opposes the target's lateral
// (tangential) velocity at the contact point, clamped to a fraction of
// the normal impulse's restitution-only component.
func (s *SoloCollisionSolver) ApplyImpulses(ctx SoloConstraintContext) {
	// Bounce solution
	bounceLambda, baumgarteLambda := ctx.ImpulseLambdaComponents(s.jacobian, s.drift, s.restitutionCoefficient)
	if bounceLambda < 0.0 {
		return // moving away
	}
	bounceImpulse := s.jacobian.Impulse(bounceLambda + baumgarteLambda)

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
		maxFrictionLambda := bounceLambda * s.frictionCoefficient
		frictionLambda = min(frictionLambda, maxFrictionLambda)
		frictionImpulse = frictionJacobian.Impulse(frictionLambda)
	}

	// Note: Make sure to apply these as late as possible, otherwise you are
	// introducing noise that is picked up by friction calculations.
	ctx.Target.ApplyImpulse(bounceImpulse)
	ctx.Target.ApplyImpulse(frictionImpulse)
}

// ApplyNudges implements [SoloConstraintSolver.ApplyNudges].
//
// If the contact is still penetrating, it nudges the target along the
// terrain's contact normal to reduce the penetration.
func (s *SoloCollisionSolver) ApplyNudges(ctx SoloConstraintContext) {
	if s.drift > 0.0 {
		nudge := ctx.NudgeSolution(s.jacobian, s.drift)
		ctx.Target.ApplyNudge(nudge)
	}
}
