package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/testing/sprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Capsule", func() {
	var capsule shape2d.Capsule

	BeforeEach(func() {
		capsule = shape2d.Capsule{
			A:      sprec.NewVec2(0.0, 0.0),
			B:      sprec.NewVec2(4.0, 0.0),
			Radius: 1.0,
		}
	})

	Describe("Spine", func() {
		It("returns the segment between the endpoints", func() {
			spine := capsule.Spine()
			Expect(spine.A).To(sprectest.HaveVec2Coords(0.0, 0.0))
			Expect(spine.B).To(sprectest.HaveVec2Coords(4.0, 0.0))
		})
	})

	Describe("ContainsPoint", func() {
		It("returns true for a point on the spine", func() {
			Expect(capsule.ContainsPoint(sprec.NewVec2(2.0, 0.0))).To(BeTrue())
		})

		It("returns true for an endpoint", func() {
			Expect(capsule.ContainsPoint(sprec.NewVec2(0.0, 0.0))).To(BeTrue())
			Expect(capsule.ContainsPoint(sprec.NewVec2(4.0, 0.0))).To(BeTrue())
		})

		It("returns true for a point within the radius of the spine", func() {
			Expect(capsule.ContainsPoint(sprec.NewVec2(2.0, 0.9))).To(BeTrue())
		})

		It("returns true for a point on the edge of the spine", func() {
			Expect(capsule.ContainsPoint(sprec.NewVec2(2.0, 1.0))).To(BeTrue())
		})

		It("returns true within the rounded cap beyond an endpoint", func() {
			Expect(capsule.ContainsPoint(sprec.NewVec2(4.5, 0.0))).To(BeTrue())
		})

		It("returns false beyond the radius of the spine", func() {
			Expect(capsule.ContainsPoint(sprec.NewVec2(2.0, 1.1))).To(BeFalse())
		})

		It("returns false beyond the rounded cap", func() {
			Expect(capsule.ContainsPoint(sprec.NewVec2(5.1, 0.0))).To(BeFalse())
		})

		It("returns false diagonally past an endpoint", func() {
			Expect(capsule.ContainsPoint(sprec.NewVec2(4.8, 0.8))).To(BeFalse())
		})

		It("behaves like a circle for a zero-length spine", func() {
			dot := shape2d.Capsule{
				A:      sprec.NewVec2(1.0, 2.0),
				B:      sprec.NewVec2(1.0, 2.0),
				Radius: 2.0,
			}
			Expect(dot.ContainsPoint(sprec.NewVec2(1.0, 2.0))).To(BeTrue())
			Expect(dot.ContainsPoint(sprec.NewVec2(3.0, 2.0))).To(BeTrue())
			Expect(dot.ContainsPoint(sprec.NewVec2(3.1, 2.0))).To(BeFalse())
		})
	})

	Describe("BoundingCircle", func() {
		It("is centered at the midpoint of the spine", func() {
			bc := capsule.BoundingCircle()
			Expect(bc.Center).To(sprectest.HaveVec2Coords(2.0, 0.0))
		})

		It("has radius equal to half the spine length plus the capsule radius", func() {
			bc := capsule.BoundingCircle()
			Expect(bc.Radius).To(BeNumerically("~", 3.0, 1e-6))
		})

		It("contains the extremes of the capsule", func() {
			bc := capsule.BoundingCircle()
			Expect(bc.ContainsPoint(sprec.NewVec2(5.0, 0.0))).To(BeTrue())
			Expect(bc.ContainsPoint(sprec.NewVec2(-1.0, 0.0))).To(BeTrue())
			Expect(bc.ContainsPoint(sprec.NewVec2(2.0, 1.0))).To(BeTrue())
		})
	})
})
