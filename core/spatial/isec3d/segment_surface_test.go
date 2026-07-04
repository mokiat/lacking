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
		It("returns true for a segment crossing from front to back", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 2.0, 0.0),
				B: dprec.NewVec3(0.0, -2.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSurface(seg, surface)).To(BeTrue())
		})

		It("returns false for a segment crossing from back to front", func() {
			// Face-culled: the segment approaches the back of the surface.
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, -2.0, 0.0),
				B: dprec.NewVec3(0.0, 2.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSurface(seg, surface)).To(BeFalse())
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

		It("returns true when the start lies on the surface and goes back", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 0.0, 0.0),
				B: dprec.NewVec3(0.0, -3.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSurface(seg, surface)).To(BeTrue())
		})

		It("returns true when the far endpoint just reaches the surface from the front", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 3.0, 0.0),
				B: dprec.NewVec3(0.0, 0.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSurface(seg, surface)).To(BeTrue())
		})

		It("respects a surface that does not pass through the origin", func() {
			raised := shape3d.Surface{
				Normal:   dprec.BasisYVec3(),
				Distance: 5.0,
			}
			behind := shape3d.Segment{
				A: dprec.NewVec3(0.0, 0.0, 0.0),
				B: dprec.NewVec3(0.0, 4.0, 0.0),
			}
			crossing := shape3d.Segment{
				A: dprec.NewVec3(0.0, 6.0, 0.0),
				B: dprec.NewVec3(0.0, 4.0, 0.0),
			}
			Expect(isec3d.CheckSegmentSurface(behind, raised)).To(BeFalse())
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
			// The crossing is at fraction 0.75, leaving 0.25 of the segment beyond it.
			Expect(contact.Depth).To(BeNumerically("~", 0.25, 1e-6))
		})

		It("does not yield a contact for a back-to-front crossing", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(1.0, -1.0, 2.0),
				B: dprec.NewVec3(1.0, 3.0, 2.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSurface(seg, surface, sink.AddContact)

			_, ok := sink.Contact()
			Expect(ok).To(BeFalse())
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

		It("yields a zero-depth contact when the far endpoint rests on the surface", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 2.0, 0.0),
				B: dprec.NewVec3(0.0, 0.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSurface(seg, surface, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("reports a depth equal to the fraction of the segment beyond the crossing", func() {
			seg := shape3d.Segment{
				A: dprec.NewVec3(0.0, 2.0, 0.0),
				B: dprec.NewVec3(0.0, -3.0, 0.0),
			}
			var sink shape3d.LastContact
			isec3d.ResolveSegmentSurface(seg, surface, sink.AddContact)
			contact, _ := sink.Contact()

			// The stretch from the crossing to B spans Depth of the segment.
			beyond := dprec.Vec3Diff(seg.B, contact.TargetPoint).Length()
			Expect(beyond).To(BeNumerically("~", contact.Depth*seg.Length(), 1e-6))
		})
	})
})
