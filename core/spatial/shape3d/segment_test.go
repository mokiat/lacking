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
		seg = shape3d.Segment{
			A: dprec.NewVec3(0.0, 0.0, 0.0),
			B: dprec.NewVec3(2.0, 3.0, 6.0),
		}
	})

	Describe("Length", func() {
		It("returns the Euclidean distance between A and B", func() {
			Expect(seg.Length()).To(BeNumerically("~", 7.0, 1e-6))
		})

		It("returns zero for a zero-length segment", func() {
			dot := shape3d.Segment{
				A: dprec.NewVec3(1.0, 2.0, 3.0),
				B: dprec.NewVec3(1.0, 2.0, 3.0),
			}
			Expect(dot.Length()).To(BeNumerically("~", 0.0, 1e-6))
		})
	})

	Describe("Midpoint", func() {
		It("returns the midpoint of the segment", func() {
			Expect(seg.Midpoint()).To(dprectest.HaveVec3Coords(1.0, 1.5, 3.0))
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
