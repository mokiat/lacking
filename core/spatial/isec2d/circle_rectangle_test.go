package isec2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("CircleRectangle", func() {
	// An axis-aligned rectangle at the origin with distinct half-extents along
	// each axis, so that any axis confusion is caught. It spans x:[-2,2] and
	// y:[-1,1].
	var rectangle shape2d.Rectangle

	newCircle := func(x, y, radius float64) shape2d.Circle {
		return shape2d.NewCircle(dprec.NewVec2(x, y), radius)
	}

	BeforeEach(func() {
		rectangle = shape2d.NewRectangle(
			dprec.NewVec2(0.0, 0.0),
			shape2d.IdentityRotation(),
			dprec.NewVec2(2.0, 1.0),
		)
	})

	Describe("CheckCircleRectangle", func() {
		It("returns true for a circle overlapping a face", func() {
			Expect(isec2d.CheckCircleRectangle(newCircle(3.0, 0.0, 1.5), rectangle)).To(BeTrue())
		})

		It("returns true for a circle that just touches a face", func() {
			// Right face is at x=2; center at x=3.5 with radius 1.5 just reaches it.
			Expect(isec2d.CheckCircleRectangle(newCircle(3.5, 0.0, 1.5), rectangle)).To(BeTrue())
		})

		It("returns false for a circle just short of a face", func() {
			Expect(isec2d.CheckCircleRectangle(newCircle(3.6, 0.0, 1.5), rectangle)).To(BeFalse())
		})

		It("returns true for a circle whose center is inside the rectangle", func() {
			Expect(isec2d.CheckCircleRectangle(newCircle(0.0, 0.0, 0.5), rectangle)).To(BeTrue())
		})

		It("returns true for a circle that fully contains the rectangle", func() {
			Expect(isec2d.CheckCircleRectangle(newCircle(0.0, 0.0, 10.0), rectangle)).To(BeTrue())
		})

		It("returns true for a circle reaching a corner", func() {
			// Past the right and top faces; nearest feature is the (2,1) corner at
			// distance sqrt(0.5) ~= 0.707.
			Expect(isec2d.CheckCircleRectangle(newCircle(2.5, 1.5, 1.0), rectangle)).To(BeTrue())
		})

		It("returns false when past two faces but short of the corner", func() {
			// The corner distance sqrt(0.5) ~= 0.707 is beyond radius 0.6, even
			// though the center is within radius of each individual face.
			Expect(isec2d.CheckCircleRectangle(newCircle(2.5, 1.5, 0.6), rectangle)).To(BeFalse())
		})

		It("respects the rectangle orientation", func() {
			// Rotate the rectangle 90 degrees, so its local X (half-width 2) now
			// points along world Y, and its local Y (half-height 1) along world X.
			// The rectangle then only spans x:[-1,1].
			rotated := rectangle
			rotated.Rotation = shape2d.RotationFromAngle(dprec.Degrees(90.0))
			// A point that would be inside the unrotated rectangle is now outside.
			Expect(isec2d.CheckCircleRectangle(newCircle(1.6, 0.0, 0.5), rectangle)).To(BeTrue())
			Expect(isec2d.CheckCircleRectangle(newCircle(1.6, 0.0, 0.5), rotated)).To(BeFalse())
		})
	})

	Describe("ResolveCircleRectangle", func() {
		resolve := func(circle shape2d.Circle, r shape2d.Rectangle) (shape2d.Contact, bool) {
			var sink shape2d.LastContact
			isec2d.ResolveCircleRectangle(circle, r, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact against a face", func() {
			contact, ok := resolve(newCircle(3.0, 0.0, 1.5), rectangle)
			Expect(ok).To(BeTrue())
			// Normal points from the rectangle (target) toward the circle (source).
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(1.0, 0.0))
			// Contact lies on the right face.
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(2.0, 0.0))
			// Depth is radius minus the gap to the face: 1.5 - (3 - 2) = 0.5.
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("resolves contacts against each face with the correct outward normal", func() {
			right, _ := resolve(newCircle(3.0, 0.0, 1.5), rectangle)
			Expect(right.TargetNormal).To(dprectest.HaveVec2Coords(1.0, 0.0))
			Expect(right.TargetPoint).To(dprectest.HaveVec2Coords(2.0, 0.0))

			left, _ := resolve(newCircle(-3.0, 0.0, 1.5), rectangle)
			Expect(left.TargetNormal).To(dprectest.HaveVec2Coords(-1.0, 0.0))
			Expect(left.TargetPoint).To(dprectest.HaveVec2Coords(-2.0, 0.0))

			top, _ := resolve(newCircle(0.0, 2.0, 1.5), rectangle)
			Expect(top.TargetNormal).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(top.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 1.0))

			bottom, _ := resolve(newCircle(0.0, -2.0, 1.5), rectangle)
			Expect(bottom.TargetNormal).To(dprectest.HaveVec2Coords(0.0, -1.0))
			Expect(bottom.TargetPoint).To(dprectest.HaveVec2Coords(0.0, -1.0))
		})

		It("yields a contact against a corner", func() {
			contact, ok := resolve(newCircle(3.0, 2.0, 1.5), rectangle)
			Expect(ok).To(BeTrue())
			// Contact lies on the (2,1) corner.
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(2.0, 1.0))
			// Normal points outward along the corner diagonal.
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(
				1.0/dprec.Sqrt(2.0), 1.0/dprec.Sqrt(2.0),
			))
			// Depth is radius minus the corner distance: 1.5 - sqrt(2).
			Expect(contact.Depth).To(BeNumerically("~", 1.5-dprec.Sqrt(2.0), 1e-6))
		})

		It("yields a contact along the least-penetration axis when inside the rectangle", func() {
			// Center is nearest to the right face (0.5 away), so resolution is
			// along +X.
			contact, ok := resolve(newCircle(1.5, 0.0, 0.5), rectangle)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(1.0, 0.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(2.0, 0.0))
			// Depth carries the center to the far side of the face: 0.5 + 0.5.
			Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("yields a zero-depth contact for a circle that just touches", func() {
			contact, ok := resolve(newCircle(3.5, 0.0, 1.5), rectangle)
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("does not yield a contact for a disjoint circle", func() {
			_, ok := resolve(newCircle(3.6, 0.0, 1.5), rectangle)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact past two faces but short of the corner", func() {
			_, ok := resolve(newCircle(2.5, 1.5, 0.6), rectangle)
			Expect(ok).To(BeFalse())
		})

		It("reports a unit normal for every contact kind", func() {
			for _, circle := range []shape2d.Circle{
				newCircle(3.0, 0.0, 1.5), // face
				newCircle(3.0, 2.0, 1.5), // corner
				newCircle(1.5, 0.0, 0.5), // inside
			} {
				contact, ok := resolve(circle, rectangle)
				Expect(ok).To(BeTrue())
				Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
			}
		})

		It("removes the overlap when the circle is moved by Depth along the normal", func() {
			for _, circle := range []shape2d.Circle{
				newCircle(3.0, 0.0, 1.5), // face
				newCircle(3.0, 2.0, 1.5), // corner
			} {
				contact, ok := resolve(circle, rectangle)
				Expect(ok).To(BeTrue())

				moved := shape2d.NewCircle(
					dprec.Vec2Sum(circle.Center, dprec.Vec2Prod(contact.TargetNormal, contact.Depth)),
					circle.Radius,
				)
				// After moving out by Depth the circle only just touches, so a
				// re-resolve reports (essentially) zero remaining penetration.
				resolved, ok := resolve(moved, rectangle)
				if ok {
					Expect(resolved.Depth).To(BeNumerically("~", 0.0, 1e-6))
				}
			}
		})

		It("reports the normal in world space for an oriented rectangle", func() {
			// Rotate the rectangle 90 degrees: its local X axis maps to world Y.
			rotated := rectangle
			rotated.Rotation = shape2d.RotationFromAngle(dprec.Degrees(90.0))
			// The circle sits beyond the rectangle's local right face, which now
			// faces world +Y (the rectangle spans y:[-2,2] after rotation).
			contact, ok := resolve(newCircle(0.0, 3.0, 1.5), rotated)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 2.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})
	})
})
