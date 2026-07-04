package isec2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("CircleCircle", func() {
	// A unit circle centered at the origin.
	var circle shape2d.Circle

	BeforeEach(func() {
		circle = shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 1.0)
	})

	Describe("CheckCircleCircle", func() {
		It("returns true for overlapping circles", func() {
			other := shape2d.NewCircle(dprec.NewVec2(1.5, 0.0), 1.0)
			Expect(isec2d.CheckCircleCircle(circle, other)).To(BeTrue())
		})

		It("returns true regardless of argument order", func() {
			other := shape2d.NewCircle(dprec.NewVec2(1.5, 0.0), 1.0)
			Expect(isec2d.CheckCircleCircle(other, circle)).To(BeTrue())
		})

		It("returns true for circles that just touch", func() {
			other := shape2d.NewCircle(dprec.NewVec2(2.0, 0.0), 1.0)
			Expect(isec2d.CheckCircleCircle(circle, other)).To(BeTrue())
		})

		It("returns false for disjoint circles", func() {
			other := shape2d.NewCircle(dprec.NewVec2(2.1, 0.0), 1.0)
			Expect(isec2d.CheckCircleCircle(circle, other)).To(BeFalse())
		})

		It("returns true when one circle fully contains the other", func() {
			big := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 5.0)
			small := shape2d.NewCircle(dprec.NewVec2(1.0, 0.0), 0.5)
			Expect(isec2d.CheckCircleCircle(big, small)).To(BeTrue())
		})

		It("returns true for concentric circles", func() {
			other := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
			Expect(isec2d.CheckCircleCircle(circle, other)).To(BeTrue())
		})
	})

	Describe("ResolveCircleCircle", func() {
		It("yields a contact on the second circle's perimeter", func() {
			first := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
			second := shape2d.NewCircle(dprec.NewVec2(3.0, 0.0), 2.0)
			var sink shape2d.LastContact
			isec2d.ResolveCircleCircle(first, second, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			// Normal points from the second (target) toward the first (source).
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(-1.0, 0.0))
			// Point on the second circle's perimeter facing the first.
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(1.0, 0.0))
			// Overlap is (2 + 2) - 3 = 1.
			Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("flips the contact when the arguments are swapped", func() {
			first := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
			second := shape2d.NewCircle(dprec.NewVec2(3.0, 0.0), 2.0)
			var sinkAB, sinkBA shape2d.LastContact
			isec2d.ResolveCircleCircle(first, second, sinkAB.AddContact)
			isec2d.ResolveCircleCircle(second, first, sinkBA.AddContact)

			ab, _ := sinkAB.Contact()
			ba, _ := sinkBA.Contact()
			Expect(ba.TargetNormal).To(dprectest.HaveVec2Coords(1.0, 0.0))
			Expect(ab.Depth).To(BeNumerically("~", ba.Depth, 1e-6))
		})

		It("yields a zero-depth contact for circles that just touch", func() {
			other := shape2d.NewCircle(dprec.NewVec2(2.0, 0.0), 1.0)
			var sink shape2d.LastContact
			isec2d.ResolveCircleCircle(circle, other, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("does not yield a contact for disjoint circles", func() {
			other := shape2d.NewCircle(dprec.NewVec2(2.1, 0.0), 1.0)
			var sink shape2d.LastContact
			isec2d.ResolveCircleCircle(circle, other, sink.AddContact)

			_, ok := sink.Contact()
			Expect(ok).To(BeFalse())
		})

		It("reports the contact point on the target perimeter along the normal", func() {
			first := shape2d.NewCircle(dprec.NewVec2(1.0, 1.0), 1.5)
			second := shape2d.NewCircle(dprec.NewVec2(2.0, 1.0), 1.0)
			var sink shape2d.LastContact
			isec2d.ResolveCircleCircle(first, second, sink.AddContact)
			contact, _ := sink.Contact()

			// The contact point lies on the second circle's perimeter ...
			Expect(dprec.Vec2Diff(contact.TargetPoint, second.Center).Length()).To(BeNumerically("~", second.Radius, 1e-6))
			// ... and the normal is a unit vector.
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("separates the circles when the source is moved by Depth along the normal", func() {
			first := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
			second := shape2d.NewCircle(dprec.NewVec2(3.0, 0.0), 2.0)
			var sink shape2d.LastContact
			isec2d.ResolveCircleCircle(first, second, sink.AddContact)
			contact, _ := sink.Contact()

			moved := shape2d.NewCircle(
				dprec.Vec2Sum(first.Center, dprec.Vec2Prod(contact.TargetNormal, contact.Depth)),
				first.Radius,
			)
			// After the move the circles only just touch, so no overlap remains.
			centerDistance := dprec.Vec2Diff(moved.Center, second.Center).Length()
			Expect(centerDistance).To(BeNumerically("~", first.Radius+second.Radius, 1e-6))
		})

		It("handles concentric circles without producing NaNs", func() {
			other := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 2.0)
			var sink shape2d.LastContact
			isec2d.ResolveCircleCircle(circle, other, sink.AddContact)

			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
			Expect(dprec.Vec2Diff(contact.TargetPoint, other.Center).Length()).To(BeNumerically("~", other.Radius, 1e-6))
			// Full overlap along the chosen normal: 1 + 2 - 0 = 3.
			Expect(contact.Depth).To(BeNumerically("~", 3.0, 1e-6))
		})
	})
})
