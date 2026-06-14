package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("Triangle", func() {
	var (
		ccwTriangle shape2d.Triangle
		cwTriangle  shape2d.Triangle
	)

	BeforeEach(func() {
		ccwTriangle = shape2d.NewTriangle(
			dprec.NewVec2(0.0, 0.0),
			dprec.NewVec2(1.0, 0.0),
			dprec.NewVec2(0.0, 1.0),
		)
		cwTriangle = shape2d.NewTriangle(
			dprec.NewVec2(0.0, 0.0),
			dprec.NewVec2(0.0, 1.0),
			dprec.NewVec2(1.0, 0.0),
		)
	})

	Describe("SignedArea", func() {
		It("returns positive area for a CCW triangle", func() {
			Expect(ccwTriangle.SignedArea()).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("returns negative area for a CW triangle", func() {
			Expect(cwTriangle.SignedArea()).To(BeNumerically("~", -0.5, 1e-6))
		})

		It("returns zero for collinear vertices", func() {
			degenerate := shape2d.NewTriangle(
				dprec.NewVec2(0.0, 0.0),
				dprec.NewVec2(1.0, 0.0),
				dprec.NewVec2(2.0, 0.0),
			)
			Expect(degenerate.SignedArea()).To(BeNumerically("~", 0.0, 1e-6))
		})
	})

	Describe("IsCCW", func() {
		It("returns true for CCW-ordered vertices", func() {
			Expect(ccwTriangle.IsCCW()).To(BeTrue())
		})

		It("returns false for CW-ordered vertices", func() {
			Expect(cwTriangle.IsCCW()).To(BeFalse())
		})

		It("returns false for collinear (zero-area) vertices", func() {
			degenerate := shape2d.NewTriangle(
				dprec.NewVec2(0.0, 0.0),
				dprec.NewVec2(1.0, 0.0),
				dprec.NewVec2(2.0, 0.0),
			)
			Expect(degenerate.IsCCW()).To(BeFalse())
		})
	})

	Describe("Centroid", func() {
		It("returns the average of the three vertices", func() {
			Expect(ccwTriangle.Centroid()).To(dprectest.HaveVec2Coords(
				1.0/3.0,
				1.0/3.0,
			))
		})
	})

	Describe("LengthAB", func() {
		It("returns the distance between A and B", func() {
			Expect(ccwTriangle.LengthAB()).To(BeNumerically("~", 1.0, 1e-6))
		})
	})

	Describe("LengthBC", func() {
		It("returns the distance between B and C", func() {
			Expect(ccwTriangle.LengthBC()).To(BeNumerically("~", dprec.Sqrt(2.0), 1e-6))
		})
	})

	Describe("LengthCA", func() {
		It("returns the distance between C and A", func() {
			Expect(ccwTriangle.LengthCA()).To(BeNumerically("~", 1.0, 1e-6))
		})
	})

	Describe("ContainsPoint", func() {
		It("returns true for a point strictly inside", func() {
			Expect(ccwTriangle.ContainsPoint(dprec.NewVec2(0.25, 0.25))).To(BeTrue())
		})

		It("returns false for a point strictly outside", func() {
			Expect(ccwTriangle.ContainsPoint(dprec.NewVec2(0.75, 0.75))).To(BeFalse())
		})

		It("returns true for vertex A", func() {
			Expect(ccwTriangle.ContainsPoint(dprec.NewVec2(0.0, 0.0))).To(BeTrue())
		})

		It("returns true for vertex B", func() {
			Expect(ccwTriangle.ContainsPoint(dprec.NewVec2(1.0, 0.0))).To(BeTrue())
		})

		It("returns true for vertex C", func() {
			Expect(ccwTriangle.ContainsPoint(dprec.NewVec2(0.0, 1.0))).To(BeTrue())
		})

		It("returns true for a point on edge AB", func() {
			Expect(ccwTriangle.ContainsPoint(dprec.NewVec2(0.5, 0.0))).To(BeTrue())
		})

		It("returns true for a point on edge BC", func() {
			Expect(ccwTriangle.ContainsPoint(dprec.NewVec2(0.5, 0.5))).To(BeTrue())
		})

		It("returns true for a point on edge CA", func() {
			Expect(ccwTriangle.ContainsPoint(dprec.NewVec2(0.0, 0.5))).To(BeTrue())
		})

		It("returns false for all points when the triangle is CW", func() {
			cw := shape2d.NewTriangle(
				dprec.NewVec2(0.0, 0.0),
				dprec.NewVec2(0.0, 1.0),
				dprec.NewVec2(1.0, 0.0),
			)
			Expect(cw.ContainsPoint(dprec.NewVec2(0.25, 0.25))).To(BeFalse())
			Expect(cw.ContainsPoint(dprec.NewVec2(0.5, 0.0))).To(BeFalse())
		})
	})

	Describe("TransformedTriangle", func() {
		It("applies transformation", func() {
			transform := shape2d.TRTransform(dprec.NewVec2(2.0, 3.0), dprec.Degrees(90.0))
			result := shape2d.TransformedTriangle(ccwTriangle, transform)
			Expect(result.A).To(dprectest.HaveVec2Coords(2.0+0.0, 3.0+0.0))
			Expect(result.B).To(dprectest.HaveVec2Coords(2.0+0.0, 3.0+1.0))
			Expect(result.C).To(dprectest.HaveVec2Coords(2.0-1.0, 3.0+0.0))
		})
	})
})
