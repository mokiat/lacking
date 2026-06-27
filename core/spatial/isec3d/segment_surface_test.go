package isec3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("SegmentSurface", func() {
	// A horizontal surface at height y=0 facing up (+Y).
	var surface shape3d.Surface

	BeforeEach(func() {
		surface = shape3d.BasisYSurface()
	})

	Describe("CheckSegmentSurface", func() {
		It("returns true for a segment crossing from the front to the back", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 2.0, 0.0),
				B: dprec.NewVec3(0.0, -2.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSurface(seg, surface)).To(BeTrue())
		})

		It("returns true regardless of endpoint order", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, -2.0, 0.0),
				B: dprec.NewVec3(0.0, 2.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSurface(seg, surface)).To(BeTrue())
		})

		It("returns false when both endpoints are in front of the surface", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 1.0, 0.0),
				B: dprec.NewVec3(0.0, 3.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSurface(seg, surface)).To(BeFalse())
		})

		It("returns false when both endpoints are behind the surface", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, -1.0, 0.0),
				B: dprec.NewVec3(0.0, -3.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSurface(seg, surface)).To(BeFalse())
		})

		It("returns true when an endpoint lies exactly on the surface", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 0.0, 0.0),
				B: dprec.NewVec3(0.0, 3.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSurface(seg, surface)).To(BeTrue())
		})

		It("respects a surface that does not pass through the origin", func() {
			raised := shape3d.Surface{
				Normal:   dprec.BasisYVec3(),
				Distance: 5.0,
			}
			below := shape3d.Segment{
				A: dprec.NewVec3(0.0, 0.0, 0.0),
				B: dprec.NewVec3(0.0, 4.0, 0.0),
			}
			crossing := shape3d.Segment{
				A: dprec.NewVec3(0.0, 4.0, 0.0),
				B: dprec.NewVec3(0.0, 6.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSurface(below, raised)).To(BeFalse())
			Expect(isec3d.CheckSegmentSurface(crossing, raised)).To(BeTrue())
		})
	})

	Describe("ResolveSegmentSurface", func() {
		It("yields a contact at the crossing point", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(1.0, 3.0, 2.0),
				B: dprec.NewVec3(1.0, -1.0, 2.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSurface(seg, surface, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			// Crosses y=0 three quarters of the way from A (y=3) to B (y=-1).
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(1.0, 0.0, 2.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			// Deepest endpoint B lies 1 unit behind the surface along the normal.
			Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("produces the same crossing point regardless of endpoint order", func() {
			forward := shape3d.Segment{
				A: dprec.NewVec3(1.0, 3.0, 2.0),
				B: dprec.NewVec3(1.0, -1.0, 2.0),
			}
			reversed := forward.Flipped()

			var sinkForward, sinkReversed shape3d.LastContact
			isec3d.ResolveSegmentSurface(forward, surface, sinkForward.AddContact)
			isec3d.ResolveSegmentSurface(reversed, surface, sinkReversed.AddContact)

			cf, okF := sinkForward.Contact()
			cr, okR := sinkReversed.Contact()
			Expect(okF).To(BeTrue())
			Expect(okR).To(BeTrue())
			Expect(cr.TargetPoint).To(dprectest.HaveVec3Coords(cf.TargetPoint.X, cf.TargetPoint.Y, cf.TargetPoint.Z))
			Expect(cr.Depth).To(BeNumerically("~", cf.Depth, 1e-6))
		})

		It("does not yield a contact when the segment does not cross", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 1.0, 0.0),
				B: dprec.NewVec3(0.0, 3.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSurface(seg, surface, sink.AddContact)

			_, ok := sink.Contact()
			Expect(ok).To(BeFalse())
		})

		It("yields a zero-depth contact when the segment lies on the surface", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(-2.0, 0.0, 0.0),
				B: dprec.NewVec3(2.0, 0.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSurface(seg, surface, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("moving the segment up by Depth along the normal clears the surface", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 2.0, 0.0),
				B: dprec.NewVec3(0.0, -3.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSurface(seg, surface, sink.AddContact)
			contact, _ := sink.Contact()

			lift := dprec.Vec3Prod(contact.TargetNormal, contact.Depth)
			resolved := shape3d.Segment{
				A: dprec.Vec3Sum(seg.A, lift),
				B: dprec.Vec3Sum(seg.B, lift),
			}
			// After lifting, the deepest endpoint sits exactly on the surface,
			// so no penetration remains.
			Expect(surface.SignedDistance(resolved.A)).To(BeNumerically(">=", -1e-6))
			Expect(surface.SignedDistance(resolved.B)).To(BeNumerically(">=", -1e-6))
		})
	})
})
