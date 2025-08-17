package aerodynamics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

func NewAirfoilSolver(width, length float64) *AirfoilSolver {
	return &AirfoilSolver{
		width:  width,
		length: length,

		stallAngle:      dprec.Degrees(20.0),
		liftCoefficient: 2.4,
	}
}

var _ physics.AerodynamicSolver = (*AirfoilSolver)(nil)

type AirfoilSolver struct {
	width  float64
	length float64

	stallAngle      dprec.Angle
	liftCoefficient float64
}

func (s *AirfoilSolver) StallAngle() dprec.Angle {
	return s.stallAngle
}

func (s *AirfoilSolver) SetStallAngle(angle dprec.Angle) *AirfoilSolver {
	s.stallAngle = angle
	return s
}

func (s *AirfoilSolver) LiftCoefficient() float64 {
	return s.liftCoefficient
}

func (s *AirfoilSolver) SetLiftCoefficient(coefficient float64) *AirfoilSolver {
	s.liftCoefficient = coefficient
	return s
}

func (s *AirfoilSolver) Force(windSpeed dprec.Vec3, density float64) dprec.Vec3 {
	windSpeedLng := windSpeed.Length()
	if windSpeedLng < 0.01 {
		return dprec.ZeroVec3()
	}

	area := s.width * s.length
	windX := dprec.Vec3Dot(windSpeed, dprec.BasisXVec3())
	windY := dprec.Vec3Dot(windSpeed, dprec.BasisYVec3())
	windZ := dprec.Vec3Dot(windSpeed, dprec.BasisZVec3())
	planarWindDir := dprec.UnitVec3(dprec.NewVec3(windX, 0.0, windZ))

	var result dprec.Vec3

	// direct (chordwise)
	directAmount := dprec.Abs(dprec.Vec3Dot(planarWindDir, dprec.BasisZVec3()))
	if directAmount > 0.01 {
		effWindVelocity := windSpeedLng * directAmount
		angleOfAttack := dprec.Atan2(windY, -windZ)

		coef := s.localLiftCoefficient(angleOfAttack)
		force := 0.5 * density * area * dprec.Sqr(effWindVelocity) * coef
		result = dprec.Vec3Sum(result, dprec.Vec3Prod(dprec.BasisYVec3(), force))
	}

	// lateral
	lateralAmount := dprec.Abs(dprec.Vec3Dot(planarWindDir, dprec.BasisXVec3()))
	if lateralAmount > 0.01 {
		effWindVelocity := windSpeedLng * lateralAmount
		angleOfAttack := dprec.Atan2(windY, dprec.Abs(windX)) // keep symmetric

		coef := s.localLiftCoefficient(angleOfAttack)
		force := 0.5 * density * area * dprec.Sqr(effWindVelocity) * coef
		result = dprec.Vec3Sum(result, dprec.Vec3Prod(dprec.BasisYVec3(), force))
	}

	return result
}

func (s *AirfoilSolver) localLiftCoefficient(angle dprec.Angle) float64 {
	if angle < 0.0 {
		return -s.localLiftCoefficient(-angle) // flipped symmetric
	}
	degrees := angle.Degrees()
	stallDegrees := s.stallAngle.Degrees()
	base := (s.liftCoefficient / 2.0) * dprec.Sin(angle)
	addition := (s.liftCoefficient / 2.0) * max(0.0, degrees*(2.0*stallDegrees-degrees)) / dprec.Sqr(stallDegrees)
	return base + addition
}
