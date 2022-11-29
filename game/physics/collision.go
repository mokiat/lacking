package physics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape"
)

const (
	epsilon    = float64(0.0001)
	sqrEpsilon = epsilon * epsilon
)

var (
	nextCollisionGroup = 1
)

func NewCollisionGroup() int {
	result := nextCollisionGroup
	nextCollisionGroup++
	return result
}

type CollisionShape = shape.Placement[shape.Shape]

var _ SBConstraintSolver = (*soloCollisionSolver)(nil)

type soloCollisionSolver struct {
	collisionNormal dprec.Vec3
	collisionPoint  dprec.Vec3
	collisionDepth  float64

	radius          dprec.Vec3
	initialDistance float64
	jacobian        Jacobian
	drift           float64
}

func (s *soloCollisionSolver) Reset(ctx SBSolverContext) {
	radiusWS := dprec.Vec3Diff(s.collisionPoint, ctx.Body.Position())
	s.radius = dprec.QuatVec3Rotation(dprec.ConjugateQuat(ctx.Body.Orientation()), radiusWS)
	s.initialDistance = dprec.Vec3Dot(s.collisionPoint, s.collisionNormal)
	s.updateJacobian(ctx)
}

func (s *soloCollisionSolver) ApplyImpulses(ctx SBSolverContext) {
	// Bounce solution
	pressureLambda := ctx.JacobianImpulseLambda(s.jacobian, 0.0, 0.0)
	if pressureLambda > 0 {
		return // moving away
	}
	bounceSolution := ctx.JacobianImpulseSolution(s.jacobian, s.collisionDepth, ctx.RestitutionCoefficient())

	// Friction solution
	radiusWS := dprec.QuatVec3Rotation(ctx.Body.Orientation(), s.radius)
	pointVelocity := dprec.Vec3Sum(ctx.Body.Velocity(), dprec.Vec3Cross(ctx.Body.AngularVelocity(), radiusWS))
	verticalVelocity := dprec.Vec3Prod(s.collisionNormal, dprec.Vec3Dot(s.collisionNormal, pointVelocity))
	lateralVelocity := dprec.Vec3Diff(pointVelocity, verticalVelocity)
	frictionSolution := SBImpulseSolution{}
	if lng := lateralVelocity.Length(); lng > sqrEpsilon {
		lateralDirection := dprec.UnitVec3(lateralVelocity)
		frictionJacobian := Jacobian{
			SlopeVelocity:        lateralDirection,
			SlopeAngularVelocity: dprec.Vec3Cross(radiusWS, lateralDirection),
		}
		frictionLambda := ctx.JacobianImpulseLambda(frictionJacobian, 0.0, 0.0)
		// TODO: Have friction coefficient configurable
		const frictionCoefficient = 0.9 // around 0.7 to 0.9 is realistic for dry asphalt
		maxFrictionLambda := pressureLambda * frictionCoefficient
		if -frictionLambda > -maxFrictionLambda {
			frictionLambda = maxFrictionLambda
		}
		frictionSolution = frictionJacobian.ImpulseSolution(frictionLambda)
	}

	// Note: Make sure to apply these as late as possible, otherwise you are
	// introducing noise that is picked up by subsequent calculations.
	ctx.ApplyImpulseSolution(bounceSolution)
	ctx.ApplyImpulseSolution(frictionSolution)
}

func (s *soloCollisionSolver) ApplyNudges(ctx SBSolverContext) {}

func (s *soloCollisionSolver) updateJacobian(ctx SBSolverContext) {
	radiusWS := dprec.QuatVec3Rotation(ctx.Body.Orientation(), s.radius)
	s.jacobian = Jacobian{
		SlopeVelocity:        dprec.InverseVec3(s.collisionNormal),
		SlopeAngularVelocity: dprec.Vec3Cross(s.collisionNormal, radiusWS),
	}
	collisionPoint := dprec.Vec3Sum(ctx.Body.Position(), radiusWS)
	distance := dprec.Vec3Dot(collisionPoint, s.collisionNormal)
	s.drift = s.collisionDepth - (distance - s.initialDistance)
}

var _ DBConstraintSolver = (*dualCollisionSolver)(nil)

type dualCollisionSolver struct {
	primaryCollisionNormal   dprec.Vec3
	primaryCollisionPoint    dprec.Vec3
	secondaryCollisionNormal dprec.Vec3
	secondaryCollisionPoint  dprec.Vec3
	collisionDepth           float64

	primaryRadius   dprec.Vec3
	secondaryRadius dprec.Vec3
	jacobian        PairJacobian
}

