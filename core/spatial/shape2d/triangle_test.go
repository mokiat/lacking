package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Triangle", func() {
	var triangle shape2d.Triangle

	BeforeEach(func() {
		// A 3-4-5 right triangle wound counter-clockwise.
		triangle = shape2d.NewTriangle(
			dprec.NewVec2(0.0, 0.0),
			dprec.NewVec2(4.0, 0.0),
			dprec.NewVec2(0.0, 3.0),
		)
	})

	Describe("NewTriangle", func() {
		It("sets the three vertices", func() {
			Expect(triangle.A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(triangle.B).To(dprectest.HaveVec2Coords(4.0, 0.0))
			Expect(triangle.C).To(dprectest.HaveVec2Coords(0.0, 3.0))
		})
	})

	Describe("TransformedTriangle", func() {
		It("applies the transform to all three vertices", func() {
			transform := shape2d.TRTransform(
				dprec.NewVec2(10.0, 20.0),
				shape2d.RotationFromAngle(dprec.Degrees(90.0)),
			)
			result := shape2d.TransformedTriangle(triangle, transform)
			// A (0,0) rotated by 90deg stays (0,0), then translated to (10,20).
			Expect(result.A).To(dprectest.HaveVec2Coords(10.0, 20.0))
			// B (4,0) rotated by 90deg becomes (0,4), then translated to (10,24).
			Expect(result.B).To(dprectest.HaveVec2Coords(10.0, 24.0))
			// C (0,3) rotated by 90deg becomes (-3,0), then translated to (7,20).
			Expect(result.C).To(dprectest.HaveVec2Coords(7.0, 20.0))
		})

		It("leaves the triangle unchanged for the identity transform", func() {
			result := shape2d.TransformedTriangle(triangle, shape2d.IdentityTransform())
			Expect(result.A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(result.B).To(dprectest.HaveVec2Coords(4.0, 0.0))
			Expect(result.C).To(dprectest.HaveVec2Coords(0.0, 3.0))
		})

		It("does not modify the original triangle", func() {
			shape2d.TransformedTriangle(triangle, shape2d.TranslationTransform(dprec.NewVec2(5.0, 5.0)))
			Expect(triangle.A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(triangle.B).To(dprectest.HaveVec2Coords(4.0, 0.0))
			Expect(triangle.C).To(dprectest.HaveVec2Coords(0.0, 3.0))
		})
	})

	Describe("Centroid", func() {
		It("returns the average of the three vertices", func() {
			Expect(triangle.Centroid()).To(dprectest.HaveVec2Coords(4.0/3.0, 1.0))
		})
	})

	Describe("edge lengths", func() {
		It("returns the length of each edge", func() {
			Expect(triangle.LengthAB()).To(BeNumerically("~", 4.0, 1e-6))
			Expect(triangle.LengthBC()).To(BeNumerically("~", 5.0, 1e-6))
			Expect(triangle.LengthCA()).To(BeNumerically("~", 3.0, 1e-6))
		})
	})

	Describe("SignedArea", func() {
		It("is positive for a counter-clockwise winding", func() {
			Expect(triangle.SignedArea()).To(BeNumerically("~", 6.0, 1e-6))
		})

		It("is negative for a clockwise winding", func() {
			cw := shape2d.NewTriangle(triangle.A, triangle.C, triangle.B)
			Expect(cw.SignedArea()).To(BeNumerically("~", -6.0, 1e-6))
		})

		It("is zero for collinear vertices", func() {
			degenerate := shape2d.NewTriangle(
				dprec.NewVec2(0.0, 0.0),
				dprec.NewVec2(1.0, 1.0),
				dprec.NewVec2(2.0, 2.0),
			)
			Expect(degenerate.SignedArea()).To(BeNumerically("~", 0.0, 1e-6))
		})
	})

	Describe("Area", func() {
		It("returns the unsigned area regardless of winding", func() {
			cw := shape2d.NewTriangle(triangle.A, triangle.C, triangle.B)
			Expect(triangle.Area()).To(BeNumerically("~", 6.0, 1e-6))
			Expect(cw.Area()).To(BeNumerically("~", 6.0, 1e-6))
		})
	})

	Describe("IsCCW", func() {
		It("returns true for a counter-clockwise winding", func() {
			Expect(triangle.IsCCW()).To(BeTrue())
		})

		It("returns false for a clockwise winding", func() {
			cw := shape2d.NewTriangle(triangle.A, triangle.C, triangle.B)
			Expect(cw.IsCCW()).To(BeFalse())
		})
	})

	Describe("ContainsPoint", func() {
		It("returns true for the centroid", func() {
			Expect(triangle.ContainsPoint(triangle.Centroid())).To(BeTrue())
		})

		It("returns true for a vertex", func() {
			Expect(triangle.ContainsPoint(triangle.A)).To(BeTrue())
		})

		It("returns true for a point on an edge", func() {
			// The midpoint of edge AB.
			Expect(triangle.ContainsPoint(dprec.NewVec2(2.0, 0.0))).To(BeTrue())
		})

		It("returns false for a point outside the triangle", func() {
			Expect(triangle.ContainsPoint(dprec.NewVec2(5.0, 5.0))).To(BeFalse())
		})

		It("returns false for a point beyond the hypotenuse", func() {
			// The hypotenuse lies on 3x+4y=12; the point (3,3) gives 21 > 12.
			Expect(triangle.ContainsPoint(dprec.NewVec2(3.0, 3.0))).To(BeFalse())
		})

		It("returns false for every point when the triangle is wound clockwise", func() {
			// ContainsPoint requires a counter-clockwise winding; a clockwise
			// triangle contains nothing, even its own centroid.
			cw := shape2d.NewTriangle(triangle.A, triangle.C, triangle.B)
			Expect(cw.ContainsPoint(cw.Centroid())).To(BeFalse())
		})
	})

	Describe("BoundingCircle", func() {
		It("is centered at the centroid", func() {
			Expect(triangle.BoundingCircle().Center).To(dprectest.HaveVec2Coords(4.0/3.0, 1.0))
		})

		It("has a radius equal to the distance to the farthest vertex", func() {
			// Vertex B is the farthest from the centroid.
			expected := dprec.Vec2Diff(triangle.B, triangle.Centroid()).Length()
			Expect(triangle.BoundingCircle().Radius).To(BeNumerically("~", expected, 1e-6))
		})

		It("contains all three vertices", func() {
			bc := triangle.BoundingCircle()
			Expect(bc.ContainsPoint(triangle.A)).To(BeTrue())
			Expect(bc.ContainsPoint(triangle.B)).To(BeTrue())
			Expect(bc.ContainsPoint(triangle.C)).To(BeTrue())
		})
	})
})
