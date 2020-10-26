package physics

import "github.com/mokiat/gomath/sprec"

type LimitTranslationConstraint struct {
	NilConstraint
	FirstBody  *Body
	SecondBody *Body
	MinY       float32
	MaxY       float32
}

func (c LimitTranslationConstraint) ApplyImpulse(ctx Context) {
	deltaPosition := sprec.Vec3Diff(c.SecondBody.Position, c.FirstBody.Position)
	if deltaPosition.SqrLength() < sqrEpsilon {
		return
	}

	deltaY := sprec.Vec3Dot(c.FirstBody.Orientation.OrientationY(), deltaPosition)
	normalY := sprec.Vec3Prod(c.FirstBody.Orientation.OrientationY(), deltaY)

	deltaVelocity := sprec.Vec3Diff(c.SecondBody.Velocity, sprec.Vec3Sum(c.FirstBody.Velocity, sprec.Vec3Cross(c.FirstBody.AngularVelocity, deltaPosition)))
	contactVelocity := sprec.Vec3Dot(normalY, deltaVelocity)

	if deltaY > c.MaxY && contactVelocity < 0 {
		firstInverseMass := (1.0 / c.FirstBody.Mass) + sprec.Vec3Dot(sprec.Mat3Vec3Prod(sprec.InverseMat3(c.FirstBody.MomentOfInertia), sprec.Vec3Cross(deltaPosition, normalY)), sprec.Vec3Cross(deltaPosition, normalY))
		secondInverseMass := (1.0 / c.SecondBody.Mass)
		totalMass := 1.0 / (firstInverseMass + secondInverseMass)
		impulseStrength := totalMass * contactVelocity
		c.FirstBody.ApplyOffsetImpulse(deltaPosition, sprec.Vec3Prod(normalY, impulseStrength))
		c.SecondBody.ApplyImpulse(sprec.Vec3Prod(normalY, -impulseStrength))
	}

	if deltaY < c.MinY && contactVelocity > 0 {
		firstInverseMass := (1.0 / c.FirstBody.Mass) + sprec.Vec3Dot(sprec.Mat3Vec3Prod(sprec.InverseMat3(c.FirstBody.MomentOfInertia), sprec.Vec3Cross(deltaPosition, normalY)), sprec.Vec3Cross(deltaPosition, normalY))
		secondInverseMass := (1.0 / c.SecondBody.Mass)
		totalMass := 1.0 / (firstInverseMass + secondInverseMass)
		impulseStrength := totalMass * contactVelocity
		c.FirstBody.ApplyOffsetImpulse(deltaPosition, sprec.Vec3Prod(normalY, impulseStrength))
		c.SecondBody.ApplyImpulse(sprec.Vec3Prod(normalY, -impulseStrength))
	}
}
