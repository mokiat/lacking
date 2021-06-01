package physics

import "github.com/mokiat/gomath/sprec"

const (
	epsilon    = float32(0.001)
	sqrEpsilon = epsilon * epsilon
)

type CollisionShape interface{}

var _ SBConstraintSolver = (*groundCollisionSolver)(nil)

type groundCollisionSolver struct {
	NilSBConstraintSolver

	Normal       sprec.Vec3
	ContactPoint sprec.Vec3
	Depth        float32
}

func (c groundCollisionSolver) CalculateImpulses(ctx SBSolverContext) SBImpulseSolution {
	primary := ctx.Body
	contactRadiusWS := sprec.Vec3Diff(c.ContactPoint, primary.position)
	contactVelocity := sprec.Vec3Sum(primary.velocity, sprec.Vec3Cross(primary.angularVelocity, contactRadiusWS))
	verticalVelocity := sprec.Vec3Dot(c.Normal, contactVelocity)

	normal := sprec.InverseVec3(c.Normal)
	normalVelocity := sprec.Vec3Dot(c.Normal, contactVelocity)
	if normalVelocity > 0.0 {
		// moving away from ground
		return SBImpulseSolution{}
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

	totalMass := (1 + primary.restitutionCoefficient*restitutionClamp) / ((1.0 / primary.mass) + sprec.Vec3Dot(sprec.Mat3Vec3Prod(sprec.InverseMat3(primary.momentOfInertia), sprec.Vec3Cross(contactRadiusWS, normal)), sprec.Vec3Cross(contactRadiusWS, normal)))
	pureImpulseStrength := totalMass * sprec.Vec3Dot(normal, contactVelocity)
	impulseStrength := pureImpulseStrength + totalMass*c.Depth // FIXME
	// FIXME: Don't apply, rather return as solution
	primary.applyOffsetImpulse(contactRadiusWS, sprec.InverseVec3(sprec.Vec3Prod(normal, impulseStrength)))

	frictionCoef := float32(0.9)
	lateralVelocity := sprec.Vec3Diff(contactVelocity, sprec.Vec3Prod(c.Normal, verticalVelocity))
	if lateralVelocity.SqrLength() > sqrEpsilon {
		// FIXME: Lateral impulse uses restitution part on top
		lateralImpulseStrength := totalMass * lateralVelocity.Length()
		if lateralImpulseStrength > sprec.Abs(impulseStrength)*frictionCoef {
			lateralImpulseStrength = sprec.Abs(impulseStrength) * frictionCoef
		}
		lateralDir := sprec.UnitVec3(lateralVelocity)
		// FIXME: Don't apply, rather return as solution
		primary.applyOffsetImpulse(contactRadiusWS, sprec.Vec3Prod(lateralDir, -lateralImpulseStrength))
	}
	return SBImpulseSolution{} // FIXME
}
