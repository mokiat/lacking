package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("Surface", func() {
	var offsetSurface shape3d.Surface

	BeforeEach(func() {
		// A plane facing along the Y axis, offset 5 units up from the origin.
		offsetSurface = shape3d.Surface{
			Normal:   dprec.NewVec3(0.0, 1.0, 0.0),
			Distance: 5.0,
		}
	})

	Describe("BasisXSurface", func() {
		It("faces along the X axis and passes through the origin", func() {
			s := shape3d.BasisXSurface()
			Expect(s.Normal).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(s.Distance).To(Equal(0.0))
		})
	})

	Describe("BasisYSurface", func() {
		It("faces along the Y axis and passes through the origin", func() {
			s := shape3d.BasisYSurface()
			Expect(s.Normal).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(s.Distance).To(Equal(0.0))
		})
	})

	Describe("BasisZSurface", func() {
		It("faces along the Z axis and passes through the origin", func() {
			s := shape3d.BasisZSurface()
			Expect(s.Normal).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
			Expect(s.Distance).To(Equal(0.0))
		})
	})

	Describe("Point", func() {
		It("returns the origin for a surface through the origin", func() {
			Expect(shape3d.BasisYSurface().Point()).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
		})

		It("returns the closest point to the origin for an offset surface", func() {
			Expect(offsetSurface.Point()).To(dprectest.HaveVec3Coords(0.0, 5.0, 0.0))
		})

		It("returns a point that lies on the surface", func() {
			point := offsetSurface.Point()
			Expect(offsetSurface.SignedDistance(point)).To(BeNumerically("~", 0.0, 1e-10))
		})

		It("handles a negative distance", func() {
			surface := shape3d.Surface{
				Normal:   dprec.NewVec3(0.0, 1.0, 0.0),
				Distance: -3.0,
			}
			Expect(surface.Point()).To(dprectest.HaveVec3Coords(0.0, -3.0, 0.0))
		})
	})

	Describe("SignedDistance", func() {
		It("returns zero for a point on the surface", func() {
			Expect(offsetSurface.SignedDistance(dprec.NewVec3(2.0, 5.0, -7.0))).To(BeNumerically("~", 0.0, 1e-10))
		})

		It("returns a positive distance for a point on the side the normal faces", func() {
			Expect(offsetSurface.SignedDistance(dprec.NewVec3(0.0, 8.0, 0.0))).To(BeNumerically("~", 3.0, 1e-10))
		})

		It("returns a negative distance for a point on the opposite side", func() {
			Expect(offsetSurface.SignedDistance(dprec.NewVec3(0.0, 1.0, 0.0))).To(BeNumerically("~", -4.0, 1e-10))
		})

		It("ignores movement parallel to the surface", func() {
			point := dprec.NewVec3(100.0, 8.0, -50.0)
			Expect(offsetSurface.SignedDistance(point)).To(BeNumerically("~", 3.0, 1e-10))
		})

		It("measures distance along the normal for a surface through the origin", func() {
			Expect(shape3d.BasisXSurface().SignedDistance(dprec.NewVec3(4.0, 9.0, 9.0))).To(BeNumerically("~", 4.0, 1e-10))
		})
	})
})