func (s *dualCollisionSolver) Reset(ctx DBSolverContext) {
	primaryRadiusWS := dprec.Vec3Diff(s.primaryCollisionPoint, ctx.Primary.Position())
	s.primaryRadius = dprec.QuatVec3Rotation(dprec.ConjugateQuat(ctx.Primary.Orientation()), primaryRadiusWS)
	secondaryRadiusWS := dprec.Vec3Diff(s.secondaryCollisionPoint, ctx.Secondary.Position())
	s.secondaryRadius = dprec.QuatVec3Rotation(dprec.ConjugateQuat(ctx.Secondary.Orientation()), secondaryRadiusWS)
	s.updateJacobian(ctx)
}

func (s *dualCollisionSolver) ApplyImpulses(ctx DBSolverContext) {
	// Bounce solution
	pressureLambda := ctx.JacobianImpulseLambda(s.jacobian, 0.0, 0.0)
	if pressureLambda > 0 {
		return // moving away
	}
	bounceSolution := ctx.JacobianImpulseSolution(s.jacobian, s.collisionDepth, ctx.RestitutionCoefficient())

	// Friction solution
	primaryRadiusWS := dprec.QuatVec3Rotation(ctx.Primary.Orientation(), s.primaryRadius)
	primaryPointVelocity := dprec.Vec3Sum(ctx.Primary.Velocity(), dprec.Vec3Cross(ctx.Primary.AngularVelocity(), primaryRadiusWS))
	secondaryRadiusWS := dprec.QuatVec3Rotation(ctx.Secondary.Orientation(), s.secondaryRadius)
	secondaryPointVelocity := dprec.Vec3Sum(ctx.Secondary.Velocity(), dprec.Vec3Cross(ctx.Secondary.AngularVelocity(), secondaryRadiusWS))
	deltaPointVelocity := dprec.Vec3Diff(primaryPointVelocity, secondaryPointVelocity)
	verticalVelocity := dprec.Vec3Prod(s.secondaryCollisionNormal, dprec.Vec3Dot(s.secondaryCollisionNormal, deltaPointVelocity))
	lateralVelocity := dprec.Vec3Diff(deltaPointVelocity, verticalVelocity)
	frictionSolution := DBImpulseSolution{}
	if lng := lateralVelocity.Length(); lng > sqrEpsilon {
		lateralDirection := dprec.UnitVec3(lateralVelocity)
		frictionJacobian := PairJacobian{
			Primary: Jacobian{
				SlopeVelocity:        lateralDirection,
				SlopeAngularVelocity: dprec.Vec3Cross(primaryRadiusWS, lateralDirection),
			},
			Secondary: Jacobian{
				SlopeVelocity:        dprec.InverseVec3(lateralDirection),
				SlopeAngularVelocity: dprec.Vec3Cross(lateralDirection, secondaryRadiusWS),
			},
		}
		frictionLambda := ctx.JacobianImpulseLambda(frictionJacobian, 0.0, 0.0)
		// TODO: Have friction coefficient configurable
		const frictionCoefficient = 0.9 // around 0.7 to 0.9 is realistic for dry asphalt
		maxFrictionLambda := pressureLambda * frictionCoefficient
		if -frictionLambda > -maxFrictionLambda {
			frictionLambda = maxFrictionLambda
		}
		frictionSolution = frictionJacobian.ImpulseSolution(frictionLambda)
	}

	// Note: Make sure to apply these as late as possible, otherwise you are
	// introducing noise that is picked up by subsequent calculations.
	ctx.ApplyImpulseSolution(bounceSolution)
	ctx.ApplyImpulseSolution(frictionSolution)
}

func (s *dualCollisionSolver) ApplyNudges(ctx DBSolverContext) {}

func (s *dualCollisionSolver) updateJacobian(ctx DBSolverContext) {
	primaryRadiusWS := dprec.QuatVec3Rotation(ctx.Primary.Orientation(), s.primaryRadius)
	secondaryRadiusWS := dprec.QuatVec3Rotation(ctx.Secondary.Orientation(), s.secondaryRadius)
	s.jacobian = PairJacobian{
		Primary: Jacobian{
			SlopeVelocity:        dprec.InverseVec3(s.primaryCollisionNormal),
			SlopeAngularVelocity: dprec.Vec3Cross(s.primaryCollisionNormal, primaryRadiusWS),
		},
		Secondary: Jacobian{
			SlopeVelocity:        dprec.InverseVec3(s.secondaryCollisionNormal),
			SlopeAngularVelocity: dprec.Vec3Cross(s.secondaryCollisionNormal, secondaryRadiusWS),
		},
	}
}
