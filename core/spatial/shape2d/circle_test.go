package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Circle", func() {

	Describe("TransformedCircle", func() {
		var circle shape2d.Circle

		BeforeEach(func() {
			circle = shape2d.NewCircle(dprec.NewVec2(3.0, 4.0), 2.0)
		})

		It("moves the center and keeps the radius", func() {
			transform := shape2d.TRTransform(
				dprec.NewVec2(10.0, 20.0),
				shape2d.RotationFromAngle(dprec.Degrees(90.0)),
			)
			result := shape2d.TransformedCircle(circle, transform)
			// Center (3,4) rotated by 90deg becomes (-4,3), then translated to (6,23).
			Expect(result.Center).To(dprectest.HaveVec2Coords(6.0, 23.0))
			Expect(result.Radius).To(BeNumerically("~", 2.0, 1e-6))
		})

		It("leaves the circle unchanged for the identity transform", func() {
			result := shape2d.TransformedCircle(circle, shape2d.IdentityTransform())
			Expect(result.Center).To(dprectest.HaveVec2Coords(3.0, 4.0))
			Expect(result.Radius).To(BeNumerically("~", 2.0, 1e-6))
		})

		It("does not modify the original circle", func() {
			_ = shape2d.TransformedCircle(circle, shape2d.TranslationTransform(dprec.NewVec2(5.0, 5.0)))
			Expect(circle.Center).To(dprectest.HaveVec2Coords(3.0, 4.0))
			Expect(circle.Radius).To(BeNumerically("~", 2.0, 1e-6))
		})
	})

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

})
