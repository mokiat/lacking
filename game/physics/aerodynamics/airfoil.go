package aerodynamics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics"
)

func NewAirfoilSolver(width, length float64) *AirfoilSolver {
	return &AirfoilSolver{
		width:  width,
		length: length,

		dragCoefficient: 0.5,
		liftCoefficient: 2.1,
	}
}

var _ physics.AerodynamicSolver = (*AirfoilSolver)(nil)

type AirfoilSolver struct {
	width  float64
	length float64

	dragCoefficient float64
	liftCoefficient float64
}

func (s *AirfoilSolver) SetDragCoefficient(coefficient float64) *AirfoilSolver {
	s.dragCoefficient = coefficient
	return s
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

		coef := localLiftCoefficient(angleOfAttack)
		force := 0.5 * density * area * dprec.Sqr(effWindVelocity) * coef
		result = dprec.Vec3Sum(result, dprec.Vec3Prod(dprec.BasisYVec3(), force))
	}

	// lateral
	lateralAmount := dprec.Abs(dprec.Vec3Dot(planarWindDir, dprec.BasisXVec3()))
	if lateralAmount > 0.01 {
		effWindVelocity := windSpeedLng * lateralAmount
		angleOfAttack := dprec.Atan2(windY, dprec.Abs(windX)) // keep symmetric

		coef := localLiftCoefficient(angleOfAttack)
		force := 0.5 * density * area * dprec.Sqr(effWindVelocity) * coef
		result = dprec.Vec3Sum(result, dprec.Vec3Prod(dprec.BasisYVec3(), force))
	}

	return result
}

func localLiftCoefficient(angle dprec.Angle) float64 {
	if angle < 0.0 {
		return -localLiftCoefficient(-angle) // flipped symmetric
	}
	degrees := angle.Degrees()
	stallDegrees := 15.0
	addition := 1.2 * max(0.0, degrees*(2.0*stallDegrees-degrees)) / dprec.Sqr(stallDegrees)
	return 1.2*dprec.Sin(angle) + addition
}

// func (s *AirfoilSolver) Force(windSpeed dprec.Vec3, density float64) dprec.Vec3 {
// 	windSpeedLng := windSpeed.Length()
// 	if windSpeedLng < 0.1 {
// 		return dprec.ZeroVec3()
// 	}

// 	area := s.width * s.length
// 	windX := dprec.Vec3Dot(windSpeed, dprec.BasisXVec3())
// 	windY := dprec.Vec3Dot(windSpeed, dprec.BasisYVec3())
// 	windZ := dprec.Vec3Dot(windSpeed, dprec.BasisZVec3())
// 	planarWindDir := dprec.UnitVec3(dprec.NewVec3(windX, 0.0, windZ))

// 	var result dprec.Vec3

// 	// direct (chordwise)
// 	directAmount := dprec.Abs(dprec.Vec3Dot(planarWindDir, dprec.BasisZVec3()))
// 	if directAmount > 0.01 {
// 		effWindVelocity := windSpeedLng * directAmount
// 		angleOfAttack := dprec.Atan2(windY, -windZ)

// 		dragCoef := dragCoefficient(angleOfAttack)
// 		dragForce := 0.5 * density * area * dprec.Sqr(effWindVelocity) * dragCoef
// 		dragDir := dprec.UnitVec3(dprec.NewVec3(0.0, windY, windZ))
// 		result = dprec.Vec3Sum(result, dprec.Vec3Prod(dragDir, dragForce))

// 		liftCoef := liftCoefficient(angleOfAttack)
// 		liftForce := 0.5 * density * area * dprec.Sqr(effWindVelocity) * liftCoef
// 		liftDir := dprec.Vec3Cross(dprec.BasisXVec3(), dragDir)
// 		result = dprec.Vec3Sum(result, dprec.Vec3Prod(liftDir, liftForce))
// 	}

// 	// lateral
// 	lateralAmount := dprec.Abs(dprec.Vec3Dot(planarWindDir, dprec.BasisXVec3()))
// 	if lateralAmount > 0.01 {
// 		effWindVelocity := windSpeedLng * lateralAmount
// 		angleOfAttack := dprec.Atan2(windY, dprec.Abs(windX)) // keep symmetric

// 		dragCoef := dragCoefficient(angleOfAttack) * 1.5 // assume flat surface
// 		dragForce := 0.5 * density * area * dprec.Sqr(effWindVelocity) * dragCoef
// 		dragDir := dprec.UnitVec3(dprec.NewVec3(windX, windY, 0.0))
// 		result = dprec.Vec3Sum(result, dprec.Vec3Prod(dragDir, dragForce))

// 		liftCoef := liftCoefficient(angleOfAttack) / 2.0 // assume flat surface
// 		liftForce := 0.5 * density * area * dprec.Sqr(effWindVelocity) * liftCoef
// 		liftDir := dprec.Vec3Cross(dprec.BasisZVec3(), dragDir)
// 		result = dprec.Vec3Sum(result, dprec.Vec3Prod(liftDir, liftForce))
// 	}

// 	return result
// }

func liftCoefficient(angle dprec.Angle) float64 {
	const ( // TODO: Configurable
		// maxLift      = 1.7
		maxLift = 1.4
		// stallDegrees = 30.0
		stallDegrees = 25.0
	)

	if angle < 0.0 {
		return -liftCoefficient(-angle)
	}

	type sample struct {
		x dprec.Angle
		y float64
	}

	keypoints := []sample{
		{
			x: dprec.Degrees(0.0),
			y: 0.0,
		},
		{
			x: dprec.Degrees(stallDegrees),
			y: maxLift,
		},
		{
			x: dprec.Degrees(stallDegrees + 5.0),
			y: maxLift / 4.0,
		},
		{
			x: dprec.Degrees(90.0),
			y: 0.0,
		},
		{
			x: dprec.Degrees(150.0),
			y: -maxLift / 3.0,
		},
		{
			x: dprec.Degrees(180.0),
			y: 0.0,
		},
	}

	for i := range len(keypoints) - 1 {
		current := keypoints[i]
		next := keypoints[i+1]
		if angle >= current.x && angle <= next.x {
			ratio := float64((angle - current.x) / (next.x - current.x))
			return dprec.Mix(current.y, next.y, ratio)
		}
	}
	panic("not found")
}

func dragCoefficient(angle dprec.Angle) float64 {
	const ( // TODO: Configurable
		maxDrag = 2.0
		// maxDrag = 1.5
		minDrag = 0.1
	)

	if angle < 0.0 {
		return dragCoefficient(-angle)
	}

	if angle > dprec.Degrees(90.0) {
		return dragCoefficient(dprec.Degrees(180.0) - angle) // symmetric
	}

	ratio := float64(angle / dprec.Degrees(90.0))
	smoothstep := 3.0*ratio*ratio - 2.0*ratio*ratio*ratio
	// smoothstep := ratio * ratio * ratio * (ratio*(6.0*ratio-15.0) + 10.0)
	return dprec.Mix(minDrag, maxDrag, smoothstep)
}
