package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Segment", func() {
	var seg shape2d.Segment

	BeforeEach(func() {
		seg = shape2d.Segment{
			A: dprec.NewVec2(0.0, 0.0),
			B: dprec.NewVec2(3.0, 4.0),
		}
	})

	Describe("TransformedSegment", func() {
		It("applies the transform to both endpoints", func() {
			transform := shape2d.TRTransform(
				dprec.NewVec2(10.0, 20.0),
				shape2d.RotationFromAngle(dprec.Degrees(90.0)),
			)
			result := shape2d.TransformedSegment(seg, transform)
			// A (0,0) rotated by 90deg stays (0,0), then translated to (10,20).
			Expect(result.A).To(dprectest.HaveVec2Coords(10.0, 20.0))
			// B (3,4) rotated by 90deg becomes (-4,3), then translated to (6,23).
			Expect(result.B).To(dprectest.HaveVec2Coords(6.0, 23.0))
		})

		It("leaves the segment unchanged for the identity transform", func() {
			result := shape2d.TransformedSegment(seg, shape2d.IdentityTransform())
			Expect(result.A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(result.B).To(dprectest.HaveVec2Coords(3.0, 4.0))
		})

		It("does not modify the original segment", func() {
			shape2d.TransformedSegment(seg, shape2d.TranslationTransform(dprec.NewVec2(5.0, 5.0)))
			Expect(seg.A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(seg.B).To(dprectest.HaveVec2Coords(3.0, 4.0))
		})
	})

	Describe("Length", func() {
		It("returns the Euclidean distance between A and B", func() {
			Expect(seg.Length()).To(BeNumerically("~", 5.0, 1e-6))
		})

		It("returns zero for a zero-length segment", func() {
			dot := shape2d.Segment{
				A: dprec.NewVec2(1.0, 2.0),
				B: dprec.NewVec2(1.0, 2.0),
			}
			Expect(dot.Length()).To(BeNumerically("~", 0.0, 1e-6))
		})
	})

	Describe("Midpoint", func() {
		It("returns the midpoint of the segment", func() {
			Expect(seg.Midpoint()).To(dprectest.HaveVec2Coords(1.5, 2.0))
		})
	})

	Describe("Flipped", func() {
		It("swaps the start and end points", func() {
			flipped := seg.Flipped()
			Expect(flipped.A).To(dprectest.HaveVec2Coords(3.0, 4.0))
			Expect(flipped.B).To(dprectest.HaveVec2Coords(0.0, 0.0))
		})

		It("does not modify the original segment", func() {
			seg.Flipped()
			Expect(seg.A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(seg.B).To(dprectest.HaveVec2Coords(3.0, 4.0))
		})
	})

	Describe("BoundingCircle", func() {
		It("is centered at the midpoint of the segment", func() {
			bc := seg.BoundingCircle()
			Expect(bc.Center).To(dprectest.HaveVec2Coords(1.5, 2.0))
		})

		It("has radius equal to half the segment length", func() {
			bc := seg.BoundingCircle()
			Expect(bc.Radius).To(BeNumerically("~", 2.5, 1e-6))
		})

		It("contains both endpoints", func() {
			bc := seg.BoundingCircle()
			Expect(bc.ContainsPoint(seg.A)).To(BeTrue())
			Expect(bc.ContainsPoint(seg.B)).To(BeTrue())
		})
	})
})
