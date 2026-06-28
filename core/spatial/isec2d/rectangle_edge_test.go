package isec2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("RectangleEdge", func() {
	// An edge running from (0,0) up to (0,4). Its normal is the right-hand
	// perpendicular of the A-to-B direction, which points along +X, so the front
	// side of the edge is to the right (+X).
	var edge shape2d.Edge

	newRectangle := func(x, y, halfWidth, halfHeight float64) shape2d.Rectangle {
		return shape2d.NewRectangle(
			dprec.NewVec2(x, y),
			shape2d.IdentityRotation(),
			dprec.NewVec2(halfWidth, halfHeight),
		)
	}

	BeforeEach(func() {
		edge = shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(0.0, 4.0))
	})

	Describe("CheckRectangleEdge", func() {
		It("returns true for a rectangle overlapping the edge span", func() {
			Expect(isec2d.CheckRectangleEdge(newRectangle(0.5, 2.0, 1.0, 0.5), edge)).To(BeTrue())
		})

		It("returns true for a rectangle that just touches the edge", func() {
			Expect(isec2d.CheckRectangleEdge(newRectangle(1.0, 2.0, 1.0, 0.5), edge)).To(BeTrue())
		})

		It("returns false for a rectangle too far in front of the edge", func() {
			Expect(isec2d.CheckRectangleEdge(newRectangle(2.0, 2.0, 0.5, 0.5), edge)).To(BeFalse())
		})

		It("returns false for a rectangle centered behind the edge", func() {
			Expect(isec2d.CheckRectangleEdge(newRectangle(-1.0, 2.0, 1.0, 0.5), edge)).To(BeFalse())
		})

		It("returns false for a rectangle centered exactly on the edge's line", func() {
			Expect(isec2d.CheckRectangleEdge(newRectangle(0.0, 2.0, 1.0, 0.5), edge)).To(BeFalse())
		})

		It("returns true for a rectangle reaching past an endpoint", func() {
			Expect(isec2d.CheckRectangleEdge(newRectangle(0.3, 4.4, 0.5, 0.5), edge)).To(BeTrue())
		})

		It("returns false for a rectangle beyond an endpoint", func() {
			Expect(isec2d.CheckRectangleEdge(newRectangle(0.5, 5.5, 0.5, 0.5), edge)).To(BeFalse())
		})

		It("returns false for a degenerate zero-length edge", func() {
			degenerate := shape2d.NewEdge(dprec.NewVec2(1.0, 2.0), dprec.NewVec2(1.0, 2.0))
			Expect(isec2d.CheckRectangleEdge(newRectangle(1.0, 2.0, 1.0, 1.0), degenerate)).To(BeFalse())
		})
	})

	Describe("ResolveRectangleEdge", func() {
		resolve := func(rectangle shape2d.Rectangle) (shape2d.Contact, bool) {
			var sink shape2d.LastContact
			isec2d.ResolveRectangleEdge(rectangle, edge, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact against the edge span", func() {
			contact, ok := resolve(newRectangle(0.5, 2.0, 1.0, 0.5))
			Expect(ok).To(BeTrue())
			// Resolved along the edge's outward normal, at the foot below the
			// rectangle's deepest corner.
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 2.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(1.0, 0.0))
			// Depth is the half-width minus the gap to the line: 1.0 - 0.5.
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("yields a contact along a rectangle axis near an endpoint", func() {
			contact, ok := resolve(newRectangle(0.3, 4.4, 0.5, 0.5))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 4.0))
			// Least penetration is along the rectangle's Y axis, sliding it off the
			// endpoint rather than out along the edge normal.
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.1, 1e-6))
		})

		It("yields a contact for a rotated rectangle on the span", func() {
			rectangle := shape2d.NewRectangle(
				dprec.NewVec2(0.3, 2.0),
				shape2d.RotationFromAngle(dprec.Degrees(45.0)),
				dprec.NewVec2(0.5, 0.5),
			)
			var sink shape2d.LastContact
			isec2d.ResolveRectangleEdge(rectangle, edge, sink.AddContact)
			contact, ok := sink.Contact()
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 2.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(1.0, 0.0))
			// The half-diagonal extent along the normal is 0.5*sqrt(2); the center
			// sits 0.3 in front.
			Expect(contact.Depth).To(BeNumerically("~", 0.5*dprec.Sqrt(2.0)-0.3, 1e-6))
		})

		It("yields a unit normal", func() {
			contact, ok := resolve(newRectangle(0.5, 2.0, 1.0, 0.5))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("does not yield a contact for a rectangle centered behind the edge", func() {
			_, ok := resolve(newRectangle(-1.0, 2.0, 1.0, 0.5))
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact for a rectangle too far in front", func() {
			_, ok := resolve(newRectangle(2.0, 2.0, 0.5, 0.5))
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact for a rectangle beyond an endpoint", func() {
			_, ok := resolve(newRectangle(0.5, 5.5, 0.5, 0.5))
			Expect(ok).To(BeFalse())
		})

		It("removes the overlap when the rectangle is moved by Depth along the normal", func() {
			for _, rectangle := range []shape2d.Rectangle{
				newRectangle(0.5, 2.0, 1.0, 0.5), // span
				newRectangle(0.3, 4.4, 0.5, 0.5), // endpoint
			} {
				contact, ok := resolve(rectangle)
				Expect(ok).To(BeTrue())

				moved := shape2d.NewRectangle(
					dprec.Vec2Sum(rectangle.Center, dprec.Vec2Prod(contact.TargetNormal, contact.Depth)),
					rectangle.Rotation,
					dprec.NewVec2(rectangle.HalfWidth, rectangle.HalfHeight),
				)
				var sink shape2d.LastContact
				isec2d.ResolveRectangleEdge(moved, edge, sink.AddContact)
				if resolved, ok := sink.Contact(); ok {
					Expect(resolved.Depth).To(BeNumerically("~", 0.0, 1e-6))
				}
			}
		})
	})
})
