package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("Polygon", func() {
	var squarePoly shape2d.Polygon

	BeforeEach(func() {
		squarePoly = shape2d.NewPolygon([]shape2d.Edge{
			shape2d.NewEdge(dprec.NewVec2(-0.5, -0.5), dprec.NewVec2(0.5, -0.5)), // bottom
			shape2d.NewEdge(dprec.NewVec2(0.5, -0.5), dprec.NewVec2(0.5, 0.5)),   // right
			shape2d.NewEdge(dprec.NewVec2(0.5, 0.5), dprec.NewVec2(-0.5, 0.5)),   // top
			shape2d.NewEdge(dprec.NewVec2(-0.5, 0.5), dprec.NewVec2(-0.5, -0.5)), // left
		})
	})

	Describe("NewPolygon", func() {
		It("stores the provided edges", func() {
			Expect(squarePoly.Edges).To(HaveLen(4))
			Expect(squarePoly.Edges[0].A).To(dprectest.HaveVec2Coords(-0.5, -0.5))
			Expect(squarePoly.Edges[0].B).To(dprectest.HaveVec2Coords(0.5, -0.5))
		})
	})

	Describe("BoundingCircle", func() {
		It("returns a zero circle for an empty polygon", func() {
			empty := shape2d.NewPolygon(nil)
			bc := empty.BoundingCircle()
			Expect(bc.Position).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(bc.Radius).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("is centred at the average of all edge endpoints", func() {
			Expect(squarePoly.BoundingCircle().Position).To(dprectest.HaveVec2Coords(0.0, 0.0))
		})

		It("encompasses all edge endpoints", func() {
			bc := squarePoly.BoundingCircle()
			for _, edge := range squarePoly.Edges {
				Expect(bc.ContainsPoint(edge.A)).To(BeTrue())
				Expect(bc.ContainsPoint(edge.B)).To(BeTrue())
			}
		})
	})

	Describe("TransformedPolygon", func() {
		It("translates all edge endpoints", func() {
			transform := shape2d.TRTransform(dprec.NewVec2(1.0, 2.0), dprec.Radians(0.0))
			result := shape2d.TransformedPolygon(squarePoly, transform)
			Expect(result.Edges[0].A).To(dprectest.HaveVec2Coords(0.5, 1.5))
			Expect(result.Edges[0].B).To(dprectest.HaveVec2Coords(1.5, 1.5))
		})

		It("rotates all edge endpoints 90° CCW", func() {
			transform := shape2d.TRTransform(dprec.ZeroVec2(), dprec.Degrees(90.0))
			result := shape2d.TransformedPolygon(squarePoly, transform)
			Expect(result.Edges[0].A).To(dprectest.HaveVec2Coords(0.5, -0.5))
			Expect(result.Edges[0].B).To(dprectest.HaveVec2Coords(0.5, 0.5))
		})

		It("preserves edge count", func() {
			transform := shape2d.TRTransform(dprec.NewVec2(3.0, -1.0), dprec.Degrees(45.0))
			result := shape2d.TransformedPolygon(squarePoly, transform)
			Expect(result.Edges).To(HaveLen(len(squarePoly.Edges)))
		})
	})
})
