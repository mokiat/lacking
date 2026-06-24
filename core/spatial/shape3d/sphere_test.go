package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("Sphere", func() {

	Describe("ContainsPoint", func() {
		var sphere shape3d.Sphere

		BeforeEach(func() {
			sphere = shape3d.Sphere{
				Center: dprec.NewVec3(3.0, 4.0, 5.0),
				Radius: 2.0,
			}
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
			dot := shape3d.Sphere{
				Center: dprec.NewVec3(1.0, 2.0, 3.0),
				Radius: 0.0,
			}
			Expect(dot.ContainsPoint(dprec.NewVec3(1.0, 2.0, 3.0))).To(BeTrue())
			Expect(dot.ContainsPoint(dprec.NewVec3(1.1, 2.0, 3.0))).To(BeFalse())
		})
	})

})
