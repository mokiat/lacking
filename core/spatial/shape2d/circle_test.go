package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Circle", func() {

	Describe("ContainsPoint", func() {
		var circle shape2d.Circle

		BeforeEach(func() {
			circle = shape2d.Circle{
				Center: dprec.NewVec2(3.0, 4.0),
				Radius: 2.0,
			}
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
			dot := shape2d.Circle{
				Center: dprec.NewVec2(1.0, 2.0),
				Radius: 0.0,
			}
			Expect(dot.ContainsPoint(dprec.NewVec2(1.0, 2.0))).To(BeTrue())
			Expect(dot.ContainsPoint(dprec.NewVec2(1.1, 2.0))).To(BeFalse())
		})
	})

})
