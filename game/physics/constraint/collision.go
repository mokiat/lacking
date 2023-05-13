package constraint

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

type CollisionState struct {
	PropFrictionCoefficient    float64
	PropRestitutionCoefficient float64

	BodyNormal                 dprec.Vec3
	BodyPoint                  dprec.Vec3
	BodyFrictionCoefficient    float64
	BodyRestitutionCoefficient float64

	Depth float64
}

var _ solver.Constraint = (*Collision)(nil)

type Collision struct {
	propFrictionCoefficient    float64
	propRestitutionCoefficient float64

	bodyCollisionNormal        dprec.Vec3
	bodyCollisionPoint         dprec.Vec3
	bodyFrictionCoefficient    float64
	bodyRestitutionCoefficient float64

	collisionDepth float64

	radius   dprec.Vec3
	jacobian solver.Jacobian
	drift    float64
}

func (s *Collision) Init(state CollisionState) {
	s.propFrictionCoefficient = state.PropFrictionCoefficient
	s.propRestitutionCoefficient = state.PropRestitutionCoefficient

	s.bodyCollisionNormal = state.BodyNormal
	s.bodyCollisionPoint = state.BodyPoint
	s.bodyFrictionCoefficient = state.BodyFrictionCoefficient
	s.bodyRestitutionCoefficient = state.BodyRestitutionCoefficient

	s.collisionDepth = state.Depth
}

func (s *Collision) Reset(ctx solver.Context) {
	radiusWS := dprec.Vec3Diff(s.bodyCollisionPoint, ctx.Target.Position())
	s.radius = dprec.QuatVec3Rotation(dprec.ConjugateQuat(ctx.Target.Rotation()), radiusWS)
	s.jacobian = solver.Jacobian{
		LinearSlope:  dprec.InverseVec3(s.bodyCollisionNormal),
		AngularSlope: dprec.Vec3Cross(s.bodyCollisionNormal, radiusWS),
	}
	s.drift = s.collisionDepth
}

func (s *Collision) ApplyImpulses(ctx solver.Context) {
	// NOTE: We include the bounce force in the max friction calculation.
	// This might actually be accurate, since you have both the force of
	// the object pushing down, as well as the elastic force pushing further
	// down, trying to bounce the object up.
	restitution := s.propRestitutionCoefficient * s.bodyRestitutionCoefficient

	// Bounce solution
	pressureLambda := ctx.JacobianImpulseLambda(s.jacobian, 0.0, restitution)
	if pressureLambda > 0 {
		return // moving away
	}
	bounceSolution := ctx.JacobianImpulseSolution(s.jacobian, s.collisionDepth, restitution)

	// Friction solution
	radiusWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.radius)
	pointVelocity := dprec.Vec3Sum(ctx.Target.LinearVelocity(), dprec.Vec3Cross(ctx.Target.AngularVelocity(), radiusWS))
	verticalVelocity := dprec.Vec3Prod(s.bodyCollisionNormal, dprec.Vec3Dot(s.bodyCollisionNormal, pointVelocity))
	lateralVelocity := dprec.Vec3Diff(pointVelocity, verticalVelocity)
	frictionSolution := solver.Impulse{}
	if lng := lateralVelocity.Length(); lng > solver.Epsilon {
		lateralDirection := dprec.UnitVec3(lateralVelocity)
		frictionJacobian := solver.Jacobian{
			LinearSlope:  lateralDirection,
			AngularSlope: dprec.Vec3Cross(radiusWS, lateralDirection),
		}
		frictionLambda := ctx.JacobianImpulseLambda(frictionJacobian, 0.0, 0.0)
		// TODO: Have friction coefficient configurable
		// const frictionCoefficient = 0.9 // around 0.7 to 0.9 is realistic for dry asphalt
		const frictionCoefficient = 1.2
		maxFrictionLambda := pressureLambda * frictionCoefficient
		if -frictionLambda > -maxFrictionLambda {
			frictionLambda = maxFrictionLambda
		}
		frictionSolution = frictionJacobian.Impulse(frictionLambda)
	}

	// Note: Make sure to apply these as late as possible, otherwise you are
	// introducing noise that is picked up by subsequent calculations.
	ctx.Target.ApplyImpulse(bounceSolution)
	ctx.Target.ApplyImpulse(frictionSolution)
}

func (s *Collision) ApplyNudges(ctx solver.Context) {
	// TODO: Add nudge solution
}

type PairCollisionState struct {
	PrimaryNormal                 dprec.Vec3
	PrimaryPoint                  dprec.Vec3
	PrimaryFrictionCoefficient    float64
	PrimaryRestitutionCoefficient float64

	SecondaryNormal                 dprec.Vec3
	SecondaryPoint                  dprec.Vec3
	SecondaryFrictionCoefficient    float64
	SecondaryRestitutionCoefficient float64

	Depth float64
}

var _ solver.PairConstraint = (*PairCollision)(nil)

type PairCollision struct {
	primaryCollisionNormal        dprec.Vec3
	primaryCollisionPoint         dprec.Vec3
	primaryFrictionCoefficient    float64
	primaryRestitutionCoefficient float64

	secondaryCollisionNormal        dprec.Vec3
	secondaryCollisionPoint         dprec.Vec3
	secondaryFrictionCoefficient    float64
	secondaryRestitutionCoefficient float64

	collisionDepth float64

	primaryRadius   dprec.Vec3
	secondaryRadius dprec.Vec3
	jacobian        solver.PairJacobian
}

