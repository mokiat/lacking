package physics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape"
)

type AerodynamicShape = shape.Placement[AerodynamicSolver]

// AerodynamicSolver represents a shape that is affected
// by air or liquid motion and inflicts a force on the body.
type AerodynamicSolver interface {
	Force(windSpeed dprec.Vec3) dprec.Vec3
	shape.Shape // FIXME
}

func NewSurfaceAerodynamicShape(width, height, length float64) *SurfaceAerodynamicShape {
	return &SurfaceAerodynamicShape{
		width:  width,
		height: height,
		length: length,
	}
}

var _ AerodynamicSolver = (*SurfaceAerodynamicShape)(nil)

type SurfaceAerodynamicShape struct {
	width  float64
	height float64
	length float64
}

func (s *SurfaceAerodynamicShape) Force(windSpeed dprec.Vec3) dprec.Vec3 {
	// DRAG
	dragX := s.length * s.height * dprec.Vec3Dot(windSpeed, dprec.BasisXVec3())
	dragY := s.width * s.length * dprec.Vec3Dot(windSpeed, dprec.BasisYVec3())
	dragZ := s.width * s.height * dprec.Vec3Dot(windSpeed, dprec.BasisZVec3())
	ro := 1.2
	cD := 0.5
	result := dprec.Vec3Prod(dprec.NewVec3(dragX, dragY, dragZ), windSpeed.Length()*ro*cD/2.0)

	// LIFT
	liftVelocity := -dprec.Vec3Dot(windSpeed, dprec.BasisZVec3())
	if liftVelocity > 0 {
		stallSN := dprec.Sin(dprec.Degrees(40))
		sn := dprec.Vec3Dot(dprec.UnitVec3(windSpeed), dprec.BasisYVec3())
		if dprec.Abs(sn) < stallSN {
			result = dprec.Vec3Sum(result, dprec.NewVec3(
				0.0,
				sn*2.5*liftVelocity*liftVelocity*s.width*s.length,
				0.0,
			))
		}
	}

	return result
}

func (s *SurfaceAerodynamicShape) BoundingSphereRadius() float64 {
	return dprec.Sqrt(s.width*s.width+s.height*s.height+s.length*s.length) / 2.0
}
