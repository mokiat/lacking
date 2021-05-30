package solver

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.ConstraintSolver = (*LimitTranslation)(nil)

type LimitTranslation struct {
	physics.NilConstraintSolver
	MinY float32
	MaxY float32
}

func (t *LimitTranslation) CalculateImpulses(primary, secondary *physics.Body, elapsedSeconds float32) physics.ConstraintImpulseSolution {
	deltaPosition := sprec.Vec3Diff(secondary.Position(), primary.Position())
	if deltaPosition.SqrLength() < sqrEpsilon {
		return physics.ConstraintImpulseSolution{}
	}

	deltaY := sprec.Vec3Dot(primary.Orientation().OrientationY(), deltaPosition)
	normalY := sprec.Vec3Prod(primary.Orientation().OrientationY(), deltaY)

	deltaVelocity := sprec.Vec3Diff(secondary.Velocity(), sprec.Vec3Sum(primary.Velocity(), sprec.Vec3Cross(primary.AngularVelocity(), deltaPosition)))
	contactVelocity := sprec.Vec3Dot(normalY, deltaVelocity)

	if deltaY > t.MaxY && contactVelocity < 0 {
		firstInverseMass := (1.0 / primary.Mass()) + sprec.Vec3Dot(sprec.Mat3Vec3Prod(sprec.InverseMat3(primary.MomentOfInertia()), sprec.Vec3Cross(deltaPosition, normalY)), sprec.Vec3Cross(deltaPosition, normalY))
		secondInverseMass := (1.0 / secondary.Mass())
		totalMass := 1.0 / (firstInverseMass + secondInverseMass)
		impulseStrength := totalMass * contactVelocity
		return physics.ConstraintImpulseSolution{
			PrimaryImpulse:        sprec.Vec3Prod(normalY, impulseStrength),
			PrimaryAngularImpulse: sprec.Vec3Cross(deltaPosition, sprec.Vec3Prod(normalY, impulseStrength)),
			SecondaryImpulse:      sprec.Vec3Prod(normalY, -impulseStrength),
		}
	}

	if deltaY < t.MinY && contactVelocity > 0 {
		firstInverseMass := (1.0 / primary.Mass()) + sprec.Vec3Dot(sprec.Mat3Vec3Prod(sprec.InverseMat3(primary.MomentOfInertia()), sprec.Vec3Cross(deltaPosition, normalY)), sprec.Vec3Cross(deltaPosition, normalY))
		secondInverseMass := (1.0 / secondary.Mass())
		totalMass := 1.0 / (firstInverseMass + secondInverseMass)
		impulseStrength := totalMass * contactVelocity
		return physics.ConstraintImpulseSolution{
			PrimaryImpulse:        sprec.Vec3Prod(normalY, impulseStrength),
			PrimaryAngularImpulse: sprec.Vec3Cross(deltaPosition, sprec.Vec3Prod(normalY, impulseStrength)),
			SecondaryImpulse:      sprec.Vec3Prod(normalY, -impulseStrength),
		}
	}

	return physics.ConstraintImpulseSolution{}
}
