package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Surface", func() {
	var offsetSurface shape2d.Surface

	BeforeEach(func() {
		// A line facing along the Y axis, offset 5 units up from the origin.
		offsetSurface = shape2d.Surface{
			Normal:   dprec.NewVec2(0.0, 1.0),
			Distance: 5.0,
		}
	})

	Describe("BasisXSurface", func() {
		It("faces along the X axis and passes through the origin", func() {
			s := shape2d.BasisXSurface()
			Expect(s.Normal).To(dprectest.HaveVec2Coords(1.0, 0.0))
			Expect(s.Distance).To(Equal(0.0))
		})
	})

	Describe("BasisYSurface", func() {
		It("faces along the Y axis and passes through the origin", func() {
			s := shape2d.BasisYSurface()
			Expect(s.Normal).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(s.Distance).To(Equal(0.0))
		})
	})

	Describe("Point", func() {
		It("returns the origin for a line through the origin", func() {
			Expect(shape2d.BasisYSurface().Point()).To(dprectest.HaveVec2Coords(0.0, 0.0))
		})

		It("returns the closest point to the origin for an offset line", func() {
			Expect(offsetSurface.Point()).To(dprectest.HaveVec2Coords(0.0, 5.0))
		})

		It("returns a point that lies on the line", func() {
			point := offsetSurface.Point()
			Expect(offsetSurface.SignedDistance(point)).To(BeNumerically("~", 0.0, 1e-10))
		})

		It("handles a negative distance", func() {
			surface := shape2d.Surface{
				Normal:   dprec.NewVec2(0.0, 1.0),
				Distance: -3.0,
			}
			Expect(surface.Point()).To(dprectest.HaveVec2Coords(0.0, -3.0))
		})
	})

	Describe("SignedDistance", func() {
		It("returns zero for a point on the line", func() {
			Expect(offsetSurface.SignedDistance(dprec.NewVec2(2.0, 5.0))).To(BeNumerically("~", 0.0, 1e-10))
		})

		It("returns a positive distance for a point on the side the normal faces", func() {
			Expect(offsetSurface.SignedDistance(dprec.NewVec2(0.0, 8.0))).To(BeNumerically("~", 3.0, 1e-10))
		})

		It("returns a negative distance for a point on the opposite side", func() {
			Expect(offsetSurface.SignedDistance(dprec.NewVec2(0.0, 1.0))).To(BeNumerically("~", -4.0, 1e-10))
		})

		It("ignores movement parallel to the line", func() {
			point := dprec.NewVec2(100.0, 8.0)
			Expect(offsetSurface.SignedDistance(point)).To(BeNumerically("~", 3.0, 1e-10))
		})

		It("measures distance along the normal for a line through the origin", func() {
			Expect(shape2d.BasisXSurface().SignedDistance(dprec.NewVec2(4.0, 9.0))).To(BeNumerically("~", 4.0, 1e-10))
		})
	})
})
