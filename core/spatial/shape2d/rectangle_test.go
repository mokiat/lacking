package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/testing/sprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Rectangle", func() {
	var rectangle shape2d.Rectangle

	BeforeEach(func() {
		rectangle = shape2d.Rectangle{
			Center: sprec.NewVec2(3.0, 4.0),
			Width:  6.0,
			Height: 8.0,
		}
	})

	Describe("ContainsPoint", func() {
		It("returns true for the center", func() {
			Expect(rectangle.ContainsPoint(sprec.NewVec2(3.0, 4.0))).To(BeTrue())
		})

		It("returns true for a point strictly inside", func() {
			Expect(rectangle.ContainsPoint(sprec.NewVec2(4.0, 5.0))).To(BeTrue())
		})

		It("returns true for a point on an edge", func() {
			Expect(rectangle.ContainsPoint(sprec.NewVec2(6.0, 4.0))).To(BeTrue())
		})

		It("returns true for a corner", func() {
			Expect(rectangle.ContainsPoint(sprec.NewVec2(6.0, 8.0))).To(BeTrue())
		})

		It("returns false for a point outside in X", func() {
			Expect(rectangle.ContainsPoint(sprec.NewVec2(6.1, 4.0))).To(BeFalse())
		})

		It("returns false for a point outside in Y", func() {
			Expect(rectangle.ContainsPoint(sprec.NewVec2(3.0, 8.1))).To(BeFalse())
		})

		It("returns true only for the center when width and height are zero", func() {
			dot := shape2d.Rectangle{
				Center: sprec.NewVec2(1.0, 2.0),
				Width:  0.0,
				Height: 0.0,
			}
			Expect(dot.ContainsPoint(sprec.NewVec2(1.0, 2.0))).To(BeTrue())
			Expect(dot.ContainsPoint(sprec.NewVec2(1.1, 2.0))).To(BeFalse())
		})
	})

	Describe("BoundingCircle", func() {
		It("is centered at the center of the rectangle", func() {
			bc := rectangle.BoundingCircle()
			Expect(bc.Center).To(sprectest.HaveVec2Coords(3.0, 4.0))
		})

		It("has radius equal to half the diagonal", func() {
			bc := rectangle.BoundingCircle()
			Expect(bc.Radius).To(BeNumerically("~", 5.0, 1e-6))
		})

		It("contains the area just inside the corners", func() {
			bc := rectangle.BoundingCircle()
			Expect(bc.ContainsPoint(sprec.NewVec2(0.01, 0.01))).To(BeTrue())
			Expect(bc.ContainsPoint(sprec.NewVec2(5.99, 0.01))).To(BeTrue())
			Expect(bc.ContainsPoint(sprec.NewVec2(0.01, 7.99))).To(BeTrue())
			Expect(bc.ContainsPoint(sprec.NewVec2(5.99, 7.99))).To(BeTrue())
		})
	})
})
