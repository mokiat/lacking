package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("Box", func() {
	var box shape3d.Box

	BeforeEach(func() {
		box = shape3d.Box{
			Center:   dprec.NewVec3(3.0, 4.0, 5.0),
			Rotation: shape3d.IdentityRotation(),
			Width:    6.0,
			Height:   8.0,
			Length:   4.0,
		}
	})

	Describe("TransformedBox", func() {
		It("moves the center, composes the rotation and keeps the size", func() {
			transform := shape3d.TRTransform(
				dprec.NewVec3(10.0, 20.0, 30.0),
				shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3())),
			)
			result := shape3d.TransformedBox(box, transform)
			// Center (3,4,5) rotated by 90deg around Z becomes (-4,3,5), then translated to (6,23,35).
			Expect(result.Center).To(dprectest.HaveVec3Coords(6.0, 23.0, 35.0))
			// Identity rotation composed with a 90deg Z rotation yields a 90deg Z rotation.
			Expect(result.Rotation.BasisX).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(result.Rotation.BasisY).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))
			Expect(result.Rotation.BasisZ).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
			Expect(result.Width).To(BeNumerically("~", 6.0, 1e-6))
			Expect(result.Height).To(BeNumerically("~", 8.0, 1e-6))
			Expect(result.Length).To(BeNumerically("~", 4.0, 1e-6))
		})

		It("leaves the box unchanged for the identity transform", func() {
			result := shape3d.TransformedBox(box, shape3d.IdentityTransform())
			Expect(result.Center).To(dprectest.HaveVec3Coords(3.0, 4.0, 5.0))
			Expect(result.Rotation.BasisX).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(result.Rotation.BasisY).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(result.Rotation.BasisZ).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
			Expect(result.Width).To(BeNumerically("~", 6.0, 1e-6))
			Expect(result.Height).To(BeNumerically("~", 8.0, 1e-6))
			Expect(result.Length).To(BeNumerically("~", 4.0, 1e-6))
		})

		It("does not modify the original box", func() {
			shape3d.TransformedBox(box, shape3d.TranslationTransform(dprec.NewVec3(5.0, 5.0, 5.0)))
			Expect(box.Center).To(dprectest.HaveVec3Coords(3.0, 4.0, 5.0))
			Expect(box.Width).To(BeNumerically("~", 6.0, 1e-6))
			Expect(box.Height).To(BeNumerically("~", 8.0, 1e-6))
			Expect(box.Length).To(BeNumerically("~", 4.0, 1e-6))
		})
	})

	Describe("ContainsPoint", func() {
		It("returns true for the center", func() {
			Expect(box.ContainsPoint(dprec.NewVec3(3.0, 4.0, 5.0))).To(BeTrue())
		})

		It("returns true for a point strictly inside", func() {
			Expect(box.ContainsPoint(dprec.NewVec3(4.0, 5.0, 6.0))).To(BeTrue())
		})

		It("returns true for a point on a face", func() {
			Expect(box.ContainsPoint(dprec.NewVec3(6.0, 4.0, 5.0))).To(BeTrue())
		})

		It("returns true for a corner", func() {
			Expect(box.ContainsPoint(dprec.NewVec3(6.0, 8.0, 7.0))).To(BeTrue())
		})

		It("returns false for a point outside in X", func() {
			Expect(box.ContainsPoint(dprec.NewVec3(6.1, 4.0, 5.0))).To(BeFalse())
		})

		It("returns false for a point outside in Y", func() {
			Expect(box.ContainsPoint(dprec.NewVec3(3.0, 8.1, 5.0))).To(BeFalse())
		})

		It("returns false for a point outside in Z", func() {
			Expect(box.ContainsPoint(dprec.NewVec3(3.0, 4.0, 7.1))).To(BeFalse())
		})

		It("returns true only for the center when all dimensions are zero", func() {
			dot := shape3d.Box{
				Center:   dprec.NewVec3(1.0, 2.0, 3.0),
				Rotation: shape3d.IdentityRotation(),
				Width:    0.0,
				Height:   0.0,
				Length:   0.0,
			}
			Expect(dot.ContainsPoint(dprec.NewVec3(1.0, 2.0, 3.0))).To(BeTrue())
			Expect(dot.ContainsPoint(dprec.NewVec3(1.1, 2.0, 3.0))).To(BeFalse())
		})

		Context("with 90-degree CCW rotation about the Z axis", func() {
			var rotated shape3d.Box

			BeforeEach(func() {
				rotated = shape3d.Box{
					Center:   dprec.NewVec3(3.0, 4.0, 5.0),
					Rotation: shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3())),
					Width:    6.0,
					Height:   8.0,
					Length:   4.0,
				}
			})

			It("contains a point that lies outside the axis-aligned box", func() {
				// Offset (3.5,0,0) exceeds the unrotated width half-extent of 3
				// along world X, but is within the rotated height half-extent of 4.
				Expect(rotated.ContainsPoint(dprec.NewVec3(6.5, 4.0, 5.0))).To(BeTrue())
			})

			It("rejects a point that lies inside the axis-aligned box", func() {
				// Offset (0,3.5,0) is within the unrotated height half-extent of
				// 4 along world Y, but exceeds the rotated width half-extent of 3.
				Expect(rotated.ContainsPoint(dprec.NewVec3(3.0, 7.5, 5.0))).To(BeFalse())
			})

			It("contains a point on the rotated width boundary in world Y", func() {
				Expect(rotated.ContainsPoint(dprec.NewVec3(3.0, 7.0, 5.0))).To(BeTrue())
			})

			It("rejects a point just beyond the rotated width boundary in world Y", func() {
				Expect(rotated.ContainsPoint(dprec.NewVec3(3.0, 7.1, 5.0))).To(BeFalse())
			})

			It("keeps the Z extent unchanged by a Z rotation", func() {
				Expect(rotated.ContainsPoint(dprec.NewVec3(3.0, 4.0, 7.0))).To(BeTrue())
				Expect(rotated.ContainsPoint(dprec.NewVec3(3.0, 4.0, 7.1))).To(BeFalse())
			})
		})
	})

	Describe("BoundingSphere", func() {
		It("is centered at the center of the box", func() {
			bs := box.BoundingSphere()
			Expect(bs.Center).To(dprectest.HaveVec3Coords(3.0, 4.0, 5.0))
		})

		It("has radius equal to half the diagonal", func() {
			bs := box.BoundingSphere()
			// half extents (3,4,2) -> sqrt(9+16+4) = sqrt(29).
			Expect(bs.Radius).To(BeNumerically("~", dprec.Sqrt(29.0), 1e-6))
		})

		It("contains the area just inside the eight corners of the box", func() {
			bs := box.BoundingSphere()
			for _, sx := range []float64{-1.0, 1.0} {
				for _, sy := range []float64{-1.0, 1.0} {
					for _, sz := range []float64{-1.0, 1.0} {
						corner := dprec.NewVec3(
							3.0+sx*2.99,
							4.0+sy*3.99,
							5.0+sz*1.99,
						)
						Expect(bs.ContainsPoint(corner)).To(BeTrue())
					}
				}
			}
		})
	})
})
