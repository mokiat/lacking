package physics

import "github.com/mokiat/gomath/dprec"

type SoloCollisionSolverConfig struct {
	TerrainFrictionCoefficient    float64
	TerrainRestitutionCoefficient float64
	TerrainContactNormal          dprec.Vec3

	BodyFrictionCoefficient    float64
	BodyRestitutionCoefficient float64
	BodyContactPoint           dprec.Vec3

	ContactDepth float64
}

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

func (s *SoloCollisionSolver) Init(config SoloCollisionSolverConfig) {
	s.terrainContactNormal = config.TerrainContactNormal
	s.bodyContactPoint = config.BodyContactPoint
	s.contactDepth = config.ContactDepth

	s.frictionCoefficient = dprec.Sqrt(config.BodyFrictionCoefficient * config.TerrainFrictionCoefficient)
	s.restitutionCoefficient = max(config.BodyRestitutionCoefficient, config.TerrainRestitutionCoefficient)
}

func (s *SoloCollisionSolver) Reset(ctx SoloConstraintContext) {
	s.pointOffsetWS = dprec.Vec3Diff(s.bodyContactPoint, ctx.Target.Position())
	s.jacobian = Jacobian{
		LinearSlope:  s.terrainContactNormal,
		AngularSlope: dprec.Vec3Cross(s.pointOffsetWS, s.terrainContactNormal),
	}
	s.drift = s.contactDepth
}

func (s *SoloCollisionSolver) ApplyImpulses(ctx SoloConstraintContext) {
	// Bounce solution
	bounceLambda, baumgarteLambda := ctx.ImpulseLambdaSplit(s.jacobian, s.drift, s.restitutionCoefficient)
	if bounceLambda < 0.0 {
		return // moving away
	}
	bounceImpulse := s.jacobian.Impulse(bounceLambda + baumgarteLambda)

	// Friction solution
	pointVelocity := dprec.Vec3Sum(ctx.Target.LinearVelocity(), dprec.Vec3Cross(ctx.Target.AngularVelocity(), s.pointOffsetWS))
	pointLateralVelocity := dprec.Vec3Projection(pointVelocity, s.terrainContactNormal)
	var frictionSolution Impulse
	if lng := pointLateralVelocity.Length(); lng > Epsilon {
		pointLateralDirection := dprec.UnitVec3(pointLateralVelocity)
		frictionJacobian := Jacobian{
			LinearSlope:  dprec.InverseVec3(pointLateralDirection),
			AngularSlope: dprec.Vec3Cross(pointLateralDirection, s.pointOffsetWS),
		}
		frictionLambda := ctx.ImpulseLambda(frictionJacobian, 0.0, 0.0)
		maxFrictionLambda := bounceLambda * s.frictionCoefficient
		frictionLambda = min(frictionLambda, maxFrictionLambda)
		frictionSolution = frictionJacobian.Impulse(frictionLambda)
	}

	// Note: Make sure to apply these as late as possible, otherwise you are
	// introducing noise that is picked up by friction calculations.
	ctx.Target.ApplyImpulse(bounceImpulse)
	ctx.Target.ApplyImpulse(frictionSolution)
}

func (s *SoloCollisionSolver) ApplyNudges(ctx SoloConstraintContext) {
	if s.drift > 0.0 {
		nudge := ctx.NudgeSolution(s.jacobian, s.drift)
		ctx.Target.ApplyNudge(nudge)
	}
}
