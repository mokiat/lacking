package isec2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("SegmentRectangle", func() {
	// A square of half-extent 1 centered at the origin, axis aligned.
	var rectangle shape2d.Rectangle

	BeforeEach(func() {
		rectangle = shape2d.NewRectangle(
			dprec.NewVec2(0.0, 0.0),
			shape2d.IdentityRotation(),
			dprec.NewVec2(1.0, 1.0),
		)
	})

	newSegment := func(ax, ay, bx, by float64) shape2d.Segment {
		return shape2d.NewSegment(dprec.NewVec2(ax, ay), dprec.NewVec2(bx, by))
	}

	// A rectangle long along its local X, rotated 90 degrees, so in world space
	// it spans +/-2 along Y and +/-0.5 along X.
	rotatedRectangle := func() shape2d.Rectangle {
		return shape2d.NewRectangle(
			dprec.NewVec2(0.0, 0.0),
			shape2d.RotationFromAngle(dprec.Degrees(90.0)),
			dprec.NewVec2(2.0, 0.5),
		)
	}

	Describe("CheckSegmentRectangle", func() {
		It("returns true for a segment entering the rectangle", func() {
			Expect(isec2d.CheckSegmentRectangle(newSegment(-3.0, 0.0, 3.0, 0.0), rectangle)).To(BeTrue())
		})

		It("returns true entering from the opposite direction", func() {
			Expect(isec2d.CheckSegmentRectangle(newSegment(3.0, 0.0, -3.0, 0.0), rectangle)).To(BeTrue())
		})

		It("returns false for a segment lying entirely inside the rectangle", func() {
			Expect(isec2d.CheckSegmentRectangle(newSegment(-0.5, 0.0, 0.5, 0.0), rectangle)).To(BeFalse())
		})

		It("returns false for a segment that starts inside the rectangle", func() {
			Expect(isec2d.CheckSegmentRectangle(newSegment(0.0, 0.0, 3.0, 0.0), rectangle)).To(BeFalse())
		})

		It("returns false for a segment that misses the rectangle", func() {
			Expect(isec2d.CheckSegmentRectangle(newSegment(-3.0, 2.0, 3.0, 2.0), rectangle)).To(BeFalse())
		})

		It("returns false when the rectangle lies beyond the segment's extent", func() {
			Expect(isec2d.CheckSegmentRectangle(newSegment(-3.0, 0.0, -2.0, 0.0), rectangle)).To(BeFalse())
		})

		It("returns false for a diagonal segment that slips past a corner", func() {
			// The line x+y=2.1 passes just outside the (1,1) corner.
			Expect(isec2d.CheckSegmentRectangle(newSegment(2.1, 0.0, 0.0, 2.1), rectangle)).To(BeFalse())
		})

		It("returns true for a diagonal segment that clips a corner", func() {
			// The line x+y=1.9 passes just inside the (1,1) corner.
			Expect(isec2d.CheckSegmentRectangle(newSegment(1.9, 0.0, 0.0, 1.9), rectangle)).To(BeTrue())
		})

		Context("with a rotated, non-square rectangle", func() {
			It("returns true for a segment along the rotated long axis", func() {
				Expect(isec2d.CheckSegmentRectangle(newSegment(0.0, -3.0, 0.0, 3.0), rotatedRectangle())).To(BeTrue())
			})

			It("returns false for a segment beyond the rotated short axis", func() {
				Expect(isec2d.CheckSegmentRectangle(newSegment(1.0, -3.0, 1.0, 3.0), rotatedRectangle())).To(BeFalse())
			})
		})
	})

	Describe("CheckSegmentRectangleOverlap", func() {
		It("returns true for a segment passing through the rectangle", func() {
			Expect(isec2d.CheckSegmentRectangleOverlap(newSegment(-3.0, 0.0, 3.0, 0.0), rectangle)).To(BeTrue())
		})

		It("returns true regardless of endpoint order", func() {
			Expect(isec2d.CheckSegmentRectangleOverlap(newSegment(3.0, 0.0, -3.0, 0.0), rectangle)).To(BeTrue())
		})

		It("returns true for a segment lying entirely inside the rectangle", func() {
			Expect(isec2d.CheckSegmentRectangleOverlap(newSegment(-0.5, 0.0, 0.5, 0.0), rectangle)).To(BeTrue())
		})

		It("returns true for a segment that starts inside the rectangle", func() {
			Expect(isec2d.CheckSegmentRectangleOverlap(newSegment(0.0, 0.0, 3.0, 0.0), rectangle)).To(BeTrue())
		})

		It("returns true for a diagonal segment that clips a corner", func() {
			Expect(isec2d.CheckSegmentRectangleOverlap(newSegment(1.9, 0.0, 0.0, 1.9), rectangle)).To(BeTrue())
		})

		It("returns false for a diagonal segment that slips past a corner", func() {
			Expect(isec2d.CheckSegmentRectangleOverlap(newSegment(2.1, 0.0, 0.0, 2.1), rectangle)).To(BeFalse())
		})

		It("returns false for a segment that misses the rectangle", func() {
			Expect(isec2d.CheckSegmentRectangleOverlap(newSegment(-3.0, 2.0, 3.0, 2.0), rectangle)).To(BeFalse())
		})

		It("returns false when the rectangle lies beyond the segment's far end", func() {
			Expect(isec2d.CheckSegmentRectangleOverlap(newSegment(-3.0, 0.0, -2.0, 0.0), rectangle)).To(BeFalse())
		})

		It("returns false when the rectangle lies behind the segment's start", func() {
			Expect(isec2d.CheckSegmentRectangleOverlap(newSegment(2.0, 0.0, 3.0, 0.0), rectangle)).To(BeFalse())
		})

		It("treats a degenerate point segment as inside the rectangle", func() {
			inside := newSegment(0.5, 0.0, 0.5, 0.0)
			onEdge := newSegment(1.0, 0.0, 1.0, 0.0)
			outside := newSegment(2.0, 0.0, 2.0, 0.0)
			Expect(isec2d.CheckSegmentRectangleOverlap(inside, rectangle)).To(BeTrue())
			Expect(isec2d.CheckSegmentRectangleOverlap(onEdge, rectangle)).To(BeTrue())
			Expect(isec2d.CheckSegmentRectangleOverlap(outside, rectangle)).To(BeFalse())
		})

		It("returns true when the segment touches the rectangle only at an endpoint", func() {
			startOnEdge := newSegment(1.0, 0.0, 3.0, 0.0)
			endOnEdge := newSegment(3.0, 0.0, 1.0, 0.0)
			Expect(isec2d.CheckSegmentRectangleOverlap(startOnEdge, rectangle)).To(BeTrue())
			Expect(isec2d.CheckSegmentRectangleOverlap(endOnEdge, rectangle)).To(BeTrue())
		})

		It("respects a rotated, non-square rectangle", func() {
			hit := newSegment(0.0, -3.0, 0.0, 3.0)
			miss := newSegment(1.0, -3.0, 1.0, 3.0)
			Expect(isec2d.CheckSegmentRectangleOverlap(hit, rotatedRectangle())).To(BeTrue())
			Expect(isec2d.CheckSegmentRectangleOverlap(miss, rotatedRectangle())).To(BeFalse())
		})
	})

	Describe("ResolveSegmentRectangle", func() {
		resolve := func(segment shape2d.Segment, r shape2d.Rectangle) (shape2d.Contact, bool) {
			var sink shape2d.LastContact
			isec2d.ResolveSegmentRectangle(segment, r, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact at the entry edge", func() {
			// Enters the top (+Y) edge at (0, 1). The segment spans y from 2 to 0.5
			// and crosses at fraction 2/3, leaving 1/3 beyond it.
			contact, ok := resolve(newSegment(0.0, 2.0, 0.0, 0.5), rectangle)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(contact.Depth).To(BeNumerically("~", 1.0/3.0, 1e-6))
		})

		It("places the contact point where the segment crosses the boundary", func() {
			// A diagonal segment that enters the top (+Y) edge at (0, 1).
			contact, ok := resolve(newSegment(-2.0, 2.0, 2.0, 0.0), rectangle)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 1.0))
		})

		It("reports the entry edge for the side the segment comes from", func() {
			// Reversing the direction makes the segment enter the +X edge instead.
			contact, ok := resolve(newSegment(3.0, 0.0, -3.0, 0.0), rectangle)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(1.0, 0.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(1.0, 0.0))
		})

		It("does not yield a contact when the segment misses the rectangle", func() {
			_, ok := resolve(newSegment(-3.0, 2.0, 3.0, 2.0), rectangle)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact when the segment starts inside the rectangle", func() {
			_, ok := resolve(newSegment(0.0, 0.0, 0.0, 3.0), rectangle)
			Expect(ok).To(BeFalse())
		})

		It("reports a contact point on the rectangle boundary with a unit normal", func() {
			contact, ok := resolve(newSegment(0.3, 2.0, 0.3, 0.5), rectangle)
			Expect(ok).To(BeTrue())
			// Enters the top edge: the entry keeps the segment's x.
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.3, 1.0))
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("reports a depth equal to the fraction of the segment beyond the entry point", func() {
			segment := newSegment(0.0, 2.0, 0.0, 0.5)
			contact, ok := resolve(segment, rectangle)
			Expect(ok).To(BeTrue())

			// The stretch from the entry point to B spans Depth of the segment.
			beyond := dprec.Vec2Diff(segment.B, contact.TargetPoint).Length()
			Expect(beyond).To(BeNumerically("~", contact.Depth*segment.Length(), 1e-6))
		})

		It("resolves against a rotated rectangle in world space", func() {
			// A rectangle long along its local X, rotated 90 degrees, spanning
			// +/-1 along world X and +/-2 along world Y. The segment enters the
			// world +X edge at (1, 0).
			rotated := shape2d.NewRectangle(
				dprec.NewVec2(0.0, 0.0),
				shape2d.RotationFromAngle(dprec.Degrees(90.0)),
				dprec.NewVec2(2.0, 1.0),
			)
			contact, ok := resolve(newSegment(3.0, 0.0, 0.0, 0.0), rotated)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(1.0, 0.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(1.0, 0.0))
			// Enters at world x=1; the segment spans x from 3 to 0, fraction 2/3.
			Expect(contact.Depth).To(BeNumerically("~", 1.0/3.0, 1e-6))
		})
	})
})
