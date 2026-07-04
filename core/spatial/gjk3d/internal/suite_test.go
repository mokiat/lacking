package internal_test

import (
	"testing"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk3d/internal"
	"github.com/mokiat/lacking/core/spatial/shape3d"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInternal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GJK 3D Internal Suite")
}

type testShape struct {
	Hull       internal.Hull
	Position   dprec.Vec3
	SkinRadius float64
}

func fromSphere(sphere shape3d.Sphere) testShape {
	return testShape{
		Hull: internal.Hull{
			Rotation:    shape3d.IdentityRotation(),
			InvRotation: shape3d.IdentityRotation(),
			Points:      []dprec.Vec3{dprec.ZeroVec3()},
		},
		Position:   sphere.Center,
		SkinRadius: sphere.Radius,
	}
}

func fromCapsule(segment shape3d.Segment, radius float64) testShape {
	return testShape{
		Hull: internal.Hull{
			Rotation:    shape3d.IdentityRotation(),
			InvRotation: shape3d.IdentityRotation(),
			Points: []dprec.Vec3{
				segment.A,
				segment.B,
			},
		},
		Position:   dprec.ZeroVec3(),
		SkinRadius: radius,
	}
}

// fromBox builds a zero-skin [testShape] from a [shape3d.Box]. The eight
// corners are stored in local space, while the box center and rotation drive
// the shape's world placement.
func fromBox(box shape3d.Box) testShape {
	hw := box.HalfWidth
	hh := box.HalfHeight
	hl := box.HalfLength
	return testShape{
		Hull: internal.Hull{
			Rotation:    box.Rotation,
			InvRotation: box.Rotation.Inverse(),
			Points: []dprec.Vec3{
				dprec.NewVec3(-hw, -hh, -hl),
				dprec.NewVec3(hw, -hh, -hl),
				dprec.NewVec3(hw, hh, -hl),
				dprec.NewVec3(-hw, hh, -hl),
				dprec.NewVec3(-hw, -hh, hl),
				dprec.NewVec3(hw, -hh, hl),
				dprec.NewVec3(hw, hh, hl),
				dprec.NewVec3(-hw, hh, hl),
			},
		},
		Position:   box.Center,
		SkinRadius: 0.0,
	}
}

// fromTriangle builds a zero-skin [testShape] from a [shape3d.Triangle]. The
// vertices are used verbatim as the hull core.
func fromTriangle(triangle shape3d.Triangle) testShape {
	return testShape{
		Hull: internal.Hull{
			Rotation:    shape3d.IdentityRotation(),
			InvRotation: shape3d.IdentityRotation(),
			Points: []dprec.Vec3{
				triangle.A,
				triangle.B,
				triangle.C,
			},
		},
		Position:   dprec.ZeroVec3(),
		SkinRadius: 0.0,
	}
}

