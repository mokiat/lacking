package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Contact", func() {
	var contact shape2d.Contact

	BeforeEach(func() {
		contact = shape2d.Contact{
			TargetPoint:  dprec.NewVec2(10.0, 0.0),
			TargetNormal: dprec.NewVec2(1.0, 0.0),
			Depth:        2.0,
		}
	})

	Describe("EvalSourcePoint", func() {
		It("lies Depth away from TargetPoint along the inverse normal", func() {
			Expect(contact.EvalSourcePoint()).To(dprectest.HaveVec2Coords(8.0, 0.0))
		})
	})

	Describe("EvalSourceNormal", func() {
		It("is the inverse of TargetNormal", func() {
			Expect(contact.EvalSourceNormal()).To(dprectest.HaveVec2Coords(-1.0, 0.0))
		})
	})

	Describe("Flipped", func() {
		It("promotes the source point and normal to the target", func() {
			flipped := contact.Flipped()
			Expect(flipped.TargetPoint).To(dprectest.HaveVec2Coords(8.0, 0.0))
			Expect(flipped.TargetNormal).To(dprectest.HaveVec2Coords(-1.0, 0.0))
			Expect(flipped.Depth).To(BeNumerically("~", 2.0, 1e-6))
		})

		It("round-trips back to the original when applied twice", func() {
			result := contact.Flipped().Flipped()
			Expect(result.TargetPoint).To(dprectest.HaveVec2Coords(contact.TargetPoint.X, contact.TargetPoint.Y))
			Expect(result.TargetNormal).To(dprectest.HaveVec2Coords(contact.TargetNormal.X, contact.TargetNormal.Y))
			Expect(result.Depth).To(BeNumerically("~", contact.Depth, 1e-6))
		})

		It("swaps the roles of the source and target points", func() {
			flipped := contact.Flipped()
			Expect(flipped.EvalSourcePoint()).To(dprectest.HaveVec2Coords(contact.TargetPoint.X, contact.TargetPoint.Y))
		})
	})
})
