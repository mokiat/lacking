package isec3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("SphereBox", func() {
	// An axis-aligned box at the origin with distinct half-extents along each
	// axis, so that any axis confusion is caught. It spans x:[-2,2], y:[-1,1]
	// and z:[-3,3].
	var box shape3d.Box

	newSphere := func(x, y, z, radius float64) shape3d.Sphere {
		return shape3d.Sphere{
			Center: dprec.NewVec3(x, y, z),
			Radius: radius,
		}
	}

	BeforeEach(func() {
		box = shape3d.Box{
			Center:     dprec.NewVec3(0.0, 0.0, 0.0),
			Rotation:   shape3d.IdentityRotation(),
			HalfWidth:  2.0,
			HalfHeight: 1.0,
			HalfLength: 3.0,
		}
	})

	Describe("CheckSphereBox", func() {
		It("returns true for a sphere overlapping a face", func() {
			Expect(isec3d.CheckSphereBox(newSphere(3.0, 0.0, 0.0, 1.5), box)).To(BeTrue())
		})

		It("returns true for a sphere that just touches a face", func() {
			// Right face is at x=2; center at x=3.5 with radius 1.5 just reaches it.
			Expect(isec3d.CheckSphereBox(newSphere(3.5, 0.0, 0.0, 1.5), box)).To(BeTrue())
		})

		It("returns false for a sphere just short of a face", func() {
			Expect(isec3d.CheckSphereBox(newSphere(3.6, 0.0, 0.0, 1.5), box)).To(BeFalse())
		})

		It("returns true for a sphere whose center is inside the box", func() {
			Expect(isec3d.CheckSphereBox(newSphere(0.0, 0.0, 0.0, 0.5), box)).To(BeTrue())
		})

		It("returns true for a sphere that fully contains the box", func() {
			Expect(isec3d.CheckSphereBox(newSphere(0.0, 0.0, 0.0, 10.0), box)).To(BeTrue())
		})

		It("returns true for a sphere reaching an edge", func() {
			// Past the right and top faces; nearest feature is the x=2,y=1 edge.
			Expect(isec3d.CheckSphereBox(newSphere(3.0, 2.0, 0.0, 1.5), box)).To(BeTrue())
		})

		It("returns false when past two faces but short of the edge", func() {
			// The corner-of-two-faces distance is sqrt(2) ~= 1.414, beyond radius.
			Expect(isec3d.CheckSphereBox(newSphere(3.0, 2.0, 0.0, 1.4), box)).To(BeFalse())
		})

		It("returns true for a sphere reaching a corner", func() {
			// Past all three faces; nearest feature is the (2,1,3) corner at
			// distance sqrt(0.75) ~= 0.866.
			Expect(isec3d.CheckSphereBox(newSphere(2.5, 1.5, 3.5, 1.0), box)).To(BeTrue())
		})

		It("returns false when past three faces but short of the corner", func() {
			// The corner distance sqrt(0.75) ~= 0.866 is beyond radius 0.8, even
			// though the center is within radius of each individual face.
			Expect(isec3d.CheckSphereBox(newSphere(2.5, 1.5, 3.5, 0.8), box)).To(BeFalse())
		})

		It("respects the box orientation", func() {
			// Rotate the box 90 degrees about Z, so its local X (half-width 2)
			// now points along world Y, and its local Y (half-height 1) along
			// world X. The box then only spans x:[-1,1].
			rotated := box
			rotated.Rotation = shape3d.RotationFromQuat(
				dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3()),
			)
			// A point that would be inside the unrotated box is now outside.
			Expect(isec3d.CheckSphereBox(newSphere(1.6, 0.0, 0.0, 0.5), box)).To(BeTrue())
			Expect(isec3d.CheckSphereBox(newSphere(1.6, 0.0, 0.0, 0.5), rotated)).To(BeFalse())
		})
	})

	Describe("ResolveSphereBox", func() {
		resolve := func(sphere shape3d.Sphere, b shape3d.Box) (shape3d.Contact, bool) {
			var sink shape3d.LastContact
			isec3d.ResolveSphereBox(sphere, b, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact against a face", func() {
			contact, ok := resolve(newSphere(3.0, 0.0, 0.0, 1.5), box)
			Expect(ok).To(BeTrue())
			// Normal points from the box (target) toward the sphere (source).
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			// Contact lies on the right face.
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(2.0, 0.0, 0.0))
			// Depth is radius minus the gap to the face: 1.5 - (3 - 2) = 0.5.
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("resolves contacts against each face with the correct outward normal", func() {
			right, _ := resolve(newSphere(3.0, 0.0, 0.0, 1.5), box)
			Expect(right.TargetNormal).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))

			left, _ := resolve(newSphere(-3.0, 0.0, 0.0, 1.5), box)
			Expect(left.TargetNormal).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))

			top, _ := resolve(newSphere(0.0, 2.0, 0.0, 1.5), box)
			Expect(top.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))

			bottom, _ := resolve(newSphere(0.0, -2.0, 0.0, 1.5), box)
			Expect(bottom.TargetNormal).To(dprectest.HaveVec3Coords(0.0, -1.0, 0.0))

			front, _ := resolve(newSphere(0.0, 0.0, 4.0, 1.5), box)
			Expect(front.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))

			back, _ := resolve(newSphere(0.0, 0.0, -4.0, 1.5), box)
			Expect(back.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 0.0, -1.0))
		})

		It("yields a contact against an edge", func() {
			contact, ok := resolve(newSphere(3.0, 2.0, 0.0, 1.5), box)
			Expect(ok).To(BeTrue())
			// Contact lies on the x=2,y=1 edge.
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(2.0, 1.0, 0.0))
			// Normal points outward along the edge diagonal.
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(
				1.0/dprec.Sqrt(2.0), 1.0/dprec.Sqrt(2.0), 0.0,
			))
			// Depth is radius minus the edge distance: 1.5 - sqrt(2).
			Expect(contact.Depth).To(BeNumerically("~", 1.5-dprec.Sqrt(2.0), 1e-6))
		})

		It("yields a contact against a corner", func() {
			contact, ok := resolve(newSphere(2.5, 1.5, 3.5, 1.5), box)
			Expect(ok).To(BeTrue())
			// Contact lies on the (2,1,3) corner.
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(2.0, 1.0, 3.0))
			// Normal points outward along the corner diagonal.
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(
				1.0/dprec.Sqrt(3.0), 1.0/dprec.Sqrt(3.0), 1.0/dprec.Sqrt(3.0),
			))
			// Depth is radius minus the corner distance: 1.5 - sqrt(0.75).
			Expect(contact.Depth).To(BeNumerically("~", 1.5-dprec.Sqrt(0.75), 1e-6))
		})

		It("yields a contact along the least-penetration axis when inside the box", func() {
			// Center is nearest to the right face (0.5 away), so resolution is
			// along +X.
			contact, ok := resolve(newSphere(1.5, 0.0, 0.0, 0.5), box)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(2.0, 0.0, 0.0))
			// Depth carries the center to the far side of the face: 0.5 + 0.5.
			Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("yields a zero-depth contact for a sphere that just touches", func() {
			contact, ok := resolve(newSphere(3.5, 0.0, 0.0, 1.5), box)
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("does not yield a contact for a disjoint sphere", func() {
			_, ok := resolve(newSphere(3.6, 0.0, 0.0, 1.5), box)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact past three faces but short of the corner", func() {
			_, ok := resolve(newSphere(2.5, 1.5, 3.5, 0.8), box)
			Expect(ok).To(BeFalse())
		})

		It("reports a unit normal for every contact kind", func() {
			for _, sphere := range []shape3d.Sphere{
				newSphere(3.0, 0.0, 0.0, 1.5), // face
				newSphere(3.0, 2.0, 0.0, 1.5), // edge
				newSphere(2.5, 1.5, 3.5, 1.5), // corner
				newSphere(1.5, 0.0, 0.0, 0.5), // inside
			} {
				contact, ok := resolve(sphere, box)
				Expect(ok).To(BeTrue())
				Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
			}
		})

		It("removes the overlap when the sphere is moved by Depth along the normal", func() {
			for _, sphere := range []shape3d.Sphere{
				newSphere(3.0, 0.0, 0.0, 1.5), // face
				newSphere(3.0, 2.0, 0.0, 1.5), // edge
				newSphere(2.5, 1.5, 3.5, 1.5), // corner
			} {
				contact, ok := resolve(sphere, box)
				Expect(ok).To(BeTrue())

				moved := shape3d.Sphere{
					Center: dprec.Vec3Sum(sphere.Center, dprec.Vec3Prod(contact.TargetNormal, contact.Depth)),
					Radius: sphere.Radius,
				}
				// After moving out by Depth the sphere only just touches, so a
				// re-resolve reports (essentially) zero remaining penetration.
				resolved, ok := resolve(moved, box)
				if ok {
					Expect(resolved.Depth).To(BeNumerically("~", 0.0, 1e-6))
				}
			}
		})

		It("reports the normal in world space for an oriented box", func() {
			// Rotate the box 90 degrees about Z: its local X axis maps to world Y.
			rotated := box
			rotated.Rotation = shape3d.RotationFromQuat(
				dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3()),
			)
			// The sphere sits beyond the box's local right face, which now faces
			// world +Y (the box spans y:[-2,2] after rotation).
			contact, ok := resolve(newSphere(0.0, 3.0, 0.0, 1.5), rotated)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.0, 2.0, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})
	})
})
