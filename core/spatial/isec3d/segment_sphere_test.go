package isec3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("SegmentSphere", func() {
	// A unit sphere centered at the origin.
	var sphere shape3d.Sphere

	BeforeEach(func() {
		sphere = shape3d.Sphere{
			Center: dprec.NewVec3(0.0, 0.0, 0.0),
			Radius: 1.0,
		}
	})

	Describe("CheckSegmentSphere", func() {
		It("returns true for a segment passing through the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 0.0, 0.0),
				B: dprec.NewVec3(2.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeTrue())
		})

		It("returns true regardless of endpoint order", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(2.0, 0.0, 0.0),
				B: dprec.NewVec3(-2.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeTrue())
		})

		It("returns true for a segment lying entirely inside the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-0.5, 0.0, 0.0),
				B: dprec.NewVec3(0.5, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeTrue())
		})

		It("returns false for a segment that misses the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 2.0, 0.0),
				B: dprec.NewVec3(2.0, 2.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeFalse())
		})

		It("returns false when the sphere lies beyond the segment's extent", func() {
			// The segment's supporting line passes through the sphere, but the
			// segment itself stops short of it.
			seg := shape3d.Segment{
				A: dprec.NewVec3(-3.0, 0.0, 0.0),
				B: dprec.NewVec3(-2.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeFalse())
		})

		It("returns true for a segment tangent to the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 1.0, 0.0),
				B: dprec.NewVec3(2.0, 1.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeTrue())
		})

		It("returns true when an endpoint lies exactly on the surface", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(1.0, 0.0, 0.0),
				B: dprec.NewVec3(3.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeTrue())
		})

		It("handles a degenerate zero-length segment", func() {
			inside := shape3d.Segment{
				A: dprec.NewVec3(0.5, 0.0, 0.0),
				B: dprec.NewVec3(0.5, 0.0, 0.0),
			}
			outside := shape3d.Segment{
				A: dprec.NewVec3(5.0, 0.0, 0.0),
				B: dprec.NewVec3(5.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(inside, sphere)).To(BeTrue())
			Expect(isec3d.CheckSegmentSphere(outside, sphere)).To(BeFalse())
		})

		It("respects a sphere that is not centered at the origin", func() {
			offsetSphere := shape3d.Sphere{
				Center: dprec.NewVec3(10.0, 0.0, 0.0),
				Radius: 1.0,
			}
			hit := shape3d.Segment{
				A: dprec.NewVec3(9.0, -2.0, 0.0),
				B: dprec.NewVec3(9.0, 2.0, 0.0),
			}
			miss := shape3d.Segment{
				A: dprec.NewVec3(0.0, 0.0, 0.0),
				B: dprec.NewVec3(5.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(hit, offsetSphere)).To(BeTrue())
			Expect(isec3d.CheckSegmentSphere(miss, offsetSphere)).To(BeFalse())
		})
	})

	Describe("ResolveSegmentSphere", func() {
		It("yields a contact at the closest point on the surface", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 0.5, 0.0),
				B: dprec.NewVec3(2.0, 0.5, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSphere(seg, sphere, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			// Closest point on the segment to the center is (0, 0.5, 0), 0.5 from
			// the center, so the surface contact is at (0, 1, 0).
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("produces the same contact regardless of endpoint order", func() {
			forward := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 0.5, 0.0),
				B: dprec.NewVec3(2.0, 0.5, 0.0),
			}
			reversed := forward.Flipped()

			var sinkForward, sinkReversed shape3d.LastContact
			isec3d.ResolveSegmentSphere(forward, sphere, sinkForward.AddContact)
			isec3d.ResolveSegmentSphere(reversed, sphere, sinkReversed.AddContact)

			cf, okF := sinkForward.Contact()
			cr, okR := sinkReversed.Contact()
			Expect(okF).To(BeTrue())
			Expect(okR).To(BeTrue())
			Expect(cr.TargetPoint).To(dprectest.HaveVec3Coords(cf.TargetPoint.X, cf.TargetPoint.Y, cf.TargetPoint.Z))
			Expect(cr.TargetNormal).To(dprectest.HaveVec3Coords(cf.TargetNormal.X, cf.TargetNormal.Y, cf.TargetNormal.Z))
			Expect(cr.Depth).To(BeNumerically("~", cf.Depth, 1e-6))
		})

		It("yields a contact when the segment starts inside the sphere", func() {
			// An endpoint inside the sphere: the entry-point model would miss
			// this, but the closest-point model stays consistent with Check.
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 0.2, 0.0),
				B: dprec.NewVec3(3.0, 0.2, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSphere(seg, sphere, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.8, 1e-6))
		})

		It("does not yield a contact when the segment misses the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 2.0, 0.0),
				B: dprec.NewVec3(2.0, 2.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSphere(seg, sphere, sink.AddContact)

			_, ok := sink.Contact()
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact when the sphere is beyond the segment's extent", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-3.0, 0.0, 0.0),
				B: dprec.NewVec3(-2.0, 0.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSphere(seg, sphere, sink.AddContact)

			_, ok := sink.Contact()
			Expect(ok).To(BeFalse())
		})

		It("yields a zero-depth contact for a tangent segment", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 1.0, 0.0),
				B: dprec.NewVec3(2.0, 1.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSphere(seg, sphere, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("yields the contact point on the surface along the normal", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 0.5, 0.3),
				B: dprec.NewVec3(2.0, 0.5, 0.3),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSphere(seg, sphere, sink.AddContact)
			contact, _ := sink.Contact()

			// The contact point lies on the sphere surface ...
			Expect(dprec.Vec3Diff(contact.TargetPoint, sphere.Center).Length()).To(BeNumerically("~", sphere.Radius, 1e-6))
			// ... and the normal is a unit vector.
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("handles a segment passing through the sphere center", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-1.0, 0.0, 0.0),
				B: dprec.NewVec3(1.0, 0.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSphere(seg, sphere, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			// The separation normal is not unique here; require only that it is a
			// valid unit normal yielding a surface contact with full depth.
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
			Expect(dprec.Vec3Diff(contact.TargetPoint, sphere.Center).Length()).To(BeNumerically("~", sphere.Radius, 1e-6))
			Expect(contact.Depth).To(BeNumerically("~", sphere.Radius, 1e-6))
		})
	})
})
