package aerodynamics_test

import (
	"testing"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/aerodynamics"
)

func TestAirfoil(t *testing.T) {

	solver := aerodynamics.NewAirfoilSolver(10.0, 2.0)

	for angle := -180.0; angle <= 180.0; angle += 5.0 {
		cs := dprec.Cos(dprec.Degrees(angle))
		sn := dprec.Sin(dprec.Degrees(angle))
		windSpeed := dprec.NewVec3(0.0, sn, -cs)
		result := solver.Force(windSpeed, 1.0)

		rotation := dprec.RotationQuat(dprec.Degrees(-angle), dprec.BasisXVec3())

		t.Logf("[%.2f] Force: %#v (rotated: %#v)\n", angle, result, dprec.QuatVec3Rotation(rotation, result))
	}
}
