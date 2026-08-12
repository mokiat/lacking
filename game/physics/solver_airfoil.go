package physics

import (
	"github.com/mokiat/gomath/dprec"
)

type AirfoilSolverConfig struct {
	RelativePosition dprec.Vec3
	RelativeRotation dprec.Quat
	SurfaceArea      float64
	StallAngle       dprec.Angle
	LiftCoefficient  float64
}

type AirfoilSolver struct {
	relativePosition dprec.Vec3
	relativeRotation dprec.Quat
	surfaceArea      float64
	stallAngle       dprec.Angle
	liftCoefficient  float64
}

var _ AccelerationSolver = (*AirfoilSolver)(nil)

func NewAirfoilSolver(config AirfoilSolverConfig) *AirfoilSolver {
	result := &AirfoilSolver{}
	result.Configure(config)
	return result
}

func (s *AirfoilSolver) Configure(config AirfoilSolverConfig) {
	s.relativePosition = config.RelativePosition
	s.relativeRotation = config.RelativeRotation
	s.surfaceArea = max(0.0, config.SurfaceArea)
	s.stallAngle = max(0.0, config.StallAngle)
	s.liftCoefficient = max(0.0, config.LiftCoefficient)
}

func (s *AirfoilSolver) RelativePosition() dprec.Vec3 {
	return s.relativePosition
}

func (s *AirfoilSolver) SetRelativePosition(position dprec.Vec3) *AirfoilSolver {
	s.relativePosition = position
	return s
}

func (s *AirfoilSolver) RelativeRotation() dprec.Quat {
	return s.relativeRotation
}

func (s *AirfoilSolver) SetRelativeRotation(rotation dprec.Quat) *AirfoilSolver {
	s.relativeRotation = rotation
	return s
}

func (s *AirfoilSolver) SurfaceArea() float64 {
	return s.surfaceArea
}

func (s *AirfoilSolver) SetSurfaceArea(area float64) *AirfoilSolver {
	s.surfaceArea = max(0.0, area)
	return s
}

func (s *AirfoilSolver) StallAngle() dprec.Angle {
	return s.stallAngle
}

func (s *AirfoilSolver) SetStallAngle(angle dprec.Angle) *AirfoilSolver {
	s.stallAngle = max(0.0, angle)
	return s
}

func (s *AirfoilSolver) LiftCoefficient() float64 {
	return s.liftCoefficient
}

func (s *AirfoilSolver) SetLiftCoefficient(coefficient float64) *AirfoilSolver {
	s.liftCoefficient = max(0.0, coefficient)
	return s
}

func (s *AirfoilSolver) ApplyAcceleration(ctx AccelerationContext) {
	bodyRotation := ctx.Target.Rotation()

	airfoilVelocity := dprec.Vec3Sum(
		ctx.Target.LinearVelocity(),
		dprec.Vec3Cross(
			ctx.Target.AngularVelocity(),
			s.relativePosition,
		),
	)

	relWindVelocity := dprec.Vec3Diff(ctx.MediumVelocity, airfoilVelocity)
	relWindVelocityLng := relWindVelocity.Length()
	if relWindVelocityLng < Epsilon {
		return // no significant wind
	}

	airfoilOffsetWS := dprec.QuatVec3Rotation(bodyRotation, s.relativePosition)
	airfoilRotationWS := dprec.QuatProd(bodyRotation, s.relativeRotation)

	basisX := airfoilRotationWS.OrientationX()
	basisY := airfoilRotationWS.OrientationY()
	basisZ := airfoilRotationWS.OrientationZ()

	windX := dprec.Vec3Dot(relWindVelocity, basisX)
	windY := dprec.Vec3Dot(relWindVelocity, basisY)
	windZ := dprec.Vec3Dot(relWindVelocity, basisZ)

	var (
		directAmount  float64
		lateralAmount float64
	)
	planarWindVelocity := dprec.NewVec3(windX, 0.0, windZ)
	if planarLng := planarWindVelocity.Length(); planarLng > Epsilon {
		directAmount = dprec.Abs(planarWindVelocity.Z / planarLng)
		lateralAmount = dprec.Abs(planarWindVelocity.X / planarLng)
	} else {
		directAmount = 1.0
		lateralAmount = 0.0
	}

	// Wind along the chord line (direct).
	if directAmount > Epsilon {
		effWindVelocity := relWindVelocityLng * directAmount
		angleOfAttack := dprec.Atan2(windY, -windZ)

		coef := s.localLiftCoefficient(angleOfAttack, true)
		magnitude := coef * 0.5 * ctx.MediumDensity * dprec.Sqr(effWindVelocity) * s.surfaceArea

		ctx.Target.ApplyOffsetForce(airfoilOffsetWS, dprec.Vec3Prod(basisY, magnitude))
	}

	// Wind along the span (lateral).
	if lateralAmount > Epsilon {
		effWindVelocity := relWindVelocityLng * lateralAmount
		angleOfAttack := dprec.Atan2(windY, dprec.Abs(windX)) // keep symmetric

		coef := s.localLiftCoefficient(angleOfAttack, false)
		magnitude := coef * 0.5 * ctx.MediumDensity * dprec.Sqr(effWindVelocity) * s.surfaceArea

		ctx.Target.ApplyOffsetForce(airfoilOffsetWS, dprec.Vec3Prod(basisY, magnitude))
	}
}

func (s *AirfoilSolver) localLiftCoefficient(angle dprec.Angle, isDirect bool) float64 {
	if angle < 0.0 {
		return -s.localLiftCoefficient(-angle, isDirect) // flipped symmetric
	}
	degrees := angle.Degrees()
	result := (s.liftCoefficient / 2.0) * dprec.Sin(angle)
	if isDirect && (s.stallAngle > 0.0) { // add the lift coefficient bump prior to stall
		stallDegrees := s.stallAngle.Degrees()
		result += (s.liftCoefficient / 2.0) * max(0.0, degrees*(2.0*stallDegrees-degrees)) / dprec.Sqr(stallDegrees)
	}
	return result
}
