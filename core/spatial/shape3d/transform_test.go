package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("Transform", func() {
	var (
		translation dprec.Vec3
		rotZ90      shape3d.Rotation
	)

	BeforeEach(func() {
		translation = dprec.NewVec3(1.0, 2.0, 3.0)
		// A 90-degree counter-clockwise rotation about the Z axis.
		rotZ90 = shape3d.Rotation{
			BasisX: dprec.NewVec3(0.0, 1.0, 0.0),
			BasisY: dprec.NewVec3(-1.0, 0.0, 0.0),
			BasisZ: dprec.NewVec3(0.0, 0.0, 1.0),
		}
	})

	Describe("IdentityTransform", func() {
		It("has no translation and an identity rotation", func() {
			t := shape3d.IdentityTransform()
			Expect(t.Translation).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
			Expect(t.Rotation.BasisX).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(t.Rotation.BasisY).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(t.Rotation.BasisZ).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
		})

		It("leaves points unchanged", func() {
			t := shape3d.IdentityTransform()
			Expect(t.Apply(dprec.NewVec3(3.0, 4.0, 5.0))).To(dprectest.HaveVec3Coords(3.0, 4.0, 5.0))
		})
	})

	Describe("TranslationTransform", func() {
		It("stores the translation and an identity rotation", func() {
			t := shape3d.TranslationTransform(translation)
			Expect(t.Translation).To(dprectest.HaveVec3Coords(1.0, 2.0, 3.0))
			Expect(t.Rotation.BasisX).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(t.Rotation.BasisY).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(t.Rotation.BasisZ).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
		})

		It("offsets points without rotating them", func() {
			t := shape3d.TranslationTransform(translation)
			Expect(t.Apply(dprec.NewVec3(3.0, 4.0, 5.0))).To(dprectest.HaveVec3Coords(4.0, 6.0, 8.0))
		})
	})

	Describe("RotationTransform", func() {
		It("stores the rotation and a zero translation", func() {
			t := shape3d.RotationTransform(rotZ90)
			Expect(t.Translation).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
			Expect(t.Rotation.BasisX).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(t.Rotation.BasisY).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))
			Expect(t.Rotation.BasisZ).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
		})

		It("rotates points without offsetting them", func() {
			t := shape3d.RotationTransform(rotZ90)
			Expect(t.Apply(dprec.NewVec3(3.0, 4.0, 5.0))).To(dprectest.HaveVec3Coords(-4.0, 3.0, 5.0))
		})
	})

	Describe("TRTransform", func() {
		It("stores both the translation and the rotation", func() {
			t := shape3d.TRTransform(translation, rotZ90)
			Expect(t.Translation).To(dprectest.HaveVec3Coords(1.0, 2.0, 3.0))
			Expect(t.Rotation.BasisX).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(t.Rotation.BasisY).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))
			Expect(t.Rotation.BasisZ).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
		})

		It("rotates points and then offsets them", func() {
			t := shape3d.TRTransform(translation, rotZ90)
			// rotate (3,4,5) -> (-4,3,5), then translate by (1,2,3) -> (-3,5,8)
			Expect(t.Apply(dprec.NewVec3(3.0, 4.0, 5.0))).To(dprectest.HaveVec3Coords(-3.0, 5.0, 8.0))
		})
	})

	Describe("ChainedTransform", func() {
		It("equals applying the child transform and then the parent", func() {
			parent := shape3d.TRTransform(dprec.NewVec3(1.0, 2.0, 3.0), rotZ90)
			child := shape3d.TRTransform(dprec.NewVec3(-2.0, 4.0, 1.0), rotZ90)
			chained := shape3d.ChainedTransform(parent, child)

			point := dprec.NewVec3(5.0, -3.0, 2.0)
			expected := parent.Apply(child.Apply(point))
			Expect(chained.Apply(point)).To(dprectest.HaveVec3Coords(expected.X, expected.Y, expected.Z))
		})

		It("yields the original transform when chained with the identity", func() {
			t := shape3d.TRTransform(translation, rotZ90)
			identity := shape3d.IdentityTransform()

			leftChained := shape3d.ChainedTransform(identity, t)
			rightChained := shape3d.ChainedTransform(t, identity)

			point := dprec.NewVec3(5.0, -3.0, 2.0)
			expected := t.Apply(point)
			Expect(leftChained.Apply(point)).To(dprectest.HaveVec3Coords(expected.X, expected.Y, expected.Z))
			Expect(rightChained.Apply(point)).To(dprectest.HaveVec3Coords(expected.X, expected.Y, expected.Z))
		})

		It("composes two rotations about the Z axis into a 180-degree rotation", func() {
			rot := shape3d.RotationTransform(rotZ90)
			chained := shape3d.ChainedTransform(rot, rot)
			Expect(chained.Apply(dprec.NewVec3(1.0, 0.0, 0.0))).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))
		})
	})

	Describe("Inverse", func() {
		It("yields the identity for the identity transform", func() {
			inv := shape3d.IdentityTransform().Inverse()
			Expect(inv.Apply(dprec.NewVec3(3.0, 4.0, 5.0))).To(dprectest.HaveVec3Coords(3.0, 4.0, 5.0))
		})

		It("negates the translation of a translation-only transform", func() {
			inv := shape3d.TranslationTransform(translation).Inverse()
			Expect(inv.Translation).To(dprectest.HaveVec3Coords(-1.0, -2.0, -3.0))
		})

		It("undoes Apply on a general point", func() {
			t := shape3d.TRTransform(translation, rotZ90)
			point := dprec.NewVec3(5.0, -3.0, 2.0)
			Expect(t.Inverse().Apply(t.Apply(point))).To(dprectest.HaveVec3Coords(point.X, point.Y, point.Z))
		})

		It("undoes the transform when applied in the other order", func() {
			t := shape3d.TRTransform(translation, rotZ90)
			point := dprec.NewVec3(5.0, -3.0, 2.0)
			Expect(t.Apply(t.Inverse().Apply(point))).To(dprectest.HaveVec3Coords(point.X, point.Y, point.Z))
		})

		It("chains with the original transform to form the identity", func() {
			t := shape3d.TRTransform(translation, rotZ90)
			chained := shape3d.ChainedTransform(t, t.Inverse())
			point := dprec.NewVec3(5.0, -3.0, 2.0)
			Expect(chained.Apply(point)).To(dprectest.HaveVec3Coords(point.X, point.Y, point.Z))
		})
	})

	Describe("Apply", func() {
		It("rotates and then translates a point", func() {
			t := shape3d.TRTransform(translation, rotZ90)
			Expect(t.Apply(dprec.NewVec3(3.0, 4.0, 5.0))).To(dprectest.HaveVec3Coords(-3.0, 5.0, 8.0))
		})

		It("returns the translation for the zero point", func() {
			t := shape3d.TRTransform(translation, rotZ90)
			Expect(t.Apply(dprec.ZeroVec3())).To(dprectest.HaveVec3Coords(1.0, 2.0, 3.0))
		})
	})
})
