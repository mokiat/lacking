package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Rectangle", func() {
	var rectangle shape2d.Rectangle

	BeforeEach(func() {
		rectangle = shape2d.Rectangle{
			Center:   dprec.NewVec2(3.0, 4.0),
			Rotation: shape2d.IdentityRotation(),
			Width:    6.0,
			Height:   8.0,
		}
	})

	Describe("ContainsPoint", func() {
		It("returns true for the center", func() {
			Expect(rectangle.ContainsPoint(dprec.NewVec2(3.0, 4.0))).To(BeTrue())
		})

		It("returns true for a point strictly inside", func() {
			Expect(rectangle.ContainsPoint(dprec.NewVec2(4.0, 5.0))).To(BeTrue())
		})

		It("returns true for a point on an edge", func() {
			Expect(rectangle.ContainsPoint(dprec.NewVec2(6.0, 4.0))).To(BeTrue())
		})

		It("returns true for a corner", func() {
			Expect(rectangle.ContainsPoint(dprec.NewVec2(6.0, 8.0))).To(BeTrue())
		})

		It("returns false for a point outside in X", func() {
			Expect(rectangle.ContainsPoint(dprec.NewVec2(6.1, 4.0))).To(BeFalse())
		})

		It("returns false for a point outside in Y", func() {
			Expect(rectangle.ContainsPoint(dprec.NewVec2(3.0, 8.1))).To(BeFalse())
		})

		It("returns true only for the center when width and height are zero", func() {
			dot := shape2d.Rectangle{
				Center:   dprec.NewVec2(1.0, 2.0),
				Rotation: shape2d.IdentityRotation(),
				Width:    0.0,
				Height:   0.0,
			}
			Expect(dot.ContainsPoint(dprec.NewVec2(1.0, 2.0))).To(BeTrue())
			Expect(dot.ContainsPoint(dprec.NewVec2(1.1, 2.0))).To(BeFalse())
		})

		Context("with 90-degree CCW rotation", func() {
			var rotated shape2d.Rectangle

			BeforeEach(func() {
				rotated = shape2d.Rectangle{
					Center:   dprec.NewVec2(3.0, 4.0),
					Rotation: shape2d.RotationFromCosSin(0.0, 1.0),
					Width:    6.0,
					Height:   8.0,
				}
			})

			It("contains a point that lies outside the axis-aligned rectangle", func() {
				Expect(rotated.ContainsPoint(dprec.NewVec2(6.5, 4.0))).To(BeTrue())
			})

			It("rejects a point that lies inside the axis-aligned rectangle", func() {
				Expect(rotated.ContainsPoint(dprec.NewVec2(3.0, 0.5))).To(BeFalse())
			})

			It("contains a point on the width boundary in world Y", func() {
				Expect(rotated.ContainsPoint(dprec.NewVec2(3.0, 7.0))).To(BeTrue())
			})

			It("rejects a point just beyond the width boundary in world Y", func() {
				Expect(rotated.ContainsPoint(dprec.NewVec2(3.0, 7.1))).To(BeFalse())
			})

			It("contains a point on the height boundary in world X", func() {
				Expect(rotated.ContainsPoint(dprec.NewVec2(7.0, 4.0))).To(BeTrue())
			})

			It("rejects a point just beyond the height boundary in world X", func() {
				Expect(rotated.ContainsPoint(dprec.NewVec2(7.1, 4.0))).To(BeFalse())
			})
		})
	})

	Describe("BoundingCircle", func() {
		It("is centered at the center of the rectangle", func() {
			bc := rectangle.BoundingCircle()
			Expect(bc.Center).To(dprectest.HaveVec2Coords(3.0, 4.0))
		})

		It("has radius equal to half the diagonal", func() {
			bc := rectangle.BoundingCircle()
			Expect(bc.Radius).To(BeNumerically("~", 5.0, 1e-6))
		})

		It("contains the area just inside the corners", func() {
			bc := rectangle.BoundingCircle()
			Expect(bc.ContainsPoint(dprec.NewVec2(0.01, 0.01))).To(BeTrue())
			Expect(bc.ContainsPoint(dprec.NewVec2(5.99, 0.01))).To(BeTrue())
			Expect(bc.ContainsPoint(dprec.NewVec2(0.01, 7.99))).To(BeTrue())
			Expect(bc.ContainsPoint(dprec.NewVec2(5.99, 7.99))).To(BeTrue())
		})
	})
})
