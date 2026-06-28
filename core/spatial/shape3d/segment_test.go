package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("Segment", func() {
	var seg shape3d.Segment

	BeforeEach(func() {
		seg = shape3d.NewSegment(
			dprec.NewVec3(0.0, 0.0, 0.0),
			dprec.NewVec3(2.0, 3.0, 6.0),
		)
	})

	Describe("TransformedSegment", func() {
		It("applies the transform to both endpoints", func() {
			transform := shape3d.TRTransform(
				dprec.NewVec3(10.0, 20.0, 30.0),
				shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3())),
			)
			result := shape3d.TransformedSegment(seg, transform)
			// A (0,0,0) rotated by 90deg around Z stays (0,0,0), then translated to (10,20,30).
			Expect(result.A).To(dprectest.HaveVec3Coords(10.0, 20.0, 30.0))
			// B (2,3,6) rotated by 90deg around Z becomes (-3,2,6), then translated to (7,22,36).
			Expect(result.B).To(dprectest.HaveVec3Coords(7.0, 22.0, 36.0))
		})

		It("leaves the segment unchanged for the identity transform", func() {
			result := shape3d.TransformedSegment(seg, shape3d.IdentityTransform())
			Expect(result.A).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
			Expect(result.B).To(dprectest.HaveVec3Coords(2.0, 3.0, 6.0))
		})

		It("does not modify the original segment", func() {
			shape3d.TransformedSegment(seg, shape3d.TranslationTransform(dprec.NewVec3(5.0, 5.0, 5.0)))
			Expect(seg.A).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
			Expect(seg.B).To(dprectest.HaveVec3Coords(2.0, 3.0, 6.0))
		})
	})

	Describe("Length", func() {
		It("returns the Euclidean distance between A and B", func() {
			Expect(seg.Length()).To(BeNumerically("~", 7.0, 1e-6))
		})

		It("returns zero for a zero-length segment", func() {
			dot := shape3d.NewSegment(
				dprec.NewVec3(1.0, 2.0, 3.0),
				dprec.NewVec3(1.0, 2.0, 3.0),
			)
			Expect(dot.Length()).To(BeNumerically("~", 0.0, 1e-6))
		})
	})

	Describe("Midpoint", func() {
		It("returns the midpoint of the segment", func() {
			Expect(seg.Midpoint()).To(dprectest.HaveVec3Coords(1.0, 1.5, 3.0))
		})
	})

	Describe("Flipped", func() {
		It("swaps the start and end points", func() {
			flipped := seg.Flipped()
			Expect(flipped.A).To(dprectest.HaveVec3Coords(2.0, 3.0, 6.0))
			Expect(flipped.B).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
		})

		It("does not modify the original segment", func() {
			seg.Flipped()
			Expect(seg.A).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
			Expect(seg.B).To(dprectest.HaveVec3Coords(2.0, 3.0, 6.0))
		})
	})

	Describe("BoundingSphere", func() {
		It("is centered at the midpoint of the segment", func() {
			bs := seg.BoundingSphere()
			Expect(bs.Center).To(dprectest.HaveVec3Coords(1.0, 1.5, 3.0))
		})

		It("has radius equal to half the segment length", func() {
			bs := seg.BoundingSphere()
			Expect(bs.Radius).To(BeNumerically("~", 3.5, 1e-6))
		})

		It("contains both endpoints", func() {
			bs := seg.BoundingSphere()
			Expect(bs.ContainsPoint(seg.A)).To(BeTrue())
			Expect(bs.ContainsPoint(seg.B)).To(BeTrue())
		})
	})
})
