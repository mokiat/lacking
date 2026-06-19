package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/testing/sprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Segment", func() {
	var seg shape2d.Segment

	BeforeEach(func() {
		seg = shape2d.Segment{
			A: sprec.NewVec2(0.0, 0.0),
			B: sprec.NewVec2(3.0, 4.0),
		}
	})

	Describe("Length", func() {
		It("returns the Euclidean distance between A and B", func() {
			Expect(seg.Length()).To(BeNumerically("~", 5.0, 1e-6))
		})

		It("returns zero for a zero-length segment", func() {
			dot := shape2d.Segment{
				A: sprec.NewVec2(1.0, 2.0),
				B: sprec.NewVec2(1.0, 2.0),
			}
			Expect(dot.Length()).To(BeNumerically("~", 0.0, 1e-6))
		})
	})

	Describe("Midpoint", func() {
		It("returns the midpoint of the segment", func() {
			Expect(seg.Midpoint()).To(sprectest.HaveVec2Coords(1.5, 2.0))
		})
	})

	Describe("BoundingCircle", func() {
		It("is centered at the midpoint of the segment", func() {
			bc := seg.BoundingCircle()
			Expect(bc.Center).To(sprectest.HaveVec2Coords(1.5, 2.0))
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
