package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("Mesh", func() {
	var mesh shape3d.Mesh

	BeforeEach(func() {
		mesh = shape3d.Mesh{
			Triangles: []shape3d.Triangle{
				{
					A: dprec.NewVec3(0.0, 0.0, 0.0),
					B: dprec.NewVec3(6.0, 0.0, 0.0),
					C: dprec.NewVec3(0.0, 6.0, 0.0),
				},
				{
					A: dprec.NewVec3(0.0, 0.0, 0.0),
					B: dprec.NewVec3(0.0, 0.0, 6.0),
					C: dprec.NewVec3(6.0, 0.0, 0.0),
				},
			},
		}
	})

	Describe("TransformedMesh", func() {
		It("applies the transform to every vertex of every triangle", func() {
			transform := shape3d.TranslationTransform(dprec.NewVec3(10.0, 20.0, 30.0))
			result := shape3d.TransformedMesh(mesh, transform)
			Expect(result.Triangles).To(HaveLen(2))

			Expect(result.Triangles[0].A).To(dprectest.HaveVec3Coords(10.0, 20.0, 30.0))
			Expect(result.Triangles[0].B).To(dprectest.HaveVec3Coords(16.0, 20.0, 30.0))
			Expect(result.Triangles[0].C).To(dprectest.HaveVec3Coords(10.0, 26.0, 30.0))

			Expect(result.Triangles[1].A).To(dprectest.HaveVec3Coords(10.0, 20.0, 30.0))
			Expect(result.Triangles[1].B).To(dprectest.HaveVec3Coords(10.0, 20.0, 36.0))
			Expect(result.Triangles[1].C).To(dprectest.HaveVec3Coords(16.0, 20.0, 30.0))
		})

		It("leaves the mesh unchanged for the identity transform", func() {
			result := shape3d.TransformedMesh(mesh, shape3d.IdentityTransform())
			Expect(result.Triangles[0].A).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
			Expect(result.Triangles[0].B).To(dprectest.HaveVec3Coords(6.0, 0.0, 0.0))
			Expect(result.Triangles[0].C).To(dprectest.HaveVec3Coords(0.0, 6.0, 0.0))
		})

		It("does not modify the original mesh", func() {
			shape3d.TransformedMesh(mesh, shape3d.TranslationTransform(dprec.NewVec3(5.0, 5.0, 5.0)))
			Expect(mesh.Triangles[0].A).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
			Expect(mesh.Triangles[0].B).To(dprectest.HaveVec3Coords(6.0, 0.0, 0.0))
			Expect(mesh.Triangles[0].C).To(dprectest.HaveVec3Coords(0.0, 6.0, 0.0))
		})

		It("returns an empty mesh for an empty mesh", func() {
			result := shape3d.TransformedMesh(shape3d.Mesh{}, shape3d.IdentityTransform())
			Expect(result.Triangles).To(BeEmpty())
		})
	})

	Describe("BoundingSphere", func() {
		It("is centered at the average of all vertices", func() {
			Expect(mesh.BoundingSphere().Center).To(dprectest.HaveVec3Coords(2.0, 1.0, 1.0))
		})

		It("has a radius equal to the distance to the farthest vertex", func() {
			// The farthest vertices from the center (2,1,1) are (0,6,0) and
			// (0,0,6), both at a distance of sqrt(30).
			Expect(mesh.BoundingSphere().Radius).To(BeNumerically("~", dprec.Sqrt(30.0), 1e-6))
		})

		It("contains every vertex of every triangle", func() {
			bs := mesh.BoundingSphere()
			for _, triangle := range mesh.Triangles {
				Expect(bs.ContainsPoint(triangle.A)).To(BeTrue())
				Expect(bs.ContainsPoint(triangle.B)).To(BeTrue())
				Expect(bs.ContainsPoint(triangle.C)).To(BeTrue())
			}
		})

		It("matches the triangle bounding sphere for a single-triangle mesh", func() {
			triangle := shape3d.Triangle{
				A: dprec.NewVec3(0.0, 0.0, 0.0),
				B: dprec.NewVec3(4.0, 0.0, 0.0),
				C: dprec.NewVec3(0.0, 3.0, 0.0),
			}
			single := shape3d.Mesh{Triangles: []shape3d.Triangle{triangle}}
			bs := single.BoundingSphere()
			expected := triangle.BoundingSphere()
			Expect(bs.Center).To(dprectest.HaveVec3Coords(expected.Center.X, expected.Center.Y, expected.Center.Z))
			Expect(bs.Radius).To(BeNumerically("~", expected.Radius, 1e-6))
		})

		It("returns the zero sphere for an empty mesh", func() {
			bs := shape3d.Mesh{}.BoundingSphere()
			Expect(bs.Center).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
			Expect(bs.Radius).To(Equal(0.0))
		})
	})
})
