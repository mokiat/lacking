package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/testing/sprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Square", func() {
	var square shape2d.Square

	BeforeEach(func() {
		square = shape2d.Square{
			Center: sprec.NewVec2(3.0, 4.0),
			Size:   2.0,
		}
	})

	Describe("ContainsPoint", func() {
		It("returns true for the center", func() {
			Expect(square.ContainsPoint(sprec.NewVec2(3.0, 4.0))).To(BeTrue())
		})

		It("returns true for a point strictly inside", func() {
			Expect(square.ContainsPoint(sprec.NewVec2(3.5, 4.5))).To(BeTrue())
		})

		It("returns true for a point on an edge", func() {
			Expect(square.ContainsPoint(sprec.NewVec2(4.0, 4.0))).To(BeTrue())
		})

		It("returns true for a corner", func() {
			Expect(square.ContainsPoint(sprec.NewVec2(4.0, 5.0))).To(BeTrue())
		})

		It("returns false for a point outside in X", func() {
			Expect(square.ContainsPoint(sprec.NewVec2(4.1, 4.0))).To(BeFalse())
		})

		It("returns false for a point outside in Y", func() {
			Expect(square.ContainsPoint(sprec.NewVec2(3.0, 5.1))).To(BeFalse())
		})

		It("returns true only for the center when size is zero", func() {
			dot := shape2d.Square{
				Center: sprec.NewVec2(1.0, 2.0),
				Size:   0.0,
			}
			Expect(dot.ContainsPoint(sprec.NewVec2(1.0, 2.0))).To(BeTrue())
			Expect(dot.ContainsPoint(sprec.NewVec2(1.1, 2.0))).To(BeFalse())
		})
	})

	Describe("BoundingCircle", func() {
		It("is centered at the center of the square", func() {
			bc := square.BoundingCircle()
			Expect(bc.Center).To(sprectest.HaveVec2Coords(3.0, 4.0))
		})

		It("has radius equal to half the diagonal", func() {
			bc := square.BoundingCircle()
			Expect(bc.Radius).To(BeNumerically("~", sprec.Sqrt(2.0), 1e-6))
		})

		It("contains the area just inside the corners", func() {
			bc := square.BoundingCircle()
			Expect(bc.ContainsPoint(sprec.NewVec2(2.01, 3.01))).To(BeTrue())
			Expect(bc.ContainsPoint(sprec.NewVec2(3.99, 3.01))).To(BeTrue())
			Expect(bc.ContainsPoint(sprec.NewVec2(2.01, 4.99))).To(BeTrue())
			Expect(bc.ContainsPoint(sprec.NewVec2(3.99, 4.99))).To(BeTrue())
		})
	})
})
