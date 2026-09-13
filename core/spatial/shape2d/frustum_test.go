package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Frustum", func() {

	Describe("NewFrustum", func() {
		It("stores the surfaces", func() {
			surfaces := [4]shape2d.Surface{
				{Normal: dprec.NewVec2(1.0, 0.0), Distance: 1.0},
				{Normal: dprec.NewVec2(-1.0, 0.0), Distance: 2.0},
				{Normal: dprec.NewVec2(0.0, 1.0), Distance: 3.0},
				{Normal: dprec.NewVec2(0.0, -1.0), Distance: 4.0},
			}
			frustum := shape2d.NewFrustum(surfaces)
			Expect(frustum.Surfaces).To(Equal(surfaces))
		})
	})

	Describe("FrustumFromProjection", func() {
		var frustum shape2d.Frustum

		BeforeEach(func() {
			// A symmetric orthographic projection covering the area from -5 to 5
			// in each direction.
			projection := dprec.OrthoMat3(-5.0, 5.0, 5.0, -5.0)
			frustum = shape2d.FrustumFromProjection(projection)
		})

		It("has unit normals", func() {
			for _, surface := range frustum.Surfaces {
				Expect(surface.Normal.Length()).To(BeNumerically("~", 1.0, 1e-9))
			}
		})

		It("contains a point inside the region", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec2(0.0, 0.0))).To(BeTrue())
		})

		It("does not contain a point outside each side of the region", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec2(6.0, 0.0))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec2(-6.0, 0.0))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec2(0.0, 6.0))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec2(0.0, -6.0))).To(BeFalse())
		})

		It("contains a point just within each side of the region", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec2(4.5, 0.0))).To(BeTrue())
			Expect(frustum.ContainsPoint(dprec.NewVec2(-4.5, 0.0))).To(BeTrue())
			Expect(frustum.ContainsPoint(dprec.NewVec2(0.0, 4.5))).To(BeTrue())
			Expect(frustum.ContainsPoint(dprec.NewVec2(0.0, -4.5))).To(BeTrue())
		})

		It("follows the view matrix when one is combined in", func() {
			projection := dprec.OrthoMat3(-5.0, 5.0, 5.0, -5.0)
			camera := dprec.TranslationMat3(100.0, 0.0)
			view := dprec.InverseMat3(camera)
			moved := shape2d.FrustumFromProjection(dprec.Mat3Prod(projection, view))

			Expect(moved.ContainsPoint(dprec.NewVec2(100.0, 0.0))).To(BeTrue())
			Expect(moved.ContainsPoint(dprec.NewVec2(0.0, 0.0))).To(BeFalse())
		})
	})

	Describe("FrustumFromAABB", func() {
		var frustum shape2d.Frustum

		BeforeEach(func() {
			aabb := shape2d.NewAABB(-1.0, -2.0, 4.0, 5.0)
			frustum = shape2d.FrustumFromAABB(aabb)
		})

		It("has axis-aligned inward facing surfaces", func() {
			Expect(frustum.Surfaces[0].Normal).To(dprectest.HaveVec2Coords(1.0, 0.0))
			Expect(frustum.Surfaces[0].Distance).To(BeNumerically("~", -1.0, 1e-9))
			Expect(frustum.Surfaces[1].Normal).To(dprectest.HaveVec2Coords(-1.0, 0.0))
			Expect(frustum.Surfaces[1].Distance).To(BeNumerically("~", -4.0, 1e-9))
		})

		It("contains a point inside the box", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec2(0.0, 0.0))).To(BeTrue())
		})

		It("contains a point on an edge of the box", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec2(4.0, 0.0))).To(BeTrue())
			Expect(frustum.ContainsPoint(dprec.NewVec2(-1.0, -2.0))).To(BeTrue())
		})

		It("does not contain points outside each side of the box", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec2(-1.1, 0.0))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec2(4.1, 0.0))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec2(0.0, -2.1))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec2(0.0, 5.1))).To(BeFalse())
		})
	})

})
