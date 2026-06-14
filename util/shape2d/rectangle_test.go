package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("Rectangle", func() {
	var rect shape2d.Rectangle

	BeforeEach(func() {
		rect = shape2d.NewRectangle(
			dprec.ZeroVec2(),
			dprec.Radians(0.0),
			dprec.NewVec2(4.0, 2.0),
		)
	})

	Describe("NewRectangle", func() {
		It("stores position and rotation and converts size to half-sizes", func() {
			r := shape2d.NewRectangle(
				dprec.NewVec2(1.0, 2.0),
				dprec.Degrees(45.0),
				dprec.NewVec2(6.0, 4.0),
			)
			Expect(r.Position).To(dprectest.HaveVec2Coords(1.0, 2.0))
			Expect(float64(r.Rotation)).To(BeNumerically("~", float64(dprec.Degrees(45.0)), 1e-6))
			Expect(r.HalfWidth).To(BeNumerically("~", 3.0, 1e-6))
			Expect(r.HalfHeight).To(BeNumerically("~", 2.0, 1e-6))
		})
	})

	Describe("BoundingCircle", func() {
		It("is centred at the rectangle position", func() {
			r := shape2d.NewRectangle(dprec.NewVec2(3.0, 4.0), dprec.Radians(0.0), dprec.NewVec2(4.0, 2.0))
			Expect(r.BoundingCircle().Position).To(dprectest.HaveVec2Coords(3.0, 4.0))
		})

		It("has radius equal to the half-diagonal", func() {
			r := shape2d.NewRectangle(dprec.ZeroVec2(), dprec.Radians(0.0), dprec.NewVec2(6.0, 8.0))
			Expect(r.BoundingCircle().Radius).To(BeNumerically("~", 5.0, 1e-6))
		})

		It("contains all four vertices", func() {
			bc := rect.BoundingCircle()
			for _, v := range rect.Vertices() {
				Expect(bc.ContainsPoint(v)).To(BeTrue())
			}
		})
	})

	Describe("Vertices", func() {
		It("returns the four corners in CCW order for an axis-aligned rectangle", func() {
			v := rect.Vertices()
			Expect(v[0]).To(dprectest.HaveVec2Coords(-2.0, 1.0))  // top-left
			Expect(v[1]).To(dprectest.HaveVec2Coords(-2.0, -1.0)) // bottom-left
			Expect(v[2]).To(dprectest.HaveVec2Coords(2.0, -1.0))  // bottom-right
			Expect(v[3]).To(dprectest.HaveVec2Coords(2.0, 1.0))   // top-right
		})

		It("applies 90° CCW rotation to the vertices", func() {
			r := shape2d.NewRectangle(dprec.ZeroVec2(), dprec.Degrees(90.0), dprec.NewVec2(4.0, 2.0))
			v := r.Vertices()
			Expect(v[0]).To(dprectest.HaveVec2Coords(-1.0, -2.0))
			Expect(v[1]).To(dprectest.HaveVec2Coords(1.0, -2.0))
			Expect(v[2]).To(dprectest.HaveVec2Coords(1.0, 2.0))
			Expect(v[3]).To(dprectest.HaveVec2Coords(-1.0, 2.0))
		})

		It("offsets all vertices by the rectangle position", func() {
			r := shape2d.NewRectangle(dprec.NewVec2(1.0, 2.0), dprec.Radians(0.0), dprec.NewVec2(4.0, 2.0))
			v := r.Vertices()
			Expect(v[0]).To(dprectest.HaveVec2Coords(-1.0, 3.0))
			Expect(v[1]).To(dprectest.HaveVec2Coords(-1.0, 1.0))
			Expect(v[2]).To(dprectest.HaveVec2Coords(3.0, 1.0))
			Expect(v[3]).To(dprectest.HaveVec2Coords(3.0, 3.0))
		})
	})

	Describe("TransformedRectangle", func() {
		It("translates the position", func() {
			r := shape2d.NewRectangle(dprec.NewVec2(1.0, 2.0), dprec.Radians(0.0), dprec.NewVec2(4.0, 2.0))
			transform := shape2d.TRTransform(dprec.NewVec2(3.0, 0.0), dprec.Radians(0.0))
			result := shape2d.TransformedRectangle(r, transform)
			Expect(result.Position).To(dprectest.HaveVec2Coords(4.0, 2.0))
			Expect(float64(result.Rotation)).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("rotates the position and adds the rotation", func() {
			r := shape2d.NewRectangle(dprec.NewVec2(1.0, 0.0), dprec.Radians(0.0), dprec.NewVec2(4.0, 2.0))
			transform := shape2d.TRTransform(dprec.ZeroVec2(), dprec.Degrees(90.0))
			result := shape2d.TransformedRectangle(r, transform)
			Expect(result.Position).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(float64(result.Rotation)).To(BeNumerically("~", float64(dprec.Degrees(90.0)), 1e-6))
		})

		It("preserves half-sizes", func() {
			transform := shape2d.TRTransform(dprec.NewVec2(5.0, -3.0), dprec.Degrees(37.0))
			result := shape2d.TransformedRectangle(rect, transform)
			Expect(result.HalfWidth).To(BeNumerically("~", rect.HalfWidth, 1e-6))
			Expect(result.HalfHeight).To(BeNumerically("~", rect.HalfHeight, 1e-6))
		})
	})
})
