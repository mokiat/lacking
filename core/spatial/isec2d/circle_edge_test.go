package isec2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("CircleEdge", func() {
	// An edge running from (0,0) up to (0,4). Its normal is the right-hand
	// perpendicular of the A-to-B direction, which points along +X, so the front
	// side is to the right of the edge.
	var edge shape2d.Edge

	newCircle := func(x, y, radius float64) shape2d.Circle {
		return shape2d.NewCircle(dprec.NewVec2(x, y), radius)
	}

	BeforeEach(func() {
		edge = shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(0.0, 4.0))
	})

	Describe("CheckCircleEdge", func() {
		It("returns true for a circle in front of the edge span", func() {
			Expect(isec2d.CheckCircleEdge(newCircle(0.5, 2.0, 1.0), edge)).To(BeTrue())
		})

		It("returns true for a circle that just touches the edge", func() {
			Expect(isec2d.CheckCircleEdge(newCircle(1.0, 2.0, 1.0), edge)).To(BeTrue())
		})

		It("returns false for a circle too far in front of the edge", func() {
			Expect(isec2d.CheckCircleEdge(newCircle(2.0, 2.0, 1.0), edge)).To(BeFalse())
		})

		It("returns false for a circle centered behind the edge", func() {
			// Back-face culled: although it overlaps the edge's line, its center is
			// on the far side of the edge's normal.
			Expect(isec2d.CheckCircleEdge(newCircle(-1.0, 2.0, 1.0), edge)).To(BeFalse())
		})

		It("returns false for a circle centered exactly on the edge's line", func() {
			Expect(isec2d.CheckCircleEdge(newCircle(0.0, 2.0, 1.0), edge)).To(BeFalse())
		})

		It("returns true for a circle reaching endpoint A", func() {
			// Beyond the A end (below it) and in front; the center is sqrt(0.5)
			// from the corner.
			Expect(isec2d.CheckCircleEdge(newCircle(0.5, -0.5, 1.0), edge)).To(BeTrue())
		})

		It("returns true for a circle reaching endpoint B", func() {
			Expect(isec2d.CheckCircleEdge(newCircle(0.5, 4.5, 1.0), edge)).To(BeTrue())
		})

		It("returns false for a circle short of an endpoint", func() {
			// Beyond endpoint A by sqrt(2) ~= 1.41, past the radius.
			Expect(isec2d.CheckCircleEdge(newCircle(1.0, -1.0, 1.0), edge)).To(BeFalse())
		})

		It("returns false for a degenerate zero-length edge", func() {
			degenerate := shape2d.NewEdge(dprec.NewVec2(1.0, 2.0), dprec.NewVec2(1.0, 2.0))
			Expect(isec2d.CheckCircleEdge(newCircle(1.5, 2.0, 1.0), degenerate)).To(BeFalse())
		})
	})

	Describe("ResolveCircleEdge", func() {
		resolve := func(circle shape2d.Circle) (shape2d.Contact, bool) {
			var sink shape2d.LastContact
			isec2d.ResolveCircleEdge(circle, edge, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact against the edge span", func() {
			contact, ok := resolve(newCircle(0.5, 2.0, 1.0))
			Expect(ok).To(BeTrue())
			// Closest point is the perpendicular foot on the edge.
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 2.0))
			// Normal is the edge's outward normal.
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(1.0, 0.0))
			// Depth is radius minus the gap to the line: 1.0 - 0.5.
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("yields a contact against endpoint A", func() {
			contact, ok := resolve(newCircle(0.5, -0.5, 1.0))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(
				1.0/dprec.Sqrt(2.0), -1.0/dprec.Sqrt(2.0),
			))
			Expect(contact.Depth).To(BeNumerically("~", 1.0-dprec.Sqrt(0.5), 1e-6))
		})

		It("yields a contact against endpoint B", func() {
			contact, ok := resolve(newCircle(0.5, 4.5, 1.0))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 4.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(
				1.0/dprec.Sqrt(2.0), 1.0/dprec.Sqrt(2.0),
			))
			Expect(contact.Depth).To(BeNumerically("~", 1.0-dprec.Sqrt(0.5), 1e-6))
		})

		It("yields a zero-depth contact for a circle that just touches", func() {
			contact, ok := resolve(newCircle(1.0, 2.0, 1.0))
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("does not yield a contact for a circle centered behind the edge", func() {
			_, ok := resolve(newCircle(-1.0, 2.0, 1.0))
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact for a circle too far in front", func() {
			_, ok := resolve(newCircle(2.0, 2.0, 1.0))
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact for a circle short of an endpoint", func() {
			_, ok := resolve(newCircle(1.0, -1.0, 1.0))
			Expect(ok).To(BeFalse())
		})

		It("reports a unit normal for every contact kind", func() {
			for _, circle := range []shape2d.Circle{
				newCircle(0.5, 2.0, 1.0),  // edge span
				newCircle(0.5, -0.5, 1.0), // endpoint A
				newCircle(0.5, 4.5, 1.0),  // endpoint B
			} {
				contact, ok := resolve(circle)
				Expect(ok).To(BeTrue())
				Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
			}
		})

		It("removes the overlap when the circle is moved by Depth along the normal", func() {
			for _, circle := range []shape2d.Circle{
				newCircle(0.5, 2.0, 1.0),  // edge span
				newCircle(0.5, -0.5, 1.0), // endpoint A
				newCircle(0.5, 4.5, 1.0),  // endpoint B
			} {
				contact, ok := resolve(circle)
				Expect(ok).To(BeTrue())

				moved := shape2d.NewCircle(
					dprec.Vec2Sum(circle.Center, dprec.Vec2Prod(contact.TargetNormal, contact.Depth)),
					circle.Radius,
				)
				var sink shape2d.LastContact
				isec2d.ResolveCircleEdge(moved, edge, sink.AddContact)
				if resolved, ok := sink.Contact(); ok {
					Expect(resolved.Depth).To(BeNumerically("~", 0.0, 1e-6))
				}
			}
		})
	})
})
