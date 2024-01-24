package medium

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

func NewStaticAirMedium() *StaticAirMedium {
	return &StaticAirMedium{
		airDensity:  1.2,
		airVelocity: dprec.ZeroVec3(),
	}
}

var _ solver.Medium = (*StaticAirMedium)(nil)

type StaticAirMedium struct {
	airDensity  float64
	airVelocity dprec.Vec3
}

func (m *StaticAirMedium) AirDensity() float64 {
	return m.airDensity
}

func (m *StaticAirMedium) SetAirDensity(density float64) {
	m.airDensity = density
}

func (m *StaticAirMedium) AirVelocity() dprec.Vec3 {
	return m.airVelocity
}

func (m *StaticAirMedium) SetAirVelocity(velocity dprec.Vec3) {
	m.airVelocity = velocity
}

func (m *StaticAirMedium) Density(position dprec.Vec3) float64 {
	return m.airDensity
}

func (m *StaticAirMedium) Velocity(position dprec.Vec3) dprec.Vec3 {
	return m.airVelocity
}
