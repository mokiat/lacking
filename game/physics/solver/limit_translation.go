package solver

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

var _ physics.DBConstraintSolver = (*LimitTranslation)(nil)

type LimitTranslation struct {
	physics.NilDBConstraintSolver

	MinY float64
	MaxY float64
}

func (t *LimitTranslation) CalculateImpulses(ctx physics.DBSolverContext) physics.DBImpulseSolution {
	primary := ctx.Primary
	secondary := ctx.Secondary
	deltaPosition := dprec.Vec3Diff(secondary.Position(), primary.Position())
	if deltaPosition.SqrLength() < sqrEpsilon {
		return physics.DBImpulseSolution{}
	}

	deltaY := dprec.Vec3Dot(primary.Orientation().OrientationY(), deltaPosition)
	normalY := dprec.Vec3Prod(primary.Orientation().OrientationY(), deltaY)

	deltaVelocity := dprec.Vec3Diff(secondary.Velocity(), dprec.Vec3Sum(primary.Velocity(), dprec.Vec3Cross(primary.AngularVelocity(), deltaPosition)))
	contactVelocity := dprec.Vec3Dot(normalY, deltaVelocity)

	if deltaY > t.MaxY && contactVelocity < 0 {
		firstInverseMass := (1.0 / primary.Mass()) + dprec.Vec3Dot(dprec.Mat3Vec3Prod(dprec.InverseMat3(primary.MomentOfInertia()), dprec.Vec3Cross(deltaPosition, normalY)), dprec.Vec3Cross(deltaPosition, normalY))
		secondInverseMass := (1.0 / secondary.Mass())
		totalMass := 1.0 / (firstInverseMass + secondInverseMass)
		impulseStrength := totalMass * contactVelocity
		return physics.DBImpulseSolution{
			Primary: physics.SBImpulseSolution{
				Impulse:        dprec.Vec3Prod(normalY, impulseStrength),
				AngularImpulse: dprec.Vec3Cross(deltaPosition, dprec.Vec3Prod(normalY, impulseStrength)),
			},
			Secondary: physics.SBImpulseSolution{
				Impulse: dprec.Vec3Prod(normalY, -impulseStrength),
			},
		}
	}

	if deltaY < t.MinY && contactVelocity > 0 {
		firstInverseMass := (1.0 / primary.Mass()) + dprec.Vec3Dot(dprec.Mat3Vec3Prod(dprec.InverseMat3(primary.MomentOfInertia()), dprec.Vec3Cross(deltaPosition, normalY)), dprec.Vec3Cross(deltaPosition, normalY))
		secondInverseMass := (1.0 / secondary.Mass())
		totalMass := 1.0 / (firstInverseMass + secondInverseMass)
		impulseStrength := totalMass * contactVelocity
		return physics.DBImpulseSolution{
			Primary: physics.SBImpulseSolution{
				Impulse:        dprec.Vec3Prod(normalY, impulseStrength),
				AngularImpulse: dprec.Vec3Cross(deltaPosition, dprec.Vec3Prod(normalY, impulseStrength)),
			},
			Secondary: physics.SBImpulseSolution{
				Impulse: dprec.Vec3Prod(normalY, -impulseStrength),
			},
		}
	}

	return physics.DBImpulseSolution{}
}
