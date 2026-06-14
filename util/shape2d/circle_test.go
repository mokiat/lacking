package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("Circle", func() {
	Describe("ContainsPoint", func() {
		var circle shape2d.Circle

		BeforeEach(func() {
			circle = shape2d.NewCircle(dprec.NewVec2(3.0, 4.0), 2.0)
		})

		It("returns true for the center", func() {
			Expect(circle.ContainsPoint(dprec.NewVec2(3.0, 4.0))).To(BeTrue())
		})

		It("returns true for a point strictly inside", func() {
			Expect(circle.ContainsPoint(dprec.NewVec2(4.0, 4.0))).To(BeTrue())
		})

		It("returns true for a point exactly on the boundary", func() {
			Expect(circle.ContainsPoint(dprec.NewVec2(5.0, 4.0))).To(BeTrue())
		})

		It("returns false for a point strictly outside", func() {
			Expect(circle.ContainsPoint(dprec.NewVec2(5.1, 4.0))).To(BeFalse())
		})

		It("returns true only for the center when radius is zero", func() {
			dot := shape2d.NewCircle(dprec.NewVec2(1.0, 2.0), 0.0)
			Expect(dot.ContainsPoint(dprec.NewVec2(1.0, 2.0))).To(BeTrue())
			Expect(dot.ContainsPoint(dprec.NewVec2(1.1, 2.0))).To(BeFalse())
		})
	})

	Describe("TransformedCircle", func() {
		var source shape2d.Circle

		BeforeEach(func() {
			source = shape2d.NewCircle(dprec.NewVec2(1.0, 0.0), 2.0)
		})

		It("translates the position and preserves the radius", func() {
			transform := shape2d.TRTransform(dprec.NewVec2(3.0, 4.0), dprec.Radians(0.0))
			result := shape2d.TransformedCircle(source, transform)
			Expect(result.Position).To(dprectest.HaveVec2Coords(4.0, 4.0))
			Expect(result.Radius).To(BeNumerically("~", 2.0, 1e-6))
		})

		It("rotates the position and preserves the radius", func() {
			transform := shape2d.TRTransform(dprec.NewVec2(0.0, 0.0), dprec.Degrees(90.0))
			result := shape2d.TransformedCircle(source, transform)
			Expect(result.Position).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(result.Radius).To(BeNumerically("~", 2.0, 1e-6))
		})

		It("applies translation and rotation together", func() {
			transform := shape2d.TRTransform(dprec.NewVec2(3.0, 4.0), dprec.Degrees(90.0))
			result := shape2d.TransformedCircle(source, transform)
			Expect(result.Position).To(dprectest.HaveVec2Coords(3.0, 5.0))
			Expect(result.Radius).To(BeNumerically("~", 2.0, 1e-6))
		})
	})
})
