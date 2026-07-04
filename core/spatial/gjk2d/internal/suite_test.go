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

// fromRectangle builds a [testShape] from a [shape2d.Rectangle]. The corners
// are stored in local space, wound counter-clockwise, while the rectangle's
// center and rotation drive the shape's world placement.
func fromRectangle(rectangle shape2d.Rectangle) testShape {
	hw := rectangle.HalfWidth
	hh := rectangle.HalfHeight
	return testShape{
		Polygon: internal.Polygon{
			Rotation:    rectangle.Rotation,
			InvRotation: rectangle.Rotation.Inverse(),
			Points: []dprec.Vec2{
				dprec.NewVec2(-hw, -hh),
				dprec.NewVec2(hw, -hh),
				dprec.NewVec2(hw, hh),
				dprec.NewVec2(-hw, hh),
			},
		},
		Position:   rectangle.Center,
		SkinRadius: 0.0,
	}
}

// fromTriangle builds a zero-skin [testShape] from a [shape2d.Triangle]. The
// vertices are used verbatim as the polygon core.
func fromTriangle(triangle shape2d.Triangle) testShape {
	return testShape{
		Polygon: internal.Polygon{
			Rotation:    shape2d.IdentityRotation(),
			InvRotation: shape2d.IdentityRotation(),
			Points: []dprec.Vec2{
				triangle.A,
				triangle.B,
				triangle.C,
			},
		},
		Position:   dprec.ZeroVec2(),
		SkinRadius: 0.0,
	}
}

// fromPolygon builds a [testShape] from local-space points positioned and
// oriented by the given position and rotation, with the given skin radius.
func fromPolygon(position dprec.Vec2, rotation shape2d.Rotation, skinRadius float64, points ...dprec.Vec2) testShape {
	return testShape{
		Polygon: internal.Polygon{
			Rotation:    rotation,
			InvRotation: rotation.Inverse(),
			Points:      points,
		},
		Position:   position,
		SkinRadius: skinRadius,
	}
}
