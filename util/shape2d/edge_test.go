package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("Edge", func() {
	var edge shape2d.Edge

	BeforeEach(func() {
		edge = shape2d.NewEdge(
			dprec.NewVec2(0.0, 0.0),
			dprec.NewVec2(3.0, 4.0),
		)
	})

	Describe("Length", func() {
		It("returns the Euclidean distance between A and B", func() {
			Expect(edge.Length()).To(BeNumerically("~", 5.0, 1e-6))
		})

		It("returns zero for a zero-length edge", func() {
			dot := shape2d.NewEdge(dprec.NewVec2(1.0, 2.0), dprec.NewVec2(1.0, 2.0))
			Expect(dot.Length()).To(BeNumerically("~", 0.0, 1e-6))
		})
	})

	Describe("Normal", func() {
		It("points right for an upward edge (CCW winding)", func() {
			e := shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(0.0, 1.0))
			Expect(e.Normal()).To(dprectest.HaveVec2Coords(1.0, 0.0))
		})

		It("points down for a rightward edge (CCW winding)", func() {
			e := shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(1.0, 0.0))
			Expect(e.Normal()).To(dprectest.HaveVec2Coords(0.0, -1.0))
		})

		It("points left for a downward edge (CCW winding)", func() {
			e := shape2d.NewEdge(dprec.NewVec2(0.0, 1.0), dprec.NewVec2(0.0, 0.0))
			Expect(e.Normal()).To(dprectest.HaveVec2Coords(-1.0, 0.0))
		})

		It("returns a unit vector", func() {
			Expect(edge.Normal().Length()).To(BeNumerically("~", 1.0, 1e-6))
		})
	})

	Describe("Center", func() {
		It("returns the midpoint of the edge", func() {
			Expect(edge.Center()).To(dprectest.HaveVec2Coords(1.5, 2.0))
		})
	})

	Describe("BoundingCircle", func() {
		It("is centered at the midpoint of the edge", func() {
			bc := edge.BoundingCircle()
			Expect(bc.Position).To(dprectest.HaveVec2Coords(1.5, 2.0))
		})

		It("has radius equal to half the edge length", func() {
			bc := edge.BoundingCircle()
			Expect(bc.Radius).To(BeNumerically("~", 2.5, 1e-6))
		})

		It("contains both endpoints", func() {
			bc := edge.BoundingCircle()
			Expect(bc.ContainsPoint(edge.A)).To(BeTrue())
			Expect(bc.ContainsPoint(edge.B)).To(BeTrue())
		})
	})

	Describe("TransformedEdge", func() {
		It("applies translation", func() {
			transform := shape2d.TRTransform(dprec.NewVec2(1.0, 2.0), dprec.Radians(0.0))
			result := shape2d.TransformedEdge(edge, transform)
			Expect(result.A).To(dprectest.HaveVec2Coords(1.0, 2.0))
			Expect(result.B).To(dprectest.HaveVec2Coords(4.0, 6.0))
		})

		It("applies 90° CCW rotation", func() {
			e := shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(1.0, 0.0))
			transform := shape2d.TRTransform(dprec.NewVec2(0.0, 0.0), dprec.Degrees(90.0))
			result := shape2d.TransformedEdge(e, transform)
			Expect(result.A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(result.B).To(dprectest.HaveVec2Coords(0.0, 1.0))
		})

		It("preserves edge length after transformation", func() {
			transform := shape2d.TRTransform(dprec.NewVec2(5.0, -3.0), dprec.Degrees(37.0))
			result := shape2d.TransformedEdge(edge, transform)
			Expect(result.Length()).To(BeNumerically("~", edge.Length(), 1e-6))
		})
	})
})
