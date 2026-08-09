package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("AABB", func() {

	Describe("NewAABB", func() {
		It("assigns the min and max components to the matching fields", func() {
			aabb := shape2d.NewAABB(1.0, 2.0, 3.0, 4.0)
			Expect(aabb.MinX).To(BeNumerically("~", 1.0, 1e-6))
			Expect(aabb.MinY).To(BeNumerically("~", 2.0, 1e-6))
			Expect(aabb.MaxX).To(BeNumerically("~", 3.0, 1e-6))
			Expect(aabb.MaxY).To(BeNumerically("~", 4.0, 1e-6))
		})
	})

	Describe("EmptyAABB", func() {
		It("is empty", func() {
			Expect(shape2d.EmptyAABB().IsEmpty()).To(BeTrue())
		})

		It("expands to exactly enclose a single point when grown with min/max", func() {
			aabb := shape2d.EmptyAABB()
			point := dprec.NewVec2(3.0, -2.0)
			aabb.MinX = min(aabb.MinX, point.X)
			aabb.MinY = min(aabb.MinY, point.Y)
			aabb.MaxX = max(aabb.MaxX, point.X)
			aabb.MaxY = max(aabb.MaxY, point.Y)

			Expect(aabb.MinX).To(BeNumerically("~", 3.0, 1e-6))
			Expect(aabb.MinY).To(BeNumerically("~", -2.0, 1e-6))
			Expect(aabb.MaxX).To(BeNumerically("~", 3.0, 1e-6))
			Expect(aabb.MaxY).To(BeNumerically("~", -2.0, 1e-6))
			Expect(aabb.IsEmpty()).To(BeFalse())
		})
	})

	Describe("AABBFromCircle", func() {
		It("encloses the circle tightly", func() {
			circle := shape2d.NewCircle(dprec.NewVec2(3.0, 4.0), 2.0)
			aabb := shape2d.AABBFromCircle(circle)
			Expect(aabb.MinX).To(BeNumerically("~", 1.0, 1e-6))
			Expect(aabb.MinY).To(BeNumerically("~", 2.0, 1e-6))
			Expect(aabb.MaxX).To(BeNumerically("~", 5.0, 1e-6))
			Expect(aabb.MaxY).To(BeNumerically("~", 6.0, 1e-6))
		})
	})

	Describe("AABBFromRectangle", func() {
		It("encloses an axis-aligned rectangle tightly", func() {
			rect := shape2d.NewRectangle(
				dprec.NewVec2(3.0, 4.0),
				shape2d.IdentityRotation(),
				dprec.NewVec2(3.0, 4.0),
			)
			aabb := shape2d.AABBFromRectangle(rect)
			Expect(aabb.MinX).To(BeNumerically("~", 0.0, 1e-6))
			Expect(aabb.MinY).To(BeNumerically("~", 0.0, 1e-6))
			Expect(aabb.MaxX).To(BeNumerically("~", 6.0, 1e-6))
			Expect(aabb.MaxY).To(BeNumerically("~", 8.0, 1e-6))
		})

		It("accounts for the rectangle orientation", func() {
			// A 90deg CCW rotation swaps the contribution of the width and
			// height half-extents between the world X and Y axes. See the
			// equivalent case in rectangle_test.go's ContainsPoint tests.
			rect := shape2d.NewRectangle(
				dprec.NewVec2(3.0, 4.0),
				shape2d.RotationFromCosSin(0.0, 1.0),
				dprec.NewVec2(3.0, 4.0),
			)
			aabb := shape2d.AABBFromRectangle(rect)
			Expect(aabb.MinX).To(BeNumerically("~", -1.0, 1e-6))
			Expect(aabb.MinY).To(BeNumerically("~", 1.0, 1e-6))
			Expect(aabb.MaxX).To(BeNumerically("~", 7.0, 1e-6))
			Expect(aabb.MaxY).To(BeNumerically("~", 7.0, 1e-6))
		})

		It("degenerates to the center point when both dimensions are zero", func() {
			rect := shape2d.NewRectangle(
				dprec.NewVec2(1.0, 2.0),
				shape2d.IdentityRotation(),
				dprec.NewVec2(0.0, 0.0),
			)
			aabb := shape2d.AABBFromRectangle(rect)
			Expect(aabb.MinX).To(BeNumerically("~", 1.0, 1e-6))
			Expect(aabb.MinY).To(BeNumerically("~", 2.0, 1e-6))
			Expect(aabb.MaxX).To(BeNumerically("~", 1.0, 1e-6))
			Expect(aabb.MaxY).To(BeNumerically("~", 2.0, 1e-6))
		})
	})

	Describe("IsEmpty", func() {
		It("returns false for a regular box", func() {
			aabb := shape2d.NewAABB(0.0, 0.0, 1.0, 1.0)
			Expect(aabb.IsEmpty()).To(BeFalse())
		})

		It("returns false for a single point", func() {
			aabb := shape2d.NewAABB(1.0, 2.0, 1.0, 2.0)
			Expect(aabb.IsEmpty()).To(BeFalse())
		})

		It("returns true when the minimum exceeds the maximum in X", func() {
			aabb := shape2d.NewAABB(1.0, 0.0, 0.0, 1.0)
			Expect(aabb.IsEmpty()).To(BeTrue())
		})

		It("returns true when the minimum exceeds the maximum in Y", func() {
			aabb := shape2d.NewAABB(0.0, 1.0, 1.0, 0.0)
			Expect(aabb.IsEmpty()).To(BeTrue())
		})

		It("can be called directly on a returned value", func() {
			circle := shape2d.NewCircle(dprec.NewVec2(0.0, 0.0), 1.0)
			Expect(shape2d.AABBFromCircle(circle).IsEmpty()).To(BeFalse())
		})
	})

})
