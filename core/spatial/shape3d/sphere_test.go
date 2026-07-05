package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("Sphere", func() {

	Describe("TransformedSphere", func() {
		var sphere shape3d.Sphere

		BeforeEach(func() {
			sphere = shape3d.NewSphere(dprec.NewVec3(3.0, 4.0, 5.0), 2.0)
		})

		It("moves the center and keeps the radius", func() {
			transform := shape3d.TRTransform(
				dprec.NewVec3(10.0, 20.0, 30.0),
				shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3())),
			)
			result := shape3d.TransformedSphere(sphere, transform)
			// Center (3,4,5) rotated by 90deg around Z becomes (-4,3,5), then translated to (6,23,35).
			Expect(result.Center).To(dprectest.HaveVec3Coords(6.0, 23.0, 35.0))
			Expect(result.Radius).To(BeNumerically("~", 2.0, 1e-6))
		})

		It("leaves the sphere unchanged for the identity transform", func() {
			result := shape3d.TransformedSphere(sphere, shape3d.IdentityTransform())
			Expect(result.Center).To(dprectest.HaveVec3Coords(3.0, 4.0, 5.0))
			Expect(result.Radius).To(BeNumerically("~", 2.0, 1e-6))
		})

		It("does not modify the original sphere", func() {
			_ = shape3d.TransformedSphere(sphere, shape3d.TranslationTransform(dprec.NewVec3(5.0, 5.0, 5.0)))
			Expect(sphere.Center).To(dprectest.HaveVec3Coords(3.0, 4.0, 5.0))
			Expect(sphere.Radius).To(BeNumerically("~", 2.0, 1e-6))
		})
	})

	Describe("ContainsPoint", func() {
		var sphere shape3d.Sphere

		BeforeEach(func() {
			sphere = shape3d.NewSphere(dprec.NewVec3(3.0, 4.0, 5.0), 2.0)
		})

		It("returns true for the center", func() {
			Expect(sphere.ContainsPoint(dprec.NewVec3(3.0, 4.0, 5.0))).To(BeTrue())
		})

		It("returns true for a point strictly inside", func() {
			Expect(sphere.ContainsPoint(dprec.NewVec3(4.0, 4.0, 5.0))).To(BeTrue())
		})

		It("returns true for a point exactly on the boundary", func() {
			Expect(sphere.ContainsPoint(dprec.NewVec3(5.0, 4.0, 5.0))).To(BeTrue())
		})

		It("returns false for a point strictly outside", func() {
			Expect(sphere.ContainsPoint(dprec.NewVec3(5.1, 4.0, 5.0))).To(BeFalse())
		})

		It("returns true only for the center when radius is zero", func() {
			dot := shape3d.NewSphere(dprec.NewVec3(1.0, 2.0, 3.0), 0.0)
			Expect(dot.ContainsPoint(dprec.NewVec3(1.0, 2.0, 3.0))).To(BeTrue())
			Expect(dot.ContainsPoint(dprec.NewVec3(1.1, 2.0, 3.0))).To(BeFalse())
		})
	})

})
