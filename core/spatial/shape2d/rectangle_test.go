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
		rectangle = shape2d.NewRectangle(
			dprec.NewVec2(3.0, 4.0),
			shape2d.IdentityRotation(),
			dprec.NewVec2(3.0, 4.0),
		)
	})

	Describe("TransformedRectangle", func() {
		It("moves the center, composes the rotation and keeps the size", func() {
			transform := shape2d.TRTransform(
				dprec.NewVec2(10.0, 20.0),
				shape2d.RotationFromAngle(dprec.Degrees(90.0)),
			)
			result := shape2d.TransformedRectangle(rectangle, transform)
			// Center (3,4) rotated by 90deg becomes (-4,3), then translated to (6,23).
			Expect(result.Center).To(dprectest.HaveVec2Coords(6.0, 23.0))
			// Identity rotation composed with a 90deg rotation yields a 90deg rotation.
			Expect(result.Rotation.BasisX).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(result.Rotation.BasisY).To(dprectest.HaveVec2Coords(-1.0, 0.0))
			Expect(result.HalfWidth).To(BeNumerically("~", 3.0, 1e-6))
			Expect(result.HalfHeight).To(BeNumerically("~", 4.0, 1e-6))
		})

		It("composes with an already rotated rectangle", func() {
			rectangle.Rotation = shape2d.RotationFromAngle(dprec.Degrees(30.0))
			transform := shape2d.RotationTransform(shape2d.RotationFromAngle(dprec.Degrees(60.0)))
			result := shape2d.TransformedRectangle(rectangle, transform)
			Expect(result.Rotation.Angle().Degrees()).To(BeNumerically("~", 90.0, 1e-6))
		})

		It("leaves the rectangle unchanged for the identity transform", func() {
			result := shape2d.TransformedRectangle(rectangle, shape2d.IdentityTransform())
			Expect(result.Center).To(dprectest.HaveVec2Coords(3.0, 4.0))
			Expect(result.Rotation.BasisX).To(dprectest.HaveVec2Coords(1.0, 0.0))
			Expect(result.Rotation.BasisY).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(result.HalfWidth).To(BeNumerically("~", 3.0, 1e-6))
			Expect(result.HalfHeight).To(BeNumerically("~", 4.0, 1e-6))
		})

		It("does not modify the original rectangle", func() {
			shape2d.TransformedRectangle(rectangle, shape2d.TranslationTransform(dprec.NewVec2(5.0, 5.0)))
			Expect(rectangle.Center).To(dprectest.HaveVec2Coords(3.0, 4.0))
			Expect(rectangle.HalfWidth).To(BeNumerically("~", 3.0, 1e-6))
			Expect(rectangle.HalfHeight).To(BeNumerically("~", 4.0, 1e-6))
		})
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

		It("returns true only for the center when half-width and half-height are zero", func() {
			dot := shape2d.NewRectangle(
				dprec.NewVec2(1.0, 2.0),
				shape2d.IdentityRotation(),
				dprec.NewVec2(0.0, 0.0),
			)
			Expect(dot.ContainsPoint(dprec.NewVec2(1.0, 2.0))).To(BeTrue())
			Expect(dot.ContainsPoint(dprec.NewVec2(1.1, 2.0))).To(BeFalse())
		})

		Context("with 90-degree CCW rotation", func() {
			var rotated shape2d.Rectangle

			BeforeEach(func() {
				rotated = shape2d.NewRectangle(
					dprec.NewVec2(3.0, 4.0),
					shape2d.RotationFromCosSin(0.0, 1.0),
					dprec.NewVec2(3.0, 4.0),
				)
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
