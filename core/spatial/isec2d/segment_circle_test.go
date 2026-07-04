package isec2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("SegmentCircle", func() {
	// A unit circle centered at the origin.
	var circle shape2d.Circle

	BeforeEach(func() {
		circle = shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 1.0)
	})

	newSegment := func(ax, ay, bx, by float64) shape2d.Segment {
		return shape2d.NewSegment(dprec.NewVec2(ax, ay), dprec.NewVec2(bx, by))
	}

	Describe("CheckSegmentCircle", func() {
		It("returns true for a segment entering the circle", func() {
			Expect(isec2d.CheckSegmentCircle(newSegment(-2.0, 0.0, 2.0, 0.0), circle)).To(BeTrue())
		})

		It("returns true entering from the opposite direction", func() {
			Expect(isec2d.CheckSegmentCircle(newSegment(2.0, 0.0, -2.0, 0.0), circle)).To(BeTrue())
		})

		It("returns false for a segment lying entirely inside the circle", func() {
			Expect(isec2d.CheckSegmentCircle(newSegment(-0.5, 0.0, 0.5, 0.0), circle)).To(BeFalse())
		})

		It("returns false for a segment that starts inside the circle", func() {
			// The start is inside, so the only crossing is an exit, which is culled.
			Expect(isec2d.CheckSegmentCircle(newSegment(0.0, 0.0, 3.0, 0.0), circle)).To(BeFalse())
		})

		It("returns false for a segment that misses the circle", func() {
			Expect(isec2d.CheckSegmentCircle(newSegment(-2.0, 2.0, 2.0, 2.0), circle)).To(BeFalse())
		})

		It("returns false when the circle lies beyond the segment's extent", func() {
			Expect(isec2d.CheckSegmentCircle(newSegment(-3.0, 0.0, -2.0, 0.0), circle)).To(BeFalse())
		})

		It("returns true for a segment tangent to the circle", func() {
			Expect(isec2d.CheckSegmentCircle(newSegment(-2.0, 1.0, 2.0, 1.0), circle)).To(BeTrue())
		})

		It("returns true when the far endpoint just reaches the perimeter", func() {
			Expect(isec2d.CheckSegmentCircle(newSegment(3.0, 0.0, 1.0, 0.0), circle)).To(BeTrue())
		})

		It("returns false when the start lies on the perimeter and leaves", func() {
			Expect(isec2d.CheckSegmentCircle(newSegment(1.0, 0.0, 3.0, 0.0), circle)).To(BeFalse())
		})

		It("respects a circle that is not centered at the origin", func() {
			offsetCircle := shape2d.NewCircle(dprec.NewVec2(10.0, 0.0), 1.0)
			hit := newSegment(9.0, -2.0, 9.0, 2.0)
			miss := newSegment(0.0, 0.0, 5.0, 0.0)
			Expect(isec2d.CheckSegmentCircle(hit, offsetCircle)).To(BeTrue())
			Expect(isec2d.CheckSegmentCircle(miss, offsetCircle)).To(BeFalse())
		})

		It("returns false for a degenerate point segment inside the circle", func() {
			// A zero-length segment has no direction, so it cannot enter the
			// circle even when its point lies inside.
			Expect(isec2d.CheckSegmentCircle(newSegment(0.5, 0.0, 0.5, 0.0), circle)).To(BeFalse())
		})
	})

	Describe("CheckSegmentCircleOverlap", func() {
		It("returns true for a segment passing through the circle", func() {
			Expect(isec2d.CheckSegmentCircleOverlap(newSegment(-2.0, 0.0, 2.0, 0.0), circle)).To(BeTrue())
		})

		It("returns true regardless of endpoint order", func() {
			Expect(isec2d.CheckSegmentCircleOverlap(newSegment(2.0, 0.0, -2.0, 0.0), circle)).To(BeTrue())
		})

		It("returns true for a segment lying entirely inside the circle", func() {
			Expect(isec2d.CheckSegmentCircleOverlap(newSegment(-0.5, 0.0, 0.5, 0.0), circle)).To(BeTrue())
		})

		It("returns true for a segment that starts inside the circle", func() {
			Expect(isec2d.CheckSegmentCircleOverlap(newSegment(0.0, 0.0, 3.0, 0.0), circle)).To(BeTrue())
		})

		It("returns true for a segment tangent to the circle", func() {
			Expect(isec2d.CheckSegmentCircleOverlap(newSegment(-2.0, 1.0, 2.0, 1.0), circle)).To(BeTrue())
		})

		It("returns false for a segment that misses the circle", func() {
			Expect(isec2d.CheckSegmentCircleOverlap(newSegment(-2.0, 2.0, 2.0, 2.0), circle)).To(BeFalse())
		})

		It("returns false when the circle lies beyond the segment's far end", func() {
			Expect(isec2d.CheckSegmentCircleOverlap(newSegment(-3.0, 0.0, -2.0, 0.0), circle)).To(BeFalse())
		})

		It("returns false when the circle lies behind the segment's start", func() {
			Expect(isec2d.CheckSegmentCircleOverlap(newSegment(2.0, 0.0, 3.0, 0.0), circle)).To(BeFalse())
		})

		It("treats a degenerate point segment as inside the circle", func() {
			inside := newSegment(0.5, 0.0, 0.5, 0.0)
			onPerimeter := newSegment(1.0, 0.0, 1.0, 0.0)
			outside := newSegment(2.0, 0.0, 2.0, 0.0)
			Expect(isec2d.CheckSegmentCircleOverlap(inside, circle)).To(BeTrue())
			Expect(isec2d.CheckSegmentCircleOverlap(onPerimeter, circle)).To(BeTrue())
			Expect(isec2d.CheckSegmentCircleOverlap(outside, circle)).To(BeFalse())
		})

		It("returns true when the segment touches the circle only at an endpoint", func() {
			startOnPerimeter := newSegment(1.0, 0.0, 3.0, 0.0)
			endOnPerimeter := newSegment(3.0, 0.0, 1.0, 0.0)
			Expect(isec2d.CheckSegmentCircleOverlap(startOnPerimeter, circle)).To(BeTrue())
			Expect(isec2d.CheckSegmentCircleOverlap(endOnPerimeter, circle)).To(BeTrue())
		})
	})

	Describe("ResolveSegmentCircle", func() {
		resolve := func(segment shape2d.Segment) (shape2d.Contact, bool) {
			var sink shape2d.LastContact
			isec2d.ResolveSegmentCircle(segment, circle, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact where the segment enters the perimeter", func() {
			// Enters the -X side at (-1, 0), halfway along the segment, so half of
			// it lies beyond the entry point.
			contact, ok := resolve(newSegment(-2.0, 0.0, 0.0, 0.0))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(-1.0, 0.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(-1.0, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("yields a zero-depth contact when the far endpoint rests on the perimeter", func() {
			contact, ok := resolve(newSegment(-2.0, 0.0, -1.0, 0.0))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(-1.0, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("does not yield a contact when the segment misses the circle", func() {
			_, ok := resolve(newSegment(-2.0, 2.0, 2.0, 2.0))
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact when the segment starts inside the circle", func() {
			_, ok := resolve(newSegment(0.0, 0.2, 3.0, 0.2))
			Expect(ok).To(BeFalse())
		})

		It("yields a contact at the tangent point", func() {
			// The segment grazes the top of the circle at (0, 1), halfway along, so
			// half of it lies beyond that point.
			contact, ok := resolve(newSegment(-2.0, 1.0, 2.0, 1.0))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("reports the entry point on the perimeter with a unit normal", func() {
			segment := newSegment(-2.0, 0.5, 2.0, 0.5)
			contact, ok := resolve(segment)
			Expect(ok).To(BeTrue())
			// The entry point lies on the circle perimeter ...
			Expect(dprec.Vec2Diff(contact.TargetPoint, circle.Center).Length()).To(BeNumerically("~", circle.Radius, 1e-6))
			// ... it is on the near (-X) side the segment comes from ...
			Expect(contact.TargetPoint.X).To(BeNumerically("<", 0.0))
			// ... and the normal is a unit vector.
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("reports a depth equal to the fraction of the segment beyond the entry point", func() {
			segment := newSegment(-2.0, 0.0, 0.0, 0.0)
			contact, ok := resolve(segment)
			Expect(ok).To(BeTrue())

			// The stretch from the entry point to B spans Depth of the segment.
			beyond := dprec.Vec2Diff(segment.B, contact.TargetPoint).Length()
			Expect(beyond).To(BeNumerically("~", contact.Depth*segment.Length(), 1e-6))
		})

		It("does not yield a contact for a degenerate point segment inside the circle", func() {
			_, ok := resolve(newSegment(0.5, 0.0, 0.5, 0.0))
			Expect(ok).To(BeFalse())
		})
	})
})
