package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("AABB", func() {

	Describe("NewAABB", func() {
		It("assigns the min and max components to the matching fields", func() {
			aabb := shape3d.NewAABB(1.0, 2.0, 3.0, 4.0, 5.0, 6.0)
			Expect(aabb.MinX).To(BeNumerically("~", 1.0, 1e-6))
			Expect(aabb.MinY).To(BeNumerically("~", 2.0, 1e-6))
			Expect(aabb.MinZ).To(BeNumerically("~", 3.0, 1e-6))
			Expect(aabb.MaxX).To(BeNumerically("~", 4.0, 1e-6))
			Expect(aabb.MaxY).To(BeNumerically("~", 5.0, 1e-6))
			Expect(aabb.MaxZ).To(BeNumerically("~", 6.0, 1e-6))
		})
	})

	Describe("EmptyAABB", func() {
		It("is empty", func() {
			Expect(shape3d.EmptyAABB().IsEmpty()).To(BeTrue())
		})

		It("expands to exactly enclose a single point when grown with min/max", func() {
			aabb := shape3d.EmptyAABB()
			point := dprec.NewVec3(3.0, -2.0, 5.0)
			aabb.MinX = min(aabb.MinX, point.X)
			aabb.MinY = min(aabb.MinY, point.Y)
			aabb.MinZ = min(aabb.MinZ, point.Z)
			aabb.MaxX = max(aabb.MaxX, point.X)
			aabb.MaxY = max(aabb.MaxY, point.Y)
			aabb.MaxZ = max(aabb.MaxZ, point.Z)

			Expect(aabb.MinX).To(BeNumerically("~", 3.0, 1e-6))
			Expect(aabb.MinY).To(BeNumerically("~", -2.0, 1e-6))
			Expect(aabb.MinZ).To(BeNumerically("~", 5.0, 1e-6))
			Expect(aabb.MaxX).To(BeNumerically("~", 3.0, 1e-6))
			Expect(aabb.MaxY).To(BeNumerically("~", -2.0, 1e-6))
			Expect(aabb.MaxZ).To(BeNumerically("~", 5.0, 1e-6))
			Expect(aabb.IsEmpty()).To(BeFalse())
		})
	})

	Describe("AABBFromSphere", func() {
		It("encloses the sphere tightly", func() {
			sphere := shape3d.NewSphere(dprec.NewVec3(3.0, 4.0, 5.0), 2.0)
			aabb := shape3d.AABBFromSphere(sphere)
			Expect(aabb.MinX).To(BeNumerically("~", 1.0, 1e-6))
			Expect(aabb.MinY).To(BeNumerically("~", 2.0, 1e-6))
			Expect(aabb.MinZ).To(BeNumerically("~", 3.0, 1e-6))
			Expect(aabb.MaxX).To(BeNumerically("~", 5.0, 1e-6))
			Expect(aabb.MaxY).To(BeNumerically("~", 6.0, 1e-6))
			Expect(aabb.MaxZ).To(BeNumerically("~", 7.0, 1e-6))
		})
	})

	Describe("AABBFromBox", func() {
		It("encloses an axis-aligned box tightly", func() {
			box := shape3d.NewBox(
				dprec.NewVec3(3.0, 4.0, 5.0),
				shape3d.IdentityRotation(),
				dprec.NewVec3(3.0, 4.0, 2.0),
			)
			aabb := shape3d.AABBFromBox(box)
			Expect(aabb.MinX).To(BeNumerically("~", 0.0, 1e-6))
			Expect(aabb.MinY).To(BeNumerically("~", 0.0, 1e-6))
			Expect(aabb.MinZ).To(BeNumerically("~", 3.0, 1e-6))
			Expect(aabb.MaxX).To(BeNumerically("~", 6.0, 1e-6))
			Expect(aabb.MaxY).To(BeNumerically("~", 8.0, 1e-6))
			Expect(aabb.MaxZ).To(BeNumerically("~", 7.0, 1e-6))
		})

		It("accounts for the box orientation", func() {
			// A 90deg rotation about Z swaps the contribution of the width and
			// height half-extents between the world X and Y axes: local X maps
			// to world Y, and local Y maps to world -X. See the equivalent case
			// in box_test.go's ContainsPoint tests.
			box := shape3d.NewBox(
				dprec.NewVec3(3.0, 4.0, 5.0),
				shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3())),
				dprec.NewVec3(3.0, 4.0, 2.0),
			)
			aabb := shape3d.AABBFromBox(box)
			Expect(aabb.MinX).To(BeNumerically("~", -1.0, 1e-6))
			Expect(aabb.MinY).To(BeNumerically("~", 1.0, 1e-6))
			Expect(aabb.MinZ).To(BeNumerically("~", 3.0, 1e-6))
			Expect(aabb.MaxX).To(BeNumerically("~", 7.0, 1e-6))
			Expect(aabb.MaxY).To(BeNumerically("~", 7.0, 1e-6))
			Expect(aabb.MaxZ).To(BeNumerically("~", 7.0, 1e-6))
		})

		It("degenerates to the center point when all dimensions are zero", func() {
			box := shape3d.NewBox(
				dprec.NewVec3(1.0, 2.0, 3.0),
				shape3d.IdentityRotation(),
				dprec.NewVec3(0.0, 0.0, 0.0),
			)
			aabb := shape3d.AABBFromBox(box)
			Expect(aabb.MinX).To(BeNumerically("~", 1.0, 1e-6))
			Expect(aabb.MinY).To(BeNumerically("~", 2.0, 1e-6))
			Expect(aabb.MinZ).To(BeNumerically("~", 3.0, 1e-6))
			Expect(aabb.MaxX).To(BeNumerically("~", 1.0, 1e-6))
			Expect(aabb.MaxY).To(BeNumerically("~", 2.0, 1e-6))
			Expect(aabb.MaxZ).To(BeNumerically("~", 3.0, 1e-6))
		})
	})

	Describe("IsEmpty", func() {
		It("returns false for a regular box", func() {
			aabb := shape3d.NewAABB(0.0, 0.0, 0.0, 1.0, 1.0, 1.0)
			Expect(aabb.IsEmpty()).To(BeFalse())
		})

		It("returns false for a single point", func() {
			aabb := shape3d.NewAABB(1.0, 2.0, 3.0, 1.0, 2.0, 3.0)
			Expect(aabb.IsEmpty()).To(BeFalse())
		})

		It("returns true when the minimum exceeds the maximum in X", func() {
			aabb := shape3d.NewAABB(1.0, 0.0, 0.0, 0.0, 1.0, 1.0)
			Expect(aabb.IsEmpty()).To(BeTrue())
		})

		It("returns true when the minimum exceeds the maximum in Y", func() {
			aabb := shape3d.NewAABB(0.0, 1.0, 0.0, 1.0, 0.0, 1.0)
			Expect(aabb.IsEmpty()).To(BeTrue())
		})

		It("returns true when the minimum exceeds the maximum in Z", func() {
			aabb := shape3d.NewAABB(0.0, 0.0, 1.0, 1.0, 1.0, 0.0)
			Expect(aabb.IsEmpty()).To(BeTrue())
		})

		It("can be called directly on a returned value", func() {
			sphere := shape3d.NewSphere(dprec.NewVec3(0.0, 0.0, 0.0), 1.0)
			Expect(shape3d.AABBFromSphere(sphere).IsEmpty()).To(BeFalse())
		})
	})

})
