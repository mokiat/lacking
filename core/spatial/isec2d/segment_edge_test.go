package isec2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("SegmentEdge", func() {
	// An edge running from (0,0) up to (0,4). Its normal is the right-hand
	// perpendicular of the A-to-B direction, which points along +X, so the front
	// side of the edge is to the right (+X).
	var edge shape2d.Edge

	BeforeEach(func() {
		edge = shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(0.0, 4.0))
	})

	newSegment := func(ax, ay, bx, by float64) shape2d.Segment {
		return shape2d.NewSegment(dprec.NewVec2(ax, ay), dprec.NewVec2(bx, by))
	}

	Describe("CheckSegmentEdge", func() {
		It("returns true for a segment crossing the edge from the front", func() {
			Expect(isec2d.CheckSegmentEdge(newSegment(1.0, 2.0, -1.0, 2.0), edge)).To(BeTrue())
		})

		It("returns false for a segment crossing from behind the edge", func() {
			// Same crossing point, but travelling from the back side to the front.
			Expect(isec2d.CheckSegmentEdge(newSegment(-1.0, 2.0, 1.0, 2.0), edge)).To(BeFalse())
		})

		It("returns false for a segment parallel to the edge", func() {
			Expect(isec2d.CheckSegmentEdge(newSegment(1.0, 0.0, 1.0, 4.0), edge)).To(BeFalse())
		})

		It("returns false when the crossing falls beyond the edge span", func() {
			// Crosses the supporting line at y=5, above endpoint B.
			Expect(isec2d.CheckSegmentEdge(newSegment(1.0, 5.0, -1.0, 5.0), edge)).To(BeFalse())
		})

		It("returns false when the edge lies beyond the segment's extent", func() {
			// Front-facing and aimed at the edge, but stops short of the line.
			Expect(isec2d.CheckSegmentEdge(newSegment(1.0, 2.0, 0.5, 2.0), edge)).To(BeFalse())
		})

		It("returns false for a segment that misses the edge entirely", func() {
			Expect(isec2d.CheckSegmentEdge(newSegment(3.0, 5.0, 1.0, 5.0), edge)).To(BeFalse())
		})
	})

	Describe("ResolveSegmentEdge", func() {
		resolve := func(segment shape2d.Segment) (shape2d.Contact, bool) {
			var sink shape2d.LastContact
			isec2d.ResolveSegmentEdge(segment, edge, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact where the segment crosses the edge", func() {
			contact, ok := resolve(newSegment(1.0, 2.0, -1.0, 2.0))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 2.0))
			// Normal is the edge's outward normal, facing back toward the segment's
			// origin side.
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(1.0, 0.0))
			// The crossing is at the segment midpoint, so half the segment lies
			// beyond it.
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("reports a depth near 1 when the segment crosses close to A", func() {
			// A just in front of the edge, B far behind it.
			contact, ok := resolve(newSegment(0.1, 2.0, -3.9, 2.0))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 2.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.975, 1e-6))
		})

		It("reports a depth near 0 when the segment crosses close to B", func() {
			// A far in front of the edge, B just behind it.
			contact, ok := resolve(newSegment(3.9, 2.0, -0.1, 2.0))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 2.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.025, 1e-6))
		})

		It("yields a unit normal", func() {
			contact, ok := resolve(newSegment(1.0, 2.0, -1.0, 2.0))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("does not yield a contact for a segment crossing from behind", func() {
			_, ok := resolve(newSegment(-1.0, 2.0, 1.0, 2.0))
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact when the crossing is beyond the edge span", func() {
			_, ok := resolve(newSegment(1.0, 5.0, -1.0, 5.0))
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact when the edge is beyond the segment's extent", func() {
			_, ok := resolve(newSegment(1.0, 2.0, 0.5, 2.0))
			Expect(ok).To(BeFalse())
		})
	})
})
