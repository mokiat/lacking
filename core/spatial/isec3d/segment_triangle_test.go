package isec3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("SegmentTriangle", func() {
	// A triangle lying in the z=0 plane, wound counter-clockwise when viewed
	// from +Z so that its normal points along +Z. It occupies the region
	// x>=0, y>=0, x+y<=2.
	var triangle shape3d.Triangle

	newSegment := func(ax, ay, az, bx, by, bz float64) shape3d.Segment {
		return shape3d.Segment{
			A: dprec.NewVec3(ax, ay, az),
			B: dprec.NewVec3(bx, by, bz),
		}
	}

	BeforeEach(func() {
		triangle = shape3d.Triangle{
			A: dprec.NewVec3(0.0, 0.0, 0.0),
			B: dprec.NewVec3(2.0, 0.0, 0.0),
			C: dprec.NewVec3(0.0, 2.0, 0.0),
		}
	})

	Describe("CheckSegmentTriangle", func() {
		It("returns true for a segment crossing the front face", func() {
			segment := newSegment(0.5, 0.5, 1.0, 0.5, 0.5, -1.0)
			Expect(isec3d.CheckSegmentTriangle(segment, triangle)).To(BeTrue())
		})

		It("returns true for a segment that crosses at an angle", func() {
			// Crosses the z=0 plane at (0.5,0.5,0), which is inside the triangle.
			segment := newSegment(0.0, 0.0, 1.0, 1.0, 1.0, -1.0)
			Expect(isec3d.CheckSegmentTriangle(segment, triangle)).To(BeTrue())
		})

		It("returns false for a segment approaching the back face", func() {
			// Same crossing point but travelling from -Z to +Z, so it is culled.
			segment := newSegment(0.5, 0.5, -1.0, 0.5, 0.5, 1.0)
			Expect(isec3d.CheckSegmentTriangle(segment, triangle)).To(BeFalse())
		})

		It("returns false for a segment crossing the plane outside the triangle", func() {
			// Crosses at (1.5,1.5,0), where x+y=3 exceeds the triangle's extent.
			segment := newSegment(1.5, 1.5, 1.0, 1.5, 1.5, -1.0)
			Expect(isec3d.CheckSegmentTriangle(segment, triangle)).To(BeFalse())
		})

		It("returns false for a segment that stops short of the plane", func() {
			// Both endpoints are on the +Z side, so the plane is never reached.
			segment := newSegment(0.5, 0.5, 1.0, 0.5, 0.5, 0.5)
			Expect(isec3d.CheckSegmentTriangle(segment, triangle)).To(BeFalse())
		})

		It("returns true when the far endpoint lies exactly on the triangle", func() {
			segment := newSegment(0.5, 0.5, 1.0, 0.5, 0.5, 0.0)
			Expect(isec3d.CheckSegmentTriangle(segment, triangle)).To(BeTrue())
		})

		It("returns false for a segment parallel to the triangle's plane", func() {
			segment := newSegment(0.5, 0.5, 1.0, 1.0, 0.5, 1.0)
			Expect(isec3d.CheckSegmentTriangle(segment, triangle)).To(BeFalse())
		})

		It("respects the triangle's winding order", func() {
			// Reversing the winding flips the normal, turning the previously
			// front-facing crossing into a culled back-facing one.
			flipped := shape3d.Triangle{A: triangle.A, B: triangle.C, C: triangle.B}
			segment := newSegment(0.5, 0.5, 1.0, 0.5, 0.5, -1.0)
			Expect(isec3d.CheckSegmentTriangle(segment, triangle)).To(BeTrue())
			Expect(isec3d.CheckSegmentTriangle(segment, flipped)).To(BeFalse())
		})
	})

	Describe("ResolveSegmentTriangle", func() {
		resolve := func(segment shape3d.Segment, t shape3d.Triangle) (shape3d.Contact, bool) {
			var sink shape3d.LastContact
			isec3d.ResolveSegmentTriangle(segment, t, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact at the crossing point", func() {
			contact, ok := resolve(newSegment(0.5, 0.5, 1.0, 0.5, 0.5, -1.0), triangle)
			Expect(ok).To(BeTrue())
			// Contact lies where the segment crosses the triangle's plane.
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.5, 0.5, 0.0))
			// Normal is the triangle's front-facing normal.
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
			// Depth is how far endpoint B has travelled past the plane: 0 - (-1).
			Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("computes the crossing point for an angled segment", func() {
			contact, ok := resolve(newSegment(0.0, 0.0, 1.0, 1.0, 1.0, -1.0), triangle)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.5, 0.5, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("scales the depth with how far B reaches past the plane", func() {
			contact, ok := resolve(newSegment(0.3, 0.4, 2.0, 0.3, 0.4, -3.0), triangle)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(0.3, 0.4, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 3.0, 1e-6))
		})

		It("reports a unit, front-facing normal", func() {
			contact, ok := resolve(newSegment(0.5, 0.5, 1.0, 0.5, 0.5, -1.0), triangle)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(triangle.Normal().X, triangle.Normal().Y, triangle.Normal().Z))
		})

		It("yields a zero-depth contact when B lies on the triangle", func() {
			contact, ok := resolve(newSegment(0.5, 0.5, 1.0, 0.5, 0.5, 0.0), triangle)
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("does not yield a contact for a back-facing crossing", func() {
			_, ok := resolve(newSegment(0.5, 0.5, -1.0, 0.5, 0.5, 1.0), triangle)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact when the crossing is outside the triangle", func() {
			_, ok := resolve(newSegment(1.5, 1.5, 1.0, 1.5, 1.5, -1.0), triangle)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact when the segment stops short", func() {
			_, ok := resolve(newSegment(0.5, 0.5, 1.0, 0.5, 0.5, 0.5), triangle)
			Expect(ok).To(BeFalse())
		})

		It("brings B onto the surface when it is moved out by Depth along the normal", func() {
			segment := newSegment(0.5, 0.5, 1.0, 0.5, 0.5, -1.0)
			contact, ok := resolve(segment, triangle)
			Expect(ok).To(BeTrue())

			movedB := dprec.Vec3Sum(segment.B, dprec.Vec3Prod(contact.TargetNormal, contact.Depth))
			// After the move B sits on the triangle's plane (z=0).
			Expect(movedB.Z).To(BeNumerically("~", 0.0, 1e-6))
		})
	})
})
