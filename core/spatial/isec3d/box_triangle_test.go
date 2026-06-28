package isec3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("BoxTriangle", func() {
	// An axis-aligned unit cube centered at the origin, spanning [-1,1] on each
	// axis.
	var box shape3d.Box

	// A helper to build a triangle lying in a plane parallel to the XY plane at
	// the given height, large enough to cover the box's XY footprint and with its
	// normal pointing along +Z.
	horizontalTriangle := func(z float64) shape3d.Triangle {
		return shape3d.Triangle{
			A: dprec.NewVec3(-2.0, -2.0, z),
			B: dprec.NewVec3(2.0, -2.0, z),
			C: dprec.NewVec3(0.0, 2.0, z),
		}
	}

	BeforeEach(func() {
		box = shape3d.Box{
			Center:     dprec.NewVec3(0.0, 0.0, 0.0),
			Rotation:   shape3d.IdentityRotation(),
			HalfWidth:  1.0,
			HalfHeight: 1.0,
			HalfLength: 1.0,
		}
	})

	Describe("CheckBoxTriangle", func() {
		It("returns true for a triangle slicing through the box", func() {
			Expect(isec3d.CheckBoxTriangle(box, horizontalTriangle(0.0))).To(BeTrue())
		})

		It("returns true for a triangle that just touches a face", func() {
			// The triangle lies in the box's top face plane z=1.
			Expect(isec3d.CheckBoxTriangle(box, horizontalTriangle(1.0))).To(BeTrue())
		})

		It("returns false for a triangle above the box", func() {
			Expect(isec3d.CheckBoxTriangle(box, horizontalTriangle(2.0))).To(BeFalse())
		})

		It("returns false for a triangle beside the box", func() {
			beside := shape3d.Triangle{
				A: dprec.NewVec3(2.0, -2.0, -2.0),
				B: dprec.NewVec3(2.0, 2.0, -2.0),
				C: dprec.NewVec3(2.0, 0.0, 2.0),
			}
			Expect(isec3d.CheckBoxTriangle(box, beside)).To(BeFalse())
		})

		It("returns false for a triangle whose plane clears a corner", func() {
			// The plane x+y+z=3.2 passes just beyond the (1,1,1) corner, whose
			// projection onto the plane normal is sqrt(3) ~= 1.73 < 3.2/sqrt(3).
			cleared := shape3d.Triangle{
				A: dprec.NewVec3(3.2, 0.0, 0.0),
				B: dprec.NewVec3(0.0, 3.2, 0.0),
				C: dprec.NewVec3(0.0, 0.0, 3.2),
			}
			Expect(isec3d.CheckBoxTriangle(box, cleared)).To(BeFalse())
		})

		It("returns true for a triangle whose plane cuts a corner", func() {
			// The plane x+y+z=1.5 passes through the box near the (1,1,1) corner.
			cut := shape3d.Triangle{
				A: dprec.NewVec3(1.5, 0.0, 0.0),
				B: dprec.NewVec3(0.0, 1.5, 0.0),
				C: dprec.NewVec3(0.0, 0.0, 1.5),
			}
			Expect(isec3d.CheckBoxTriangle(box, cut)).To(BeTrue())
		})
	})

	Describe("ResolveBoxTriangle", func() {
		resolve := func(b shape3d.Box, t shape3d.Triangle) (shape3d.Contact, bool) {
			var sink shape3d.LastContact
			isec3d.ResolveBoxTriangle(b, t, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact along the shallowest separating axis", func() {
			// The triangle slices the box at z=0.5, so the box escapes downward
			// (-Z) by 0.5, the smallest of the per-axis overlaps.
			contact, ok := resolve(box, horizontalTriangle(0.5))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 0.0, -1.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.5))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("reports a zero-depth contact for a triangle touching a face", func() {
			contact, ok := resolve(box, horizontalTriangle(1.0))
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("does not yield a contact for a disjoint triangle", func() {
			_, ok := resolve(box, horizontalTriangle(2.0))
			Expect(ok).To(BeFalse())
		})

		It("reports a unit normal", func() {
			contact, ok := resolve(box, horizontalTriangle(0.5))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("places the contact point on the triangle", func() {
			contact, ok := resolve(box, horizontalTriangle(0.5))
			Expect(ok).To(BeTrue())
			// The contact point lies in the triangle's plane (z=0.5).
			Expect(contact.TargetPoint.Z).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("separates the box when it is moved by Depth along the normal", func() {
			contact, ok := resolve(box, horizontalTriangle(0.5))
			Expect(ok).To(BeTrue())

			moved := box
			moved.Center = dprec.Vec3Sum(box.Center, dprec.Vec3Prod(contact.TargetNormal, contact.Depth))
			// After moving out by Depth the box only just touches, so a re-resolve
			// reports essentially zero remaining penetration.
			resolved, ok := resolve(moved, horizontalTriangle(0.5))
			if ok {
				Expect(resolved.Depth).To(BeNumerically("~", 0.0, 1e-6))
			}
		})

		It("resolves against an oriented box in world space", func() {
			// A box long along its local X, rotated 90 degrees about Z, so in world
			// space it spans +/-2 along Y and +/-1 along X and Z.
			rotated := shape3d.Box{
				Center:     dprec.NewVec3(0.0, 0.0, 0.0),
				Rotation:   shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3())),
				HalfWidth:  2.0,
				HalfHeight: 1.0,
				HalfLength: 1.0,
			}
			// A triangle slicing the box at world y=0.5; the box escapes along -Y
			// by 2 - 0.5 = 1.5.
			triangle := shape3d.Triangle{
				A: dprec.NewVec3(-2.0, 0.5, -2.0),
				B: dprec.NewVec3(2.0, 0.5, -2.0),
				C: dprec.NewVec3(0.0, 0.5, 2.0),
			}
			contact, ok := resolve(rotated, triangle)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, -1.0, 0.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.0, 0.5, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 1.5, 1e-6))
		})
	})
})
