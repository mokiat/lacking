package physics

import "github.com/mokiat/gomath/dprec"

type PairCollisionSolverConfig struct {
	PrimaryFrictionCoefficient    float64
	PrimaryRestitutionCoefficient float64
	PrimaryContactNormal          dprec.Vec3
	PrimaryContactPoint           dprec.Vec3

	SecondaryFrictionCoefficient    float64
	SecondaryRestitutionCoefficient float64
	SecondaryContactNormal          dprec.Vec3
	SecondaryContactPoint           dprec.Vec3

	ContactDepth float64
}

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

func (s *PairCollisionSolver) Init(config PairCollisionSolverConfig) {
	s.primaryContactNormal = config.PrimaryContactNormal
	s.primaryContactPoint = config.PrimaryContactPoint
	s.secondaryContactNormal = config.SecondaryContactNormal
	s.secondaryContactPoint = config.SecondaryContactPoint
	s.contactDepth = config.ContactDepth

	s.frictionCoefficient = dprec.Sqrt(config.PrimaryFrictionCoefficient * config.SecondaryFrictionCoefficient)
	s.restitutionCoefficient = max(config.PrimaryRestitutionCoefficient, config.SecondaryRestitutionCoefficient)
}

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

func (s *PairCollisionSolver) ApplyNudges(ctx PairConstraintContext) {
	if s.drift > 0.0 {
		primaryNudge, secondaryNudge := ctx.NudgeSolution(s.primaryJacobian, s.secondaryJacobian, s.drift)
		ctx.PrimaryTarget.ApplyNudge(primaryNudge)
		ctx.SecondaryTarget.ApplyNudge(secondaryNudge)
	}
}
