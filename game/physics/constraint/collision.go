package constraint

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

type CollisionState struct {
	Normal dprec.Vec3
	Point  dprec.Vec3
	Depth  float64
}

var _ solver.Constraint = (*Collision)(nil)

type Collision struct {
	collisionNormal dprec.Vec3
	collisionPoint  dprec.Vec3
	collisionDepth  float64
	// collisionFrictionCoefficient float64
	// collisionRestitutionCoefficient float64

	radius   dprec.Vec3
	jacobian solver.Jacobian
	drift    float64
}

func (s *Collision) Init(state CollisionState) {
	s.collisionNormal = state.Normal
	s.collisionPoint = state.Point
	s.collisionDepth = state.Depth
}

func (s *Collision) Reset(ctx solver.Context) {
	radiusWS := dprec.Vec3Diff(s.collisionPoint, ctx.Target.Position())
	s.radius = dprec.QuatVec3Rotation(dprec.ConjugateQuat(ctx.Target.Rotation()), radiusWS)
	s.jacobian = solver.Jacobian{
		LinearSlope:  dprec.InverseVec3(s.collisionNormal),
		AngularSlope: dprec.Vec3Cross(s.collisionNormal, radiusWS),
	}
	s.drift = s.collisionDepth
}

func (s *Collision) ApplyImpulses(ctx solver.Context) {
	// Bounce solution
	pressureLambda := ctx.JacobianImpulseLambda(s.jacobian, 0.0, 0.0)
	if pressureLambda > 0 {
		return // moving away
	}
	restitution := 0.5 // FIXME: ctx.Target.RestitutionCoefficient
	bounceSolution := ctx.JacobianImpulseSolution(s.jacobian, s.collisionDepth, restitution)

	// Friction solution
	radiusWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.radius)
	pointVelocity := dprec.Vec3Sum(ctx.Target.LinearVelocity(), dprec.Vec3Cross(ctx.Target.AngularVelocity(), radiusWS))
	verticalVelocity := dprec.Vec3Prod(s.collisionNormal, dprec.Vec3Dot(s.collisionNormal, pointVelocity))
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
	TargetNormal dprec.Vec3
	TargetPoint  dprec.Vec3
	SourceNormal dprec.Vec3
	SourcePoint  dprec.Vec3
	Depth        float64
}

var _ solver.PairConstraint = (*PairCollision)(nil)

type PairCollision struct {
	primaryCollisionNormal   dprec.Vec3
	primaryCollisionPoint    dprec.Vec3
	secondaryCollisionNormal dprec.Vec3
	secondaryCollisionPoint  dprec.Vec3
	collisionDepth           float64

	primaryRadius   dprec.Vec3
	secondaryRadius dprec.Vec3
	jacobian        solver.PairJacobian
}

func (s *PairCollision) Init(state PairCollisionState) {
	s.primaryCollisionNormal = state.TargetNormal
	s.primaryCollisionPoint = state.TargetPoint
	s.secondaryCollisionNormal = state.SourceNormal
	s.secondaryCollisionPoint = state.SourcePoint
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
	// Bounce solution
	pressureLambda := ctx.JacobianImpulseLambda(s.jacobian, 0.0, 0.0)
	if pressureLambda > 0 {
		return // moving away
	}
	restitution := 0.5 // FIXME
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
