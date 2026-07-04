package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Transform", func() {
	var (
		translation dprec.Vec2
		rot90       shape2d.Rotation
	)

	BeforeEach(func() {
		translation = dprec.NewVec2(1.0, 2.0)
		// A 90-degree counter-clockwise rotation.
		rot90 = shape2d.Rotation{
			BasisX: dprec.NewVec2(0.0, 1.0),
			BasisY: dprec.NewVec2(-1.0, 0.0),
		}
	})

	Describe("IdentityTransform", func() {
		It("has no translation and an identity rotation", func() {
			t := shape2d.IdentityTransform()
			Expect(t.Translation).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(t.Rotation.BasisX).To(dprectest.HaveVec2Coords(1.0, 0.0))
			Expect(t.Rotation.BasisY).To(dprectest.HaveVec2Coords(0.0, 1.0))
		})

		It("leaves points unchanged", func() {
			t := shape2d.IdentityTransform()
			Expect(t.Apply(dprec.NewVec2(3.0, 4.0))).To(dprectest.HaveVec2Coords(3.0, 4.0))
		})
	})

	Describe("TranslationTransform", func() {
		It("stores the translation and an identity rotation", func() {
			t := shape2d.TranslationTransform(translation)
			Expect(t.Translation).To(dprectest.HaveVec2Coords(1.0, 2.0))
			Expect(t.Rotation.BasisX).To(dprectest.HaveVec2Coords(1.0, 0.0))
			Expect(t.Rotation.BasisY).To(dprectest.HaveVec2Coords(0.0, 1.0))
		})

		It("offsets points without rotating them", func() {
			t := shape2d.TranslationTransform(translation)
			Expect(t.Apply(dprec.NewVec2(3.0, 4.0))).To(dprectest.HaveVec2Coords(4.0, 6.0))
		})
	})

	Describe("RotationTransform", func() {
		It("stores the rotation and a zero translation", func() {
			t := shape2d.RotationTransform(rot90)
			Expect(t.Translation).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(t.Rotation.BasisX).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(t.Rotation.BasisY).To(dprectest.HaveVec2Coords(-1.0, 0.0))
		})

		It("rotates points without offsetting them", func() {
			t := shape2d.RotationTransform(rot90)
			Expect(t.Apply(dprec.NewVec2(3.0, 4.0))).To(dprectest.HaveVec2Coords(-4.0, 3.0))
		})
	})

	Describe("TRTransform", func() {
		It("stores both the translation and the rotation", func() {
			t := shape2d.TRTransform(translation, rot90)
			Expect(t.Translation).To(dprectest.HaveVec2Coords(1.0, 2.0))
			Expect(t.Rotation.BasisX).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(t.Rotation.BasisY).To(dprectest.HaveVec2Coords(-1.0, 0.0))
		})

		It("rotates points and then offsets them", func() {
			t := shape2d.TRTransform(translation, rot90)
			// rotate (3,4) -> (-4,3), then translate by (1,2) -> (-3,5)
			Expect(t.Apply(dprec.NewVec2(3.0, 4.0))).To(dprectest.HaveVec2Coords(-3.0, 5.0))
		})
	})

	Describe("ChainedTransform", func() {
		It("equals applying the child transform and then the parent", func() {
			parent := shape2d.TRTransform(dprec.NewVec2(1.0, 2.0), rot90)
			child := shape2d.TRTransform(dprec.NewVec2(-2.0, 4.0), rot90)
			chained := shape2d.ChainedTransform(parent, child)

			point := dprec.NewVec2(5.0, -3.0)
			expected := parent.Apply(child.Apply(point))
			Expect(chained.Apply(point)).To(dprectest.HaveVec2Coords(expected.X, expected.Y))
		})

		It("yields the original transform when chained with the identity", func() {
			t := shape2d.TRTransform(translation, rot90)
			identity := shape2d.IdentityTransform()

			leftChained := shape2d.ChainedTransform(identity, t)
			rightChained := shape2d.ChainedTransform(t, identity)

			point := dprec.NewVec2(5.0, -3.0)
			expected := t.Apply(point)
			Expect(leftChained.Apply(point)).To(dprectest.HaveVec2Coords(expected.X, expected.Y))
			Expect(rightChained.Apply(point)).To(dprectest.HaveVec2Coords(expected.X, expected.Y))
		})

		It("composes two 90-degree rotations into a 180-degree rotation", func() {
			rot := shape2d.RotationTransform(rot90)
			chained := shape2d.ChainedTransform(rot, rot)
			Expect(chained.Apply(dprec.NewVec2(1.0, 0.0))).To(dprectest.HaveVec2Coords(-1.0, 0.0))
		})
	})

	Describe("Inverse", func() {
		It("yields the identity for the identity transform", func() {
			inv := shape2d.IdentityTransform().Inverse()
			Expect(inv.Apply(dprec.NewVec2(3.0, 4.0))).To(dprectest.HaveVec2Coords(3.0, 4.0))
		})

		It("negates the translation of a translation-only transform", func() {
			inv := shape2d.TranslationTransform(translation).Inverse()
			Expect(inv.Translation).To(dprectest.HaveVec2Coords(-1.0, -2.0))
		})

		It("undoes Apply on a general point", func() {
			t := shape2d.TRTransform(translation, rot90)
			point := dprec.NewVec2(5.0, -3.0)
			Expect(t.Inverse().Apply(t.Apply(point))).To(dprectest.HaveVec2Coords(point.X, point.Y))
		})

		It("undoes the transform when applied in the other order", func() {
			t := shape2d.TRTransform(translation, rot90)
			point := dprec.NewVec2(5.0, -3.0)
			Expect(t.Apply(t.Inverse().Apply(point))).To(dprectest.HaveVec2Coords(point.X, point.Y))
		})

		It("chains with the original transform to form the identity", func() {
			t := shape2d.TRTransform(translation, rot90)
			chained := shape2d.ChainedTransform(t, t.Inverse())
			point := dprec.NewVec2(5.0, -3.0)
			Expect(chained.Apply(point)).To(dprectest.HaveVec2Coords(point.X, point.Y))
		})
	})

	Describe("Apply", func() {
		It("rotates and then translates a point", func() {
			t := shape2d.TRTransform(translation, rot90)
			Expect(t.Apply(dprec.NewVec2(3.0, 4.0))).To(dprectest.HaveVec2Coords(-3.0, 5.0))
		})

		It("returns the translation for the zero point", func() {
			t := shape2d.TRTransform(translation, rot90)
			Expect(t.Apply(dprec.ZeroVec2())).To(dprectest.HaveVec2Coords(1.0, 2.0))
		})
	})
})
