package physics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape"
)

const (
	epsilon    = float64(0.0001)
	sqrEpsilon = epsilon * epsilon
)

type CollisionShape shape.Shape

var _ SBConstraintSolver = (*groundCollisionSolver)(nil)

type groundCollisionSolver struct {
	NilSBConstraintSolver

	Normal       dprec.Vec3
	ContactPoint dprec.Vec3
	Depth        float64
}

func (c groundCollisionSolver) CalculateImpulses(ctx SBSolverContext) SBImpulseSolution {
	primary := ctx.Body
	contactRadiusWS := dprec.Vec3Diff(c.ContactPoint, primary.position)
	contactVelocity := dprec.Vec3Sum(primary.velocity, dprec.Vec3Cross(primary.angularVelocity, contactRadiusWS))
	verticalVelocity := dprec.Vec3Dot(c.Normal, contactVelocity)

	normal := dprec.InverseVec3(c.Normal)
	normalVelocity := dprec.Vec3Dot(c.Normal, contactVelocity)
	if normalVelocity > 0.0 {
		// moving away from ground
		return SBImpulseSolution{}
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

	totalMass := (1 + primary.restitutionCoefficient*restitutionClamp) / ((1.0 / primary.mass) + dprec.Vec3Dot(dprec.Mat3Vec3Prod(dprec.InverseMat3(primary.momentOfInertia), dprec.Vec3Cross(contactRadiusWS, normal)), dprec.Vec3Cross(contactRadiusWS, normal)))
	pureImpulseStrength := totalMass * dprec.Vec3Dot(normal, contactVelocity)
	impulseStrength := pureImpulseStrength + totalMass*c.Depth // FIXME
	// FIXME: Don't apply, rather return as solution
	primary.applyOffsetImpulse(contactRadiusWS, dprec.InverseVec3(dprec.Vec3Prod(normal, impulseStrength)))

	frictionCoef := float64(0.9) // around 0.7 to 0.9 is realistic for dry asphalt
	lateralVelocity := dprec.Vec3Diff(contactVelocity, dprec.Vec3Prod(c.Normal, verticalVelocity))
	if lateralVelocity.SqrLength() > sqrEpsilon {
		// FIXME: Lateral impulse uses restitution part on top
		lateralImpulseStrength := totalMass * lateralVelocity.Length()
		if lateralImpulseStrength > dprec.Abs(impulseStrength)*frictionCoef {
			lateralImpulseStrength = dprec.Abs(impulseStrength) * frictionCoef
		}
		lateralDir := dprec.UnitVec3(lateralVelocity)
		// FIXME: Don't apply, rather return as solution
		primary.applyOffsetImpulse(contactRadiusWS, dprec.Vec3Prod(lateralDir, -lateralImpulseStrength))
	}
	return SBImpulseSolution{} // FIXME
}
