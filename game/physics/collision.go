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

var _ ExplicitSBConstraintSolver = (*soloCollisionSolver)(nil)

type soloCollisionSolver struct {
	NilSBConstraintSolver // TODO: Remove

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
	verticalVelocityAmount := -s.jacobian.EffectiveVelocity(ctx.Body)
	if verticalVelocityAmount > 0 {
		return // moving away
	}
	lambda := verticalVelocityAmount / s.jacobian.InverseEffectiveMass(ctx.Body)
	restitution := ctx.RestitutionCoefficient() * RestitutionClamp(verticalVelocityAmount)
	bounceSolution := s.jacobian.ImpulseSolution((1 + restitution) * lambda)

	// Friction solution
	radiusWS := dprec.QuatVec3Rotation(ctx.Body.Orientation(), s.radius)
	pointVelocity := dprec.Vec3Sum(ctx.Body.Velocity(), dprec.Vec3Cross(ctx.Body.AngularVelocity(), radiusWS))
	verticalVelocity := dprec.Vec3Prod(s.collisionNormal, verticalVelocityAmount)
	lateralVelocity := dprec.Vec3Diff(pointVelocity, verticalVelocity)
	frictionSolution := SBImpulseSolution{}
	if lng := lateralVelocity.Length(); lng > sqrEpsilon {
		lateralDirection := dprec.UnitVec3(lateralVelocity)
		frictionJacobian := Jacobian{
			SlopeVelocity:        lateralDirection,
			SlopeAngularVelocity: dprec.Vec3Cross(radiusWS, lateralDirection),
		}
		frictionLambda := frictionJacobian.ImpulseLambda(ctx.Body)
		// TODO: Have friction coefficient configurable
		const frictionCoefficient = 0.9 // around 0.7 to 0.9 is realistic for dry asphalt
		maxFrictionLambda := lambda * frictionCoefficient
		if frictionLambda < maxFrictionLambda {
			frictionLambda = maxFrictionLambda
		}
		frictionSolution = frictionJacobian.ImpulseSolution(frictionLambda)
	}

	// Note: Make sure to apply these as late as possible, otherwise you are
	// introducing noise that is picked up by subsequent calculations.
	ctx.ApplyImpulseSolution(bounceSolution)
	ctx.ApplyImpulseSolution(frictionSolution)
}

func (s *soloCollisionSolver) ApplyNudges(ctx SBSolverContext) {
	s.updateJacobian(ctx)
	if s.drift > 0 {
		ctx.ApplyNudge(s.jacobian, s.drift)
	}
}

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
	NilDBConstraintSolver

	Normal       dprec.Vec3
	ContactPoint dprec.Vec3
	Depth        float64
}

func (c dualCollisionSolver) CalculateImpulses(ctx DBSolverContext) DBImpulseSolution {
	primary := ctx.Primary
	secondary := ctx.Secondary

	primaryContactRadiusWS := dprec.Vec3Diff(c.ContactPoint, primary.position)
	primaryContactVelocity := dprec.Vec3Sum(primary.velocity, dprec.Vec3Cross(primary.angularVelocity, primaryContactRadiusWS))

	secondaryContactRadiusWS := dprec.Vec3Diff(c.ContactPoint, secondary.position)
	secondaryContactVelocity := dprec.Vec3Sum(secondary.velocity, dprec.Vec3Cross(secondary.angularVelocity, secondaryContactRadiusWS))

	normal := dprec.InverseVec3(c.Normal)
	normalVelocity := dprec.Vec3Dot(c.Normal, primaryContactVelocity) - dprec.Vec3Dot(c.Normal, secondaryContactVelocity)
	if normalVelocity > 0.0 {
		// moving away from ground
		return DBImpulseSolution{}
	}

	restitutionClamp := float64(1.0)
	if dprec.Abs(normalVelocity) < 2.0 {
		restitutionClamp = 0.1
	}
	if dprec.Abs(normalVelocity) < 1.0 {
		restitutionClamp = 0.05
	}
	if dprec.Abs(normalVelocity) < 0.5 {
		restitutionClamp = 0.0
	}

	totalMass := (1 + primary.restitutionCoefficient*secondary.restitutionCoefficient*restitutionClamp) / ((1.0 / primary.mass) + dprec.Vec3Dot(dprec.Mat3Vec3Prod(dprec.InverseMat3(primary.momentOfInertia), dprec.Vec3Cross(primaryContactRadiusWS, normal)), dprec.Vec3Cross(primaryContactRadiusWS, normal)) + (1.0 / secondary.mass) + dprec.Vec3Dot(dprec.Mat3Vec3Prod(dprec.InverseMat3(secondary.momentOfInertia), dprec.Vec3Cross(secondaryContactRadiusWS, normal)), dprec.Vec3Cross(secondaryContactRadiusWS, normal)))
	pureImpulseStrength := totalMass * dprec.Vec3Dot(normal, primaryContactVelocity)
	impulseStrength := pureImpulseStrength + totalMass*c.Depth // FIXME
	// FIXME: Don't apply, rather return as solution
	primary.applyOffsetImpulse(primaryContactRadiusWS, dprec.InverseVec3(dprec.Vec3Prod(normal, impulseStrength)))
	secondary.applyOffsetImpulse(secondaryContactRadiusWS, dprec.Vec3Prod(normal, impulseStrength))

	// frictionCoef := float64(0.9) // around 0.7 to 0.9 is realistic for dry asphalt
	// lateralVelocity := dprec.Vec3Diff(primaryContactVelocity, dprec.Vec3Prod(c.Normal, verticalVelocity))
	// if lateralVelocity.SqrLength() > sqrEpsilon {
	// 	// FIXME: Lateral impulse uses restitution part on top
	// 	lateralImpulseStrength := totalMass * lateralVelocity.Length()
	// 	if lateralImpulseStrength > dprec.Abs(impulseStrength)*frictionCoef {
	// 		lateralImpulseStrength = dprec.Abs(impulseStrength) * frictionCoef
	// 	}
	// 	lateralDir := dprec.UnitVec3(lateralVelocity)
	// 	// FIXME: Don't apply, rather return as solution
	// 	primary.applyOffsetImpulse(primaryContactRadiusWS, dprec.Vec3Prod(lateralDir, -lateralImpulseStrength))
	// }
	return DBImpulseSolution{} // FIXME
}
