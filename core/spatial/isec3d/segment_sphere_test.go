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
		It("returns true for a segment entering the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 0.0, 0.0),
				B: dprec.NewVec3(2.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeTrue())
		})

		It("returns true entering from the opposite direction", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(2.0, 0.0, 0.0),
				B: dprec.NewVec3(-2.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeTrue())
		})

		It("returns false for a segment lying entirely inside the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-0.5, 0.0, 0.0),
				B: dprec.NewVec3(0.5, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeFalse())
		})

		It("returns false for a segment that starts inside the sphere", func() {
			// The start is inside, so the only crossing is an exit, which is culled.
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 0.0, 0.0),
				B: dprec.NewVec3(3.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeFalse())
		})

		It("returns false for a segment that misses the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 2.0, 0.0),
				B: dprec.NewVec3(2.0, 2.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeFalse())
		})

		It("returns false when the sphere lies beyond the segment's extent", func() {
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

		It("returns true when the far endpoint just reaches the surface", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(3.0, 0.0, 0.0),
				B: dprec.NewVec3(1.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeTrue())
		})

		It("returns false when the start lies on the surface and leaves", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(1.0, 0.0, 0.0),
				B: dprec.NewVec3(3.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeFalse())
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

		It("returns false for a degenerate point segment inside the sphere", func() {
			// A zero-length segment has no direction, so it cannot enter the
			// sphere even when its point lies inside.
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.5, 0.0, 0.0),
				B: dprec.NewVec3(0.5, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphere(seg, sphere)).To(BeFalse())
		})
	})

	Describe("CheckSegmentSphereOverlap", func() {
		It("returns true for a segment passing through the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 0.0, 0.0),
				B: dprec.NewVec3(2.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphereOverlap(seg, sphere)).To(BeTrue())
		})

		It("returns true regardless of endpoint order", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(2.0, 0.0, 0.0),
				B: dprec.NewVec3(-2.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphereOverlap(seg, sphere)).To(BeTrue())
		})

		It("returns true for a segment lying entirely inside the sphere", func() {
			// Unlike the oriented check, containment counts as an overlap.
			seg := shape3d.Segment{
				A: dprec.NewVec3(-0.5, 0.0, 0.0),
				B: dprec.NewVec3(0.5, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphereOverlap(seg, sphere)).To(BeTrue())
		})

		It("returns true for a segment that starts inside the sphere", func() {
			// Unlike the oriented check, a segment that exits without entering
			// from outside still counts as an overlap.
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 0.0, 0.0),
				B: dprec.NewVec3(3.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphereOverlap(seg, sphere)).To(BeTrue())
		})

		It("returns true for a segment tangent to the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 1.0, 0.0),
				B: dprec.NewVec3(2.0, 1.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphereOverlap(seg, sphere)).To(BeTrue())
		})

		It("returns false for a segment that misses the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 2.0, 0.0),
				B: dprec.NewVec3(2.0, 2.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphereOverlap(seg, sphere)).To(BeFalse())
		})

		It("returns false when the sphere lies beyond the segment's far end", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-3.0, 0.0, 0.0),
				B: dprec.NewVec3(-2.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphereOverlap(seg, sphere)).To(BeFalse())
		})

		It("returns false when the sphere lies behind the segment's start", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(2.0, 0.0, 0.0),
				B: dprec.NewVec3(3.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphereOverlap(seg, sphere)).To(BeFalse())
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
			Expect(isec3d.CheckSegmentSphereOverlap(hit, offsetSphere)).To(BeTrue())
			Expect(isec3d.CheckSegmentSphereOverlap(miss, offsetSphere)).To(BeFalse())
		})

		It("treats a degenerate point segment as inside the sphere", func() {
			// A zero-length segment overlaps the sphere when its point lies within
			// it, on its surface, but not when it lies outside.
			inside := shape3d.Segment{
				A: dprec.NewVec3(0.5, 0.0, 0.0),
				B: dprec.NewVec3(0.5, 0.0, 0.0),
			}
			onSurface := shape3d.Segment{
				A: dprec.NewVec3(1.0, 0.0, 0.0),
				B: dprec.NewVec3(1.0, 0.0, 0.0),
			}
			outside := shape3d.Segment{
				A: dprec.NewVec3(2.0, 0.0, 0.0),
				B: dprec.NewVec3(2.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphereOverlap(inside, sphere)).To(BeTrue())
			Expect(isec3d.CheckSegmentSphereOverlap(onSurface, sphere)).To(BeTrue())
			Expect(isec3d.CheckSegmentSphereOverlap(outside, sphere)).To(BeFalse())
		})

		It("returns true when the segment touches the sphere only at an endpoint", func() {
			// The boundary is inclusive: a single point of contact at either
			// endpoint, with the rest of the segment outside, still overlaps.
			startOnSurface := shape3d.Segment{
				A: dprec.NewVec3(1.0, 0.0, 0.0),
				B: dprec.NewVec3(3.0, 0.0, 0.0),
			}
			endOnSurface := shape3d.Segment{
				A: dprec.NewVec3(3.0, 0.0, 0.0),
				B: dprec.NewVec3(1.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSphereOverlap(startOnSurface, sphere)).To(BeTrue())
			Expect(isec3d.CheckSegmentSphereOverlap(endOnSurface, sphere)).To(BeTrue())
		})
	})

	Describe("ResolveSegmentSphere", func() {
		It("yields a contact where the segment enters the surface", func() {
			// Enters the -X side at (-1, 0, 0); B sits at the center, 1 past it.
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 0.0, 0.0),
				B: dprec.NewVec3(0.0, 0.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSphere(seg, sphere, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
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

		It("does not yield a contact when the segment starts inside the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 0.2, 0.0),
				B: dprec.NewVec3(3.0, 0.2, 0.0),
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

		It("reports the entry point on the surface with a unit normal", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 0.5, 0.3),
				B: dprec.NewVec3(2.0, 0.5, 0.3),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSphere(seg, sphere, sink.AddContact)
			contact, ok := sink.Contact()

			Expect(ok).To(BeTrue())
			// The entry point lies on the sphere surface ...
			Expect(dprec.Vec3Diff(contact.TargetPoint, sphere.Center).Length()).To(BeNumerically("~", sphere.Radius, 1e-6))
			// ... it is on the near (-X) side the segment comes from ...
			Expect(contact.TargetPoint.X).To(BeNumerically("<", 0.0))
			// ... and the normal is a unit vector.
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("brings the far endpoint onto the surface when moved by Depth", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 0.0, 0.0),
				B: dprec.NewVec3(0.0, 0.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSphere(seg, sphere, sink.AddContact)
			contact, _ := sink.Contact()

			movedB := dprec.Vec3Sum(seg.B, dprec.Vec3Prod(contact.TargetNormal, contact.Depth))
			Expect(dprec.Vec3Diff(movedB, sphere.Center).Length()).To(BeNumerically("~", sphere.Radius, 1e-6))
		})

		It("does not yield a contact for a degenerate point segment inside the sphere", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.5, 0.0, 0.0),
				B: dprec.NewVec3(0.5, 0.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSphere(seg, sphere, sink.AddContact)

			_, ok := sink.Contact()
			Expect(ok).To(BeFalse())
		})
	})
})
