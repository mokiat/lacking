package physics

import (
	"github.com/mokiat/gomath/dprec"
)

// TODO: Move under aerodynamics package

// IdentityTransform returns a new Transform that represents the origin.
func IdentityTransform() Transform {
	return Transform{
		position: dprec.ZeroVec3(),
		rotation: dprec.IdentityQuat(),
	}
}

// NewTransform creates a new Transform with the specified position and
// rotation.
func NewTransform(position dprec.Vec3, rotation dprec.Quat) Transform {
	return Transform{
		position: position,
		rotation: rotation,
	}
}

// Transform represents a shape transformation - translation and rotation.
type Transform struct {
	position dprec.Vec3
	rotation dprec.Quat
}

// Position returns the translation of this Transform.
func (t Transform) Position() dprec.Vec3 {
	return t.position
}

// Rotation returns the orientation of this Transform.
func (t Transform) Rotation() dprec.Quat {
	return t.rotation
}

// Transformed returns a new Transform that is based on this one but has the
// specified Transform applied to it.
func (t Transform) Transformed(transform Transform) Transform {
	// Note: Doing an identity check on the current or parent transform,
	// as a form of quick return, actually worsens the performance.
	return Transform{
		position: dprec.Vec3Sum(
			transform.position,
			dprec.QuatVec3Rotation(transform.rotation, t.position),
		),
		rotation: dprec.QuatProd(transform.rotation, t.rotation),
	}
}

func NewAerodynamicShape(transform Transform, solver AerodynamicSolver) AerodynamicShape {
	return AerodynamicShape{
		Transform: transform,
		solver:    solver,
	}
}

type AerodynamicShape struct {
	Transform
	solver AerodynamicSolver
}

// Transformed returns a new Placement that is based on this one but has
// the specified transform applied to it.
func (p AerodynamicShape) Transformed(parent Transform) AerodynamicShape {
	p.Transform = p.Transform.Transformed(parent)
	return p
}

// AerodynamicSolver represents a shape that is affected
// by air or liquid motion and inflicts a force on the body.
type AerodynamicSolver interface {
	Force(windSpeed dprec.Vec3, density float64) dprec.Vec3
}

func NewSurfaceAerodynamicShape(width, height, length float64) *SurfaceAerodynamicShape {
	return &SurfaceAerodynamicShape{
		width:  width,
		height: height,
		length: length,

		dragCoefficient: 0.5,
		liftCoefficient: 2.1,
	}
}

var _ AerodynamicSolver = (*SurfaceAerodynamicShape)(nil)

type SurfaceAerodynamicShape struct {
	width  float64
	height float64
	length float64

	dragCoefficient float64
	liftCoefficient float64
}

func (s *SurfaceAerodynamicShape) SetDragCoefficient(coefficient float64) *SurfaceAerodynamicShape {
	s.dragCoefficient = coefficient
	return s
}

func (s *SurfaceAerodynamicShape) SetLiftCoefficient(coefficient float64) *SurfaceAerodynamicShape {
	s.liftCoefficient = coefficient
	return s
}

func (s *SurfaceAerodynamicShape) Force(windSpeed dprec.Vec3, density float64) dprec.Vec3 {
	// DRAG
	dragX := s.length * s.height * dprec.Vec3Dot(windSpeed, dprec.BasisXVec3())
	dragY := s.width * s.length * dprec.Vec3Dot(windSpeed, dprec.BasisYVec3())
	dragZ := s.width * s.height * dprec.Vec3Dot(windSpeed, dprec.BasisZVec3())
	result := dprec.Vec3Prod(dprec.NewVec3(dragX, dragY, dragZ), windSpeed.Length()*density*s.dragCoefficient/2.0)

	// LIFT
	liftVelocity := -dprec.Vec3Dot(windSpeed, dprec.BasisZVec3())
	if liftVelocity > 0 {
		stallSN := dprec.Sin(dprec.Degrees(40))
		sn := dprec.Vec3Dot(dprec.UnitVec3(windSpeed), dprec.BasisYVec3())
		if dprec.Abs(sn) < stallSN {
			result = dprec.Vec3Sum(result, dprec.NewVec3(
				0.0,
				sn*density*s.liftCoefficient*liftVelocity*liftVelocity*s.width*s.length,
				0.0,
			))
		}
	}

	return result
}

func (s *SurfaceAerodynamicShape) BoundingSphereRadius() float64 {
	return dprec.Sqrt(s.width*s.width+s.height*s.height+s.length*s.length) / 2.0
}
