package shape2d_test

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape2d"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Triangle", func() {

	Describe("IsCCW", func() {
		Specify("counter-clockwise ordered vertices", func() {
			triangle := shape2d.NewTriangle(
				dprec.NewVec2(0.0, 1.0),
				dprec.NewVec2(-1.0, -0.5),
				dprec.NewVec2(1.0, 0.5),
			)
			Expect(triangle.IsCCW()).To(BeTrue())
		})

		Specify("clockwise ordered vertices", func() {
			triangle := shape2d.NewTriangle(
				dprec.NewVec2(0.0, 1.0),
				dprec.NewVec2(1.0, 0.5),
				dprec.NewVec2(-1.0, -0.5),
			)
			Expect(triangle.IsCCW()).To(BeFalse())
		})
	})

	Describe("ContainsPoint", func() {
		Specify("point inside triangle", func() {
			point := dprec.NewVec2(0.15, 0.15)
			triangle := shape2d.NewTriangle(
				dprec.NewVec2(0.0, 1.0),
				dprec.NewVec2(-1.0, -0.5),
				dprec.NewVec2(1.0, 0.5),
			)
			Expect(triangle.ContainsPoint(point)).To(BeTrue())
		})

		Specify("point outside triangle", func() {
			point := dprec.NewVec2(0.25, 1.25)
			triangle := shape2d.NewTriangle(
				dprec.NewVec2(0.0, 1.0),
				dprec.NewVec2(-1.0, -0.5),
				dprec.NewVec2(1.0, 0.5),
			)
			Expect(triangle.ContainsPoint(point)).To(BeFalse())
		})
	})

})
