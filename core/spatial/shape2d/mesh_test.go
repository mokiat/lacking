package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Mesh", func() {
	var mesh shape2d.Mesh

	BeforeEach(func() {
		// Two parallel edges whose four endpoints are the corners of a square
		// centered at (1,1), so the bounding circle comes out to clean values.
		mesh = shape2d.NewMesh([]shape2d.Edge{
			shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(2.0, 0.0)),
			shape2d.NewEdge(dprec.NewVec2(0.0, 2.0), dprec.NewVec2(2.0, 2.0)),
		})
	})

	Describe("NewMesh", func() {
		It("holds the given edges", func() {
			edges := []shape2d.Edge{
				shape2d.NewEdge(dprec.NewVec2(1.0, 1.0), dprec.NewVec2(2.0, 2.0)),
			}
			Expect(shape2d.NewMesh(edges).Edges).To(Equal(edges))
		})
	})

	Describe("TransformedMesh", func() {
		It("applies the transform to every edge", func() {
			transform := shape2d.TRTransform(
				dprec.NewVec2(10.0, 20.0),
				shape2d.RotationFromAngle(dprec.Degrees(90.0)),
			)
			result := shape2d.TransformedMesh(mesh, transform)

			// Each endpoint is rotated 90deg, then translated by (10,20).
			Expect(result.Edges[0].A).To(dprectest.HaveVec2Coords(10.0, 20.0))
			Expect(result.Edges[0].B).To(dprectest.HaveVec2Coords(10.0, 22.0))
			Expect(result.Edges[1].A).To(dprectest.HaveVec2Coords(8.0, 20.0))
			Expect(result.Edges[1].B).To(dprectest.HaveVec2Coords(8.0, 22.0))
		})

		It("leaves the mesh unchanged for the identity transform", func() {
			result := shape2d.TransformedMesh(mesh, shape2d.IdentityTransform())
			Expect(result.Edges[0].A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(result.Edges[0].B).To(dprectest.HaveVec2Coords(2.0, 0.0))
			Expect(result.Edges[1].A).To(dprectest.HaveVec2Coords(0.0, 2.0))
			Expect(result.Edges[1].B).To(dprectest.HaveVec2Coords(2.0, 2.0))
		})

		It("does not modify the original mesh", func() {
			shape2d.TransformedMesh(mesh, shape2d.TranslationTransform(dprec.NewVec2(5.0, 5.0)))
			Expect(mesh.Edges[0].A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(mesh.Edges[0].B).To(dprectest.HaveVec2Coords(2.0, 0.0))
		})

		It("returns an empty mesh for an empty mesh", func() {
			result := shape2d.TransformedMesh(shape2d.Mesh{}, shape2d.TranslationTransform(dprec.NewVec2(5.0, 5.0)))
			Expect(result.Edges).To(BeEmpty())
		})
	})

	Describe("BoundingCircle", func() {
		It("is centered at the average of all endpoints", func() {
			Expect(mesh.BoundingCircle().Center).To(dprectest.HaveVec2Coords(1.0, 1.0))
		})

		It("has a radius equal to the distance to the farthest endpoint", func() {
			// All four corners are sqrt(2) from the center (1,1).
			Expect(mesh.BoundingCircle().Radius).To(BeNumerically("~", dprec.Sqrt(2.0), 1e-6))
		})

		It("contains every endpoint", func() {
			bc := mesh.BoundingCircle()
			for _, edge := range mesh.Edges {
				Expect(bc.ContainsPoint(edge.A)).To(BeTrue())
				Expect(bc.ContainsPoint(edge.B)).To(BeTrue())
			}
		})

		It("matches the edge's bounding circle for a single-edge mesh", func() {
			edge := shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(4.0, 0.0))
			single := shape2d.NewMesh([]shape2d.Edge{edge})

			meshCircle := single.BoundingCircle()
			edgeCircle := edge.BoundingCircle()
			Expect(meshCircle.Center).To(dprectest.HaveVec2Coords(edgeCircle.Center.X, edgeCircle.Center.Y))
			Expect(meshCircle.Radius).To(BeNumerically("~", edgeCircle.Radius, 1e-6))
		})

		It("returns the zero circle for an empty mesh", func() {
			bc := shape2d.Mesh{}.BoundingCircle()
			Expect(bc.Center).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(bc.Radius).To(Equal(0.0))
		})
	})
})
