package internal_test

import (
	"testing"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk2d/internal"
	"github.com/mokiat/lacking/core/spatial/shape2d"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInternal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GJK 2D Internal Suite")
}

type testShape struct {
	Polygon    internal.Polygon
	Position   dprec.Vec2
	SkinRadius float64
}

func fromCircle(circle shape2d.Circle) testShape {
	return testShape{
		Polygon: internal.Polygon{
			Rotation:    shape2d.IdentityRotation(),
			InvRotation: shape2d.IdentityRotation(),
			Points:      []dprec.Vec2{dprec.ZeroVec2()},
		},
		Position:   circle.Center,
		SkinRadius: circle.Radius,
	}
}

func fromCapsule(capsule shape2d.Capsule) testShape {
	return testShape{
		Polygon: internal.Polygon{
			Rotation:    shape2d.IdentityRotation(),
			InvRotation: shape2d.IdentityRotation(),
			Points: []dprec.Vec2{
				capsule.A,
				capsule.B,
			},
		},
		Position:   dprec.ZeroVec2(),
		SkinRadius: capsule.Radius,
	}
}