func (s *PairCollision) Init(state PairCollisionState) {
	s.primaryCollisionNormal = state.PrimaryNormal
	s.primaryCollisionPoint = state.PrimaryPoint
	s.primaryFrictionCoefficient = state.PrimaryFrictionCoefficient
	s.primaryRestitutionCoefficient = state.PrimaryRestitutionCoefficient

	s.secondaryCollisionNormal = state.SecondaryNormal
	s.secondaryCollisionPoint = state.SecondaryPoint
	s.secondaryFrictionCoefficient = state.SecondaryFrictionCoefficient
	s.secondaryRestitutionCoefficient = state.SecondaryRestitutionCoefficient

	s.collisionDepth = state.Depth
}

func (s *PairCollision) Reset(ctx solver.PairContext) {
	primaryRadiusWS := dprec.Vec3Diff(s.primaryCollisionPoint, ctx.Target.Position())
	s.primaryRadius = dprec.QuatVec3Rotation(dprec.ConjugateQuat(ctx.Target.Rotation()), primaryRadiusWS)
	secondaryRadiusWS := dprec.Vec3Diff(s.secondaryCollisionPoint, ctx.Source.Position())
	s.secondaryRadius = dprec.QuatVec3Rotation(dprec.ConjugateQuat(ctx.Source.Rotation()), secondaryRadiusWS)
	s.jacobian = solver.PairJacobian{
		Target: solver.Jacobian{
			LinearSlope:  dprec.InverseVec3(s.primaryCollisionNormal),
			AngularSlope: dprec.Vec3Cross(s.primaryCollisionNormal, primaryRadiusWS),
		},
		Source: solver.Jacobian{
			LinearSlope:  dprec.InverseVec3(s.secondaryCollisionNormal),
			AngularSlope: dprec.Vec3Cross(s.secondaryCollisionNormal, secondaryRadiusWS),
		},
	}
}

func (s *PairCollision) ApplyImpulses(ctx solver.PairContext) {
	// NOTE: We include the bounce force in the max friction calculation.
	// This might actually be accurate, since you have both the force of
	// the object pushing down, as well as the elastic force pushing further
	// down, trying to bounce the object up.
	restitution := s.primaryRestitutionCoefficient * s.secondaryRestitutionCoefficient

	// Bounce solution
	pressureLambda := ctx.JacobianImpulseLambda(s.jacobian, 0.0, restitution)
	if pressureLambda > 0 {
		return // moving away
	}
	bounceSolution := ctx.JacobianImpulseSolution(s.jacobian, s.collisionDepth, restitution)

	// Friction solution
	primaryRadiusWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.primaryRadius)
	primaryPointVelocity := dprec.Vec3Sum(ctx.Target.LinearVelocity(), dprec.Vec3Cross(ctx.Target.AngularVelocity(), primaryRadiusWS))
	secondaryRadiusWS := dprec.QuatVec3Rotation(ctx.Source.Rotation(), s.secondaryRadius)
	secondaryPointVelocity := dprec.Vec3Sum(ctx.Source.LinearVelocity(), dprec.Vec3Cross(ctx.Source.AngularVelocity(), secondaryRadiusWS))
	deltaPointVelocity := dprec.Vec3Diff(primaryPointVelocity, secondaryPointVelocity)
	verticalVelocity := dprec.Vec3Prod(s.secondaryCollisionNormal, dprec.Vec3Dot(s.secondaryCollisionNormal, deltaPointVelocity))
	lateralVelocity := dprec.Vec3Diff(deltaPointVelocity, verticalVelocity)
	frictionSolution := solver.PairImpulse{}
	if lng := lateralVelocity.Length(); lng > solver.Epsilon {
		lateralDirection := dprec.UnitVec3(lateralVelocity)
		frictionJacobian := solver.PairJacobian{
			Target: solver.Jacobian{
				LinearSlope:  lateralDirection,
				AngularSlope: dprec.Vec3Cross(primaryRadiusWS, lateralDirection),
			},
			Source: solver.Jacobian{
				LinearSlope:  dprec.InverseVec3(lateralDirection),
				AngularSlope: dprec.Vec3Cross(lateralDirection, secondaryRadiusWS),
			},
		}
		frictionLambda := ctx.JacobianImpulseLambda(frictionJacobian, 0.0, 0.0)
		// TODO: Have friction coefficient configurable
		const frictionCoefficient = 0.9 // around 0.7 to 0.9 is realistic for dry asphalt
		maxFrictionLambda := pressureLambda * frictionCoefficient
		if -frictionLambda > -maxFrictionLambda {
			frictionLambda = maxFrictionLambda
		}
		frictionSolution = frictionJacobian.Impulse(frictionLambda)
	}

	// Note: Make sure to apply these as late as possible, otherwise you are
	// introducing noise that is picked up by subsequent calculations.
	ctx.Target.ApplyImpulse(bounceSolution.Target)
	ctx.Source.ApplyImpulse(bounceSolution.Source)
	ctx.Target.ApplyImpulse(frictionSolution.Target)
	ctx.Source.ApplyImpulse(frictionSolution.Source)
}

func (s *PairCollision) ApplyNudges(ctx solver.PairContext) {
	// TODO: Add nudge solution
}
