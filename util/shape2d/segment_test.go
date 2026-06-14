package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("Segment", func() {
	var seg shape2d.Segment

	BeforeEach(func() {
		seg = shape2d.NewSegment(
			dprec.NewVec2(0.0, 0.0),
			dprec.NewVec2(3.0, 4.0),
		)
	})

	Describe("Length", func() {
		It("returns the Euclidean distance between A and B", func() {
			Expect(seg.Length()).To(BeNumerically("~", 5.0, 1e-6))
		})

		It("returns zero for a zero-length segment", func() {
			dot := shape2d.NewSegment(dprec.NewVec2(1.0, 2.0), dprec.NewVec2(1.0, 2.0))
			Expect(dot.Length()).To(BeNumerically("~", 0.0, 1e-6))
		})
	})

	Describe("Center", func() {
		It("returns the midpoint of the segment", func() {
			Expect(seg.Center()).To(dprectest.HaveVec2Coords(1.5, 2.0))
		})
	})

	Describe("BoundingCircle", func() {
		It("is centered at the midpoint of the segment", func() {
			bc := seg.BoundingCircle()
			Expect(bc.Position).To(dprectest.HaveVec2Coords(1.5, 2.0))
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

	Describe("TransformedSegment", func() {
		It("applies translation", func() {
			transform := shape2d.TRTransform(dprec.NewVec2(1.0, 2.0), dprec.Radians(0.0))
			result := shape2d.TransformedSegment(seg, transform)
			Expect(result.A).To(dprectest.HaveVec2Coords(1.0, 2.0))
			Expect(result.B).To(dprectest.HaveVec2Coords(4.0, 6.0))
		})

		It("applies 90° CCW rotation", func() {
			s := shape2d.NewSegment(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(1.0, 0.0))
			transform := shape2d.TRTransform(dprec.NewVec2(0.0, 0.0), dprec.Degrees(90.0))
			result := shape2d.TransformedSegment(s, transform)
			Expect(result.A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(result.B).To(dprectest.HaveVec2Coords(0.0, 1.0))
		})

		It("preserves segment length after transformation", func() {
			transform := shape2d.TRTransform(dprec.NewVec2(5.0, -3.0), dprec.Degrees(37.0))
			result := shape2d.TransformedSegment(seg, transform)
			Expect(result.Length()).To(BeNumerically("~", seg.Length(), 1e-6))
		})
	})
})
