package isec3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("SphereTriangle", func() {
	// A triangle in the z=0 plane, wound counter-clockwise when viewed from +Z
	// so its normal points along +Z. It covers x>=0, y>=0, x+y<=4.
	var triangle shape3d.Triangle

	newSphere := func(x, y, z, radius float64) shape3d.Sphere {
		return shape3d.Sphere{
			Center: dprec.NewVec3(x, y, z),
			Radius: radius,
		}
	}

	BeforeEach(func() {
		triangle = shape3d.Triangle{
			A: dprec.NewVec3(0.0, 0.0, 0.0),
			B: dprec.NewVec3(4.0, 0.0, 0.0),
			C: dprec.NewVec3(0.0, 4.0, 0.0),
		}
	})

	Describe("CheckSphereTriangle", func() {
		It("returns true for a sphere above the face", func() {
			Expect(isec3d.CheckSphereTriangle(newSphere(1.0, 1.0, 1.0, 1.5), triangle)).To(BeTrue())
		})

		It("returns true for a sphere that just touches the face", func() {
			Expect(isec3d.CheckSphereTriangle(newSphere(1.0, 1.0, 1.5, 1.5), triangle)).To(BeTrue())
		})

		It("returns false for a sphere too far in front of the face", func() {
			Expect(isec3d.CheckSphereTriangle(newSphere(1.0, 1.0, 2.0, 1.5), triangle)).To(BeFalse())
		})

		It("returns false for a sphere centered behind the triangle", func() {
			// Back-face culled: although it overlaps the plane, its center is on
			// the far side of the triangle's normal.
			Expect(isec3d.CheckSphereTriangle(newSphere(1.0, 1.0, -1.0, 1.5), triangle)).To(BeFalse())
		})

		It("returns false for a sphere centered exactly on the triangle's plane", func() {
			// The center must lie strictly in front of the plane (height > 0), so
			// a center on the plane itself is culled.
			Expect(isec3d.CheckSphereTriangle(newSphere(1.0, 1.0, 0.0, 1.5), triangle)).To(BeFalse())
		})

		It("returns true for a sphere reaching edge AB", func() {
			// Center sits in front of the plane and off the y=0 edge; the closest
			// point (2,0,0) is sqrt(0.8^2 + 0.6^2) = 1 away.
			Expect(isec3d.CheckSphereTriangle(newSphere(2.0, -0.8, 0.6, 1.5), triangle)).To(BeTrue())
		})

		It("returns true for a sphere reaching edge BC", func() {
			// Nearest feature is the x+y=4 edge, in front of the plane.
			Expect(isec3d.CheckSphereTriangle(newSphere(3.0, 3.0, 0.3, 1.5), triangle)).To(BeTrue())
		})

		It("returns false for a sphere short of an edge", func() {
			Expect(isec3d.CheckSphereTriangle(newSphere(2.0, -2.0, 0.6, 1.5), triangle)).To(BeFalse())
		})

		It("returns true for a sphere reaching vertex A", func() {
			// Beyond the A corner and in front of the plane; the center is
			// sqrt(3) ~= 1.732 from the corner.
			Expect(isec3d.CheckSphereTriangle(newSphere(-1.0, -1.0, 1.0, 2.0), triangle)).To(BeTrue())
		})

		It("returns true for a sphere reaching vertex C", func() {
			Expect(isec3d.CheckSphereTriangle(newSphere(-1.0, 4.5, 0.3, 1.5), triangle)).To(BeTrue())
		})

		It("returns false for a sphere short of a vertex", func() {
			Expect(isec3d.CheckSphereTriangle(newSphere(-1.5, -1.5, 0.5, 1.5), triangle)).To(BeFalse())
		})

		It("returns true for a sphere whose center is just above the face interior", func() {
			Expect(isec3d.CheckSphereTriangle(newSphere(1.0, 1.0, 0.25, 0.5), triangle)).To(BeTrue())
		})
	})

	Describe("ResolveSphereTriangle", func() {
		resolve := func(sphere shape3d.Sphere, t shape3d.Triangle) (shape3d.Contact, bool) {
			var sink shape3d.LastContact
			isec3d.ResolveSphereTriangle(sphere, t, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact against the face", func() {
			contact, ok := resolve(newSphere(1.0, 1.0, 1.0, 1.5), triangle)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(1.0, 1.0, 0.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
			// Depth is radius minus the gap to the plane: 1.5 - 1.0.
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("yields a contact against an edge", func() {
			contact, ok := resolve(newSphere(2.0, -0.8, 0.6, 1.5), triangle)
			Expect(ok).To(BeTrue())
			// Closest point lies on edge AB.
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(2.0, 0.0, 0.0))
			// Outward normal points away from the edge toward the sphere, which
			// sits 0.8 to the side and 0.6 above the closest point.
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, -0.8, 0.6))
			// Depth is radius minus the distance to the edge: 1.5 - 1.0.
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("yields a contact against a vertex", func() {
			contact, ok := resolve(newSphere(-1.0, -1.0, 1.0, 2.0), triangle)
			Expect(ok).To(BeTrue())
			// Closest point is vertex A.
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(
				-1.0/dprec.Sqrt(3.0), -1.0/dprec.Sqrt(3.0), 1.0/dprec.Sqrt(3.0),
			))
			Expect(contact.Depth).To(BeNumerically("~", 2.0-dprec.Sqrt(3.0), 1e-6))
		})

		It("yields a contact along the face normal when the center is above the interior", func() {
			contact, ok := resolve(newSphere(1.0, 1.0, 0.5, 1.0), triangle)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(1.0, 1.0, 0.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("yields a zero-depth contact for a sphere that just touches", func() {
			contact, ok := resolve(newSphere(1.0, 1.0, 1.5, 1.5), triangle)
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("does not yield a contact for a sphere centered behind the triangle", func() {
			_, ok := resolve(newSphere(1.0, 1.0, -1.0, 1.5), triangle)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact for a sphere too far in front", func() {
			_, ok := resolve(newSphere(1.0, 1.0, 2.0, 1.5), triangle)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact for a sphere short of an edge", func() {
			_, ok := resolve(newSphere(2.0, -2.0, 0.6, 1.5), triangle)
			Expect(ok).To(BeFalse())
		})

		It("reports a unit normal for every contact kind", func() {
			for _, sphere := range []shape3d.Sphere{
				newSphere(1.0, 1.0, 1.0, 1.5),   // face
				newSphere(2.0, -0.8, 0.6, 1.5),  // edge
				newSphere(-1.0, -1.0, 1.0, 2.0), // vertex
				newSphere(1.0, 1.0, 0.5, 1.0),   // inside
			} {
				contact, ok := resolve(sphere, triangle)
				Expect(ok).To(BeTrue())
				Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
			}
		})

		It("removes the overlap when the sphere is moved by Depth along the normal", func() {
			for _, sphere := range []shape3d.Sphere{
				newSphere(1.0, 1.0, 1.0, 1.5),   // face
				newSphere(2.0, -0.8, 0.6, 1.5),  // edge
				newSphere(-1.0, -1.0, 1.0, 2.0), // vertex
				newSphere(1.0, 1.0, 0.5, 1.0),   // inside
			} {
				contact, ok := resolve(sphere, triangle)
				Expect(ok).To(BeTrue())

				movedCenter := dprec.Vec3Sum(sphere.Center, dprec.Vec3Prod(contact.TargetNormal, contact.Depth))
				// After the move the closest point sits exactly at radius, so the
				// sphere only just touches the triangle.
				distance := dprec.Vec3Diff(movedCenter, contact.TargetPoint).Length()
				Expect(distance).To(BeNumerically("~", sphere.Radius, 1e-6))
			}
		})
	})
})
