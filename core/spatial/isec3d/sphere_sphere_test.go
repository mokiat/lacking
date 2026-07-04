package isec3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("SphereSphere", func() {
	// A unit sphere centered at the origin.
	var sphere shape3d.Sphere

	BeforeEach(func() {
		sphere = shape3d.Sphere{
			Center: dprec.NewVec3(0.0, 0.0, 0.0),
			Radius: 1.0,
		}
	})

	Describe("CheckSphereSphere", func() {
		It("returns true for overlapping spheres", func() {
			other := shape3d.Sphere{
				Center: dprec.NewVec3(1.5, 0.0, 0.0),
				Radius: 1.0,
			}
			Expect(isec3d.CheckSphereSphere(sphere, other)).To(BeTrue())
		})

		It("returns true regardless of argument order", func() {
			other := shape3d.Sphere{
				Center: dprec.NewVec3(1.5, 0.0, 0.0),
				Radius: 1.0,
			}
			Expect(isec3d.CheckSphereSphere(other, sphere)).To(BeTrue())
		})

		It("returns true for spheres that just touch", func() {
			other := shape3d.Sphere{
				Center: dprec.NewVec3(2.0, 0.0, 0.0),
				Radius: 1.0,
			}
			Expect(isec3d.CheckSphereSphere(sphere, other)).To(BeTrue())
		})

		It("returns false for disjoint spheres", func() {
			other := shape3d.Sphere{
				Center: dprec.NewVec3(2.1, 0.0, 0.0),
				Radius: 1.0,
			}
			Expect(isec3d.CheckSphereSphere(sphere, other)).To(BeFalse())
		})

		It("returns true when one sphere fully contains the other", func() {
			big := shape3d.Sphere{
				Center: dprec.NewVec3(0.0, 0.0, 0.0),
				Radius: 5.0,
			}
			small := shape3d.Sphere{
				Center: dprec.NewVec3(1.0, 0.0, 0.0),
				Radius: 0.5,
			}
			Expect(isec3d.CheckSphereSphere(big, small)).To(BeTrue())
		})

		It("returns true for concentric spheres", func() {
			other := shape3d.Sphere{
				Center: dprec.NewVec3(0.0, 0.0, 0.0),
				Radius: 2.0,
			}
			Expect(isec3d.CheckSphereSphere(sphere, other)).To(BeTrue())
		})
	})

	Describe("ResolveSphereSphere", func() {
		It("yields a contact on the second sphere's surface", func() {
			first := shape3d.Sphere{
				Center: dprec.NewVec3(0.0, 0.0, 0.0),
				Radius: 2.0,
			}
			second := shape3d.Sphere{
				Center: dprec.NewVec3(3.0, 0.0, 0.0),
				Radius: 2.0,
			}
			var sink shape3d.LastContact
			isec3d.ResolveSphereSphere(first, second, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			// Normal points from the second (target) toward the first (source).
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))
			// Point on the second sphere's surface facing the first.
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			// Overlap is (2 + 2) - 3 = 1.
			Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("flips the contact when the arguments are swapped", func() {
			first := shape3d.Sphere{
				Center: dprec.NewVec3(0.0, 0.0, 0.0),
				Radius: 2.0,
			}
			second := shape3d.Sphere{
				Center: dprec.NewVec3(3.0, 0.0, 0.0),
				Radius: 2.0,
			}
			var sinkAB, sinkBA shape3d.LastContact
			isec3d.ResolveSphereSphere(first, second, sinkAB.AddContact)
			isec3d.ResolveSphereSphere(second, first, sinkBA.AddContact)

			ab, _ := sinkAB.Contact()
			ba, _ := sinkBA.Contact()
			Expect(ba.TargetNormal).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(ab.Depth).To(BeNumerically("~", ba.Depth, 1e-6))
		})

		It("yields a zero-depth contact for spheres that just touch", func() {
			other := shape3d.Sphere{
				Center: dprec.NewVec3(2.0, 0.0, 0.0),
				Radius: 1.0,
			}
			var sink shape3d.LastContact
			isec3d.ResolveSphereSphere(sphere, other, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("does not yield a contact for disjoint spheres", func() {
			other := shape3d.Sphere{
				Center: dprec.NewVec3(2.1, 0.0, 0.0),
				Radius: 1.0,
			}
			var sink shape3d.LastContact
			isec3d.ResolveSphereSphere(sphere, other, sink.AddContact)

			_, ok := sink.Contact()
			Expect(ok).To(BeFalse())
		})

		It("reports the contact point on the target surface along the normal", func() {
			first := shape3d.Sphere{
				Center: dprec.NewVec3(1.0, 1.0, 1.0),
				Radius: 1.5,
			}
			second := shape3d.Sphere{
				Center: dprec.NewVec3(2.0, 1.0, 1.0),
				Radius: 1.0,
			}
			var sink shape3d.LastContact
			isec3d.ResolveSphereSphere(first, second, sink.AddContact)
			contact, _ := sink.Contact()

			// The contact point lies on the second sphere's surface ...
			Expect(dprec.Vec3Diff(contact.TargetPoint, second.Center).Length()).To(BeNumerically("~", second.Radius, 1e-6))
			// ... and the normal is a unit vector.
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("separates the spheres when the source is moved by Depth along the normal", func() {
			first := shape3d.Sphere{
				Center: dprec.NewVec3(0.0, 0.0, 0.0),
				Radius: 2.0,
			}
			second := shape3d.Sphere{
				Center: dprec.NewVec3(3.0, 0.0, 0.0),
				Radius: 2.0,
			}
			var sink shape3d.LastContact
			isec3d.ResolveSphereSphere(first, second, sink.AddContact)
			contact, _ := sink.Contact()

			moved := shape3d.Sphere{
				Center: dprec.Vec3Sum(first.Center, dprec.Vec3Prod(contact.TargetNormal, contact.Depth)),
				Radius: first.Radius,
			}
			// After the move the spheres only just touch, so no overlap remains.
			centerDistance := dprec.Vec3Diff(moved.Center, second.Center).Length()
			Expect(centerDistance).To(BeNumerically("~", first.Radius+second.Radius, 1e-6))
		})

		It("handles concentric spheres without producing NaNs", func() {
			other := shape3d.Sphere{
				Center: dprec.NewVec3(0.0, 0.0, 0.0),
				Radius: 2.0,
			}
			var sink shape3d.LastContact
			isec3d.ResolveSphereSphere(sphere, other, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
			Expect(dprec.Vec3Diff(contact.TargetPoint, other.Center).Length()).To(BeNumerically("~", other.Radius, 1e-6))
			// Full overlap along the chosen normal: 1 + 2 - 0 = 3.
			Expect(contact.Depth).To(BeNumerically("~", 3.0, 1e-6))
		})
	})
})
