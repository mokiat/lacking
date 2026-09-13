package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("Frustum", func() {

	Describe("NewFrustum", func() {
		It("stores the surfaces", func() {
			surfaces := [6]shape3d.Surface{
				{Normal: dprec.NewVec3(1.0, 0.0, 0.0), Distance: 1.0},
				{Normal: dprec.NewVec3(-1.0, 0.0, 0.0), Distance: 2.0},
				{Normal: dprec.NewVec3(0.0, 1.0, 0.0), Distance: 3.0},
				{Normal: dprec.NewVec3(0.0, -1.0, 0.0), Distance: 4.0},
				{Normal: dprec.NewVec3(0.0, 0.0, 1.0), Distance: 5.0},
				{Normal: dprec.NewVec3(0.0, 0.0, -1.0), Distance: 6.0},
			}
			frustum := shape3d.NewFrustum(surfaces)
			Expect(frustum.Surfaces).To(Equal(surfaces))
		})
	})

	Describe("FrustumFromProjection", func() {
		var frustum shape3d.Frustum

		BeforeEach(func() {
			// A symmetric perspective projection looking down the negative Z
			// axis, with a near plane at 1 and a far plane at 10.
			projection := dprec.PerspectiveMat4(-1.0, 1.0, -1.0, 1.0, 1.0, 10.0)
			frustum = shape3d.FrustumFromProjection(projection)
		})

		It("has unit normals", func() {
			for _, surface := range frustum.Surfaces {
				Expect(surface.Normal.Length()).To(BeNumerically("~", 1.0, 1e-9))
			}
		})

		It("contains a point in front of the camera", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, 0.0, -5.0))).To(BeTrue())
		})

		It("does not contain a point behind the camera", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, 0.0, 5.0))).To(BeFalse())
		})

		It("does not contain a point before the near plane", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, 0.0, -0.5))).To(BeFalse())
		})

		It("does not contain a point beyond the far plane", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, 0.0, -20.0))).To(BeFalse())
		})

		It("does not contain a point to the side of the frustum", func() {
			// At a distance of 5 the frustum is 5 units wide in each direction.
			Expect(frustum.ContainsPoint(dprec.NewVec3(8.0, 0.0, -5.0))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec3(-8.0, 0.0, -5.0))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, 8.0, -5.0))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, -8.0, -5.0))).To(BeFalse())
		})

		It("contains a point just within the side of the frustum", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec3(4.5, 0.0, -5.0))).To(BeTrue())
			Expect(frustum.ContainsPoint(dprec.NewVec3(-4.5, 0.0, -5.0))).To(BeTrue())
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, 4.5, -5.0))).To(BeTrue())
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, -4.5, -5.0))).To(BeTrue())
		})

		It("follows the view matrix when one is combined in", func() {
			projection := dprec.PerspectiveMat4(-1.0, 1.0, -1.0, 1.0, 1.0, 10.0)
			camera := dprec.TranslationMat4(100.0, 0.0, 0.0)
			view := dprec.InverseMat4(camera)
			moved := shape3d.FrustumFromProjection(dprec.Mat4Prod(projection, view))

			Expect(moved.ContainsPoint(dprec.NewVec3(100.0, 0.0, -5.0))).To(BeTrue())
			Expect(moved.ContainsPoint(dprec.NewVec3(0.0, 0.0, -5.0))).To(BeFalse())
		})
	})

	Describe("FrustumFromAABB", func() {
		var frustum shape3d.Frustum

		BeforeEach(func() {
			aabb := shape3d.NewAABB(-1.0, -2.0, -3.0, 4.0, 5.0, 6.0)
			frustum = shape3d.FrustumFromAABB(aabb)
		})

		It("has axis-aligned inward facing surfaces", func() {
			Expect(frustum.Surfaces[0].Normal).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(frustum.Surfaces[0].Distance).To(BeNumerically("~", -1.0, 1e-9))
			Expect(frustum.Surfaces[1].Normal).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))
			Expect(frustum.Surfaces[1].Distance).To(BeNumerically("~", -4.0, 1e-9))
		})

		It("contains a point inside the box", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, 0.0, 0.0))).To(BeTrue())
		})

		It("contains a point on a face of the box", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec3(4.0, 0.0, 0.0))).To(BeTrue())
			Expect(frustum.ContainsPoint(dprec.NewVec3(-1.0, -2.0, -3.0))).To(BeTrue())
		})

		It("does not contain points outside each side of the box", func() {
			Expect(frustum.ContainsPoint(dprec.NewVec3(-1.1, 0.0, 0.0))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec3(4.1, 0.0, 0.0))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, -2.1, 0.0))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, 5.1, 0.0))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, 0.0, -3.1))).To(BeFalse())
			Expect(frustum.ContainsPoint(dprec.NewVec3(0.0, 0.0, 6.1))).To(BeFalse())
		})
	})

})
