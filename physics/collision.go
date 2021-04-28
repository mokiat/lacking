package physics

import "github.com/mokiat/gomath/sprec"

type GroundCollisionConstraint struct {
	NilConstraint
	Body         *Body
	Normal       sprec.Vec3
	ContactPoint sprec.Vec3
	Depth        float32
}

func (c GroundCollisionConstraint) ApplyImpulse(ctx Context) {
	contactRadiusWS := sprec.Vec3Diff(c.ContactPoint, c.Body.Position)
	contactVelocity := sprec.Vec3Sum(c.Body.Velocity, sprec.Vec3Cross(c.Body.AngularVelocity, contactRadiusWS))
	verticalVelocity := sprec.Vec3Dot(c.Normal, contactVelocity)

	normal := sprec.InverseVec3(c.Normal)
	normalVelocity := sprec.Vec3Dot(c.Normal, contactVelocity)
	if normalVelocity > 0.0 {
		return // moving away from ground
	}

	restitutionClamp := float32(1.0)
	if sprec.Abs(normalVelocity) < 2.0 {
		restitutionClamp = 0.1
	}
	if sprec.Abs(normalVelocity) < 1.0 {
		restitutionClamp = 0.05
	}
	if sprec.Abs(normalVelocity) < 0.5 {
		restitutionClamp = 0.0
	}

	totalMass := (1 + c.Body.RestitutionCoef*restitutionClamp) / ((1.0 / c.Body.Mass) + sprec.Vec3Dot(sprec.Mat3Vec3Prod(sprec.InverseMat3(c.Body.MomentOfInertia), sprec.Vec3Cross(contactRadiusWS, normal)), sprec.Vec3Cross(contactRadiusWS, normal)))
	pureImpulseStrength := totalMass * sprec.Vec3Dot(normal, contactVelocity)
	impulseStrength := pureImpulseStrength + totalMass*c.Depth // FIXME
	c.Body.ApplyOffsetImpulse(contactRadiusWS, sprec.InverseVec3(sprec.Vec3Prod(normal, impulseStrength)))

	frictionCoef := float32(0.9)
	lateralVelocity := sprec.Vec3Diff(contactVelocity, sprec.Vec3Prod(c.Normal, verticalVelocity))
	if lateralVelocity.SqrLength() > sqrEpsilon {
		// FIXME: Lateral impulse uses restitution part on top
		lateralImpulseStrength := totalMass * lateralVelocity.Length()
		if lateralImpulseStrength > sprec.Abs(impulseStrength)*frictionCoef {
			lateralImpulseStrength = sprec.Abs(impulseStrength) * frictionCoef
		}
		lateralDir := sprec.UnitVec3(lateralVelocity)
		c.Body.ApplyOffsetImpulse(contactRadiusWS, sprec.Vec3Prod(lateralDir, -lateralImpulseStrength))
	}
}
