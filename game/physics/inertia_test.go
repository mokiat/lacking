package physics_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/game/physics"
)

var _ = Describe("Inertia", func() {

	Describe("DiagonalMomentOfInertia", func() {
		It("places the moments on the diagonal", func() {
			tensor := physics.DiagonalMomentOfInertia(1.0, 2.0, 3.0)
			Expect(tensor).To(dprectest.HaveMat3Elements(
				1.0, 0.0, 0.0,
				0.0, 2.0, 0.0,
				0.0, 0.0, 3.0,
			))
		})
	})

	Describe("SymmetricMomentOfInertia", func() {
		It("uses the same moment for all axes", func() {
			tensor := physics.SymmetricMomentOfInertia(5.0)
			Expect(tensor).To(dprectest.HaveMat3Elements(
				5.0, 0.0, 0.0,
				0.0, 5.0, 0.0,
				0.0, 0.0, 5.0,
			))
		})
	})

	Describe("SolidSphereMomentOfInertia", func() {
		It("matches the analytical solution", func() {
			// I = (2/5) * m * r^2 = 0.4 * 3.0 * 4.0
			tensor := physics.SolidSphereMomentOfInertia(3.0, 2.0)
			Expect(tensor).To(dprectest.HaveMat3Elements(
				4.8, 0.0, 0.0,
				0.0, 4.8, 0.0,
				0.0, 0.0, 4.8,
			))
		})
	})

	Describe("HollowSphereMomentOfInertia", func() {
		It("matches the analytical solution", func() {
			// I = (2/3) * m * r^2 = (2/3) * 3.0 * 4.0
			tensor := physics.HollowSphereMomentOfInertia(3.0, 2.0)
			Expect(tensor).To(dprectest.HaveMat3Elements(
				8.0, 0.0, 0.0,
				0.0, 8.0, 0.0,
				0.0, 0.0, 8.0,
			))
		})

		It("exceeds the moment of a solid sphere of equal mass and radius", func() {
			hollow := physics.HollowSphereMomentOfInertia(3.0, 2.0)
			solid := physics.SolidSphereMomentOfInertia(3.0, 2.0)
			Expect(hollow.M11).To(BeNumerically(">", solid.M11))
		})
	})

	Describe("SolidBoxMomentOfInertia", func() {
		It("matches the analytical solution", func() {
			// Ixx = (m/12) * (h^2 + l^2) = (12/12) * (4 + 9)
			tensor := physics.SolidBoxMomentOfInertia(12.0, 1.0, 2.0, 3.0)
			Expect(tensor).To(dprectest.HaveMat3Elements(
				13.0, 0.0, 0.0,
				0.0, 10.0, 0.0,
				0.0, 0.0, 5.0,
			))
		})

		It("is symmetric for a cube", func() {
			// I = (1/6) * m * s^2 = (1/6) * 3.0 * 4.0
			tensor := physics.SolidBoxMomentOfInertia(3.0, 2.0, 2.0, 2.0)
			Expect(tensor).To(dprectest.HaveMat3Elements(
				2.0, 0.0, 0.0,
				0.0, 2.0, 0.0,
				0.0, 0.0, 2.0,
			))
		})
	})

	Describe("HollowBoxMomentOfInertia", func() {
		It("matches the analytical solution for a cube", func() {
			// I = (5/18) * m * s^2 = (5/18) * 3.0 * 4.0
			tensor := physics.HollowBoxMomentOfInertia(3.0, 2.0, 2.0, 2.0)
			Expect(tensor).To(dprectest.HaveMat3Elements(
				10.0/3.0, 0.0, 0.0,
				0.0, 10.0/3.0, 0.0,
				0.0, 0.0, 10.0/3.0,
			))
		})

		It("orders the moments according to the dimensions", func() {
			// The Z axis is the longest one, so it has the smallest moment,
			// whereas the shortest X axis has the largest one.
			tensor := physics.HollowBoxMomentOfInertia(1.0, 1.0, 2.0, 3.0)
			Expect(tensor.M11).To(BeNumerically(">", tensor.M22))
			Expect(tensor.M22).To(BeNumerically(">", tensor.M33))
		})

		It("exceeds the moment of a solid box of equal mass and dimensions", func() {
			hollow := physics.HollowBoxMomentOfInertia(3.0, 1.0, 2.0, 3.0)
			solid := physics.SolidBoxMomentOfInertia(3.0, 1.0, 2.0, 3.0)
			Expect(hollow.M11).To(BeNumerically(">", solid.M11))
			Expect(hollow.M22).To(BeNumerically(">", solid.M22))
			Expect(hollow.M33).To(BeNumerically(">", solid.M33))
		})
	})

	Describe("RotatedMomentOfInertia", func() {
		It("leaves the tensor unchanged for an identity rotation", func() {
			tensor := physics.DiagonalMomentOfInertia(1.0, 2.0, 3.0)
			Expect(physics.RotatedMomentOfInertia(tensor, dprec.IdentityQuat())).To(dprectest.HaveMat3Elements(
				1.0, 0.0, 0.0,
				0.0, 2.0, 0.0,
				0.0, 0.0, 3.0,
			))
		})

		It("swaps the moments when rotating 90 degrees around Z", func() {
			// The X axis takes the place of Y and vice versa, so their
			// moments are exchanged, whereas Z is left alone.
			tensor := physics.DiagonalMomentOfInertia(1.0, 2.0, 3.0)
			rotation := dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3())
			Expect(physics.RotatedMomentOfInertia(tensor, rotation)).To(dprectest.HaveMat3Elements(
				2.0, 0.0, 0.0,
				0.0, 1.0, 0.0,
				0.0, 0.0, 3.0,
			))
		})

		It("leaves a symmetric tensor unchanged for any rotation", func() {
			tensor := physics.SymmetricMomentOfInertia(5.0)
			rotation := dprec.RotationQuat(dprec.Degrees(37.0), dprec.UnitVec3(dprec.NewVec3(1.0, 2.0, 3.0)))
			Expect(physics.RotatedMomentOfInertia(tensor, rotation)).To(dprectest.HaveMat3Elements(
				5.0, 0.0, 0.0,
				0.0, 5.0, 0.0,
				0.0, 0.0, 5.0,
			))
		})

		It("preserves the trace, which is rotation invariant", func() {
			tensor := physics.DiagonalMomentOfInertia(1.0, 2.0, 3.0)
			rotation := dprec.RotationQuat(dprec.Degrees(37.0), dprec.UnitVec3(dprec.NewVec3(1.0, 2.0, 3.0)))
			result := physics.RotatedMomentOfInertia(tensor, rotation)
			Expect(result.M11 + result.M22 + result.M33).To(BeNumerically("~", 6.0, 1e-9))
		})

		It("produces a symmetric matrix for an arbitrary rotation", func() {
			tensor := physics.DiagonalMomentOfInertia(1.0, 2.0, 3.0)
			rotation := dprec.RotationQuat(dprec.Degrees(37.0), dprec.UnitVec3(dprec.NewVec3(1.0, 2.0, 3.0)))
			result := physics.RotatedMomentOfInertia(tensor, rotation)
			Expect(result.M12).To(BeNumerically("~", result.M21, 1e-9))
			Expect(result.M13).To(BeNumerically("~", result.M31, 1e-9))
			Expect(result.M23).To(BeNumerically("~", result.M32, 1e-9))
		})
	})

	Describe("OffsetMomentOfInertia", func() {
		It("leaves the tensor unchanged for a zero offset", func() {
			tensor := physics.DiagonalMomentOfInertia(1.0, 2.0, 3.0)
			Expect(physics.OffsetMomentOfInertia(tensor, 5.0, dprec.ZeroVec3())).To(dprectest.HaveMat3Elements(
				1.0, 0.0, 0.0,
				0.0, 2.0, 0.0,
				0.0, 0.0, 3.0,
			))
		})

		It("matches the parallel axis theorem for an axis aligned offset", func() {
			// An offset of 2 along X adds mass * 4 to the Y and Z moments,
			// but leaves the X moment untouched.
			tensor := physics.DiagonalMomentOfInertia(1.0, 2.0, 3.0)
			offset := dprec.NewVec3(2.0, 0.0, 0.0)
			Expect(physics.OffsetMomentOfInertia(tensor, 3.0, offset)).To(dprectest.HaveMat3Elements(
				1.0, 0.0, 0.0,
				0.0, 14.0, 0.0,
				0.0, 0.0, 15.0,
			))
		})

		It("introduces products of inertia for a diagonal offset", func() {
			// The off-diagonal term is -mass * x * y = -2 * 3 * 4.
			tensor := dprec.ZeroMat3()
			offset := dprec.NewVec3(3.0, 4.0, 0.0)
			Expect(physics.OffsetMomentOfInertia(tensor, 2.0, offset)).To(dprectest.HaveMat3Elements(
				32.0, -24.0, 0.0,
				-24.0, 18.0, 0.0,
				0.0, 0.0, 50.0,
			))
		})

		It("depends only on the distance for a point mass", func() {
			// A point mass has no moment of its own, so the result is
			// mass * distance^2 around the axes perpendicular to the offset.
			alongX := physics.OffsetMomentOfInertia(dprec.ZeroMat3(), 2.0, dprec.NewVec3(5.0, 0.0, 0.0))
			alongY := physics.OffsetMomentOfInertia(dprec.ZeroMat3(), 2.0, dprec.NewVec3(0.0, 5.0, 0.0))
			Expect(alongX.M22).To(BeNumerically("~", 50.0, 1e-9))
			Expect(alongY.M11).To(BeNumerically("~", 50.0, 1e-9))
		})
	})

	Describe("composing a body out of parts", func() {
		It("matches the analytical solution for a split solid box", func() {
			// A 2x2x2 box of mass 4 is modelled as two 1x2x2 halves of mass
			// 2, each offset by 0.5 along X from the center of the whole.
			half := physics.SolidBoxMomentOfInertia(2.0, 1.0, 2.0, 2.0)
			left := physics.OffsetMomentOfInertia(half, 2.0, dprec.NewVec3(-0.5, 0.0, 0.0))
			right := physics.OffsetMomentOfInertia(half, 2.0, dprec.NewVec3(0.5, 0.0, 0.0))

			whole := physics.SolidBoxMomentOfInertia(4.0, 2.0, 2.0, 2.0)
			Expect(physics.MomentOfInertiaSum(left, right)).To(dprectest.HaveMat3Elements(
				whole.M11, whole.M12, whole.M13,
				whole.M21, whole.M22, whole.M23,
				whole.M31, whole.M32, whole.M33,
			))
		})

		It("is unaffected by rotating a part around its own axis of symmetry", func() {
			// A part rotated 90 degrees around Z and placed along X gives the
			// same result as reflecting it to the other side.
			part := physics.SolidBoxMomentOfInertia(2.0, 1.0, 2.0, 3.0)
			rotation := dprec.RotationQuat(dprec.Degrees(180.0), dprec.BasisZVec3())

			plain := physics.OffsetMomentOfInertia(part, 2.0, dprec.NewVec3(4.0, 0.0, 0.0))
			rotated := physics.OffsetMomentOfInertia(
				physics.RotatedMomentOfInertia(part, rotation),
				2.0,
				dprec.NewVec3(4.0, 0.0, 0.0),
			)
			Expect(rotated).To(dprectest.HaveMat3Elements(
				plain.M11, plain.M12, plain.M13,
				plain.M21, plain.M22, plain.M23,
				plain.M31, plain.M32, plain.M33,
			))
		})

		It("cancels the products of inertia for a symmetric pair of wings", func() {
			// Two identical wings mirrored across the fuselage produce
			// opposite products of inertia, leaving a diagonal total.
			wing := physics.SolidBoxMomentOfInertia(1.0, 4.0, 0.2, 1.0)
			leftWing := physics.OffsetMomentOfInertia(wing, 1.0, dprec.NewVec3(-3.0, 0.5, 0.0))
			rightWing := physics.OffsetMomentOfInertia(wing, 1.0, dprec.NewVec3(3.0, 0.5, 0.0))
			fuselage := physics.SolidBoxMomentOfInertia(5.0, 0.8, 0.8, 6.0)

			total := physics.MomentOfInertiaMultiSum(fuselage, leftWing, rightWing)
			Expect(total.M12).To(BeNumerically("~", 0.0, 1e-9))
			Expect(total.M21).To(BeNumerically("~", 0.0, 1e-9))
			Expect(total.M13).To(BeNumerically("~", 0.0, 1e-9))
			Expect(total.M23).To(BeNumerically("~", 0.0, 1e-9))
		})
	})

	Describe("MomentOfInertiaSum", func() {
		It("adds the tensors element by element", func() {
			first := dprec.NewMat3(
				1.0, 2.0, 3.0,
				4.0, 5.0, 6.0,
				7.0, 8.0, 9.0,
			)
			second := dprec.NewMat3(
				10.0, 20.0, 30.0,
				40.0, 50.0, 60.0,
				70.0, 80.0, 90.0,
			)
			Expect(physics.MomentOfInertiaSum(first, second)).To(dprectest.HaveMat3Elements(
				11.0, 22.0, 33.0,
				44.0, 55.0, 66.0,
				77.0, 88.0, 99.0,
			))
		})

		It("treats the zero tensor as a neutral element", func() {
			tensor := physics.DiagonalMomentOfInertia(1.0, 2.0, 3.0)
			Expect(physics.MomentOfInertiaSum(tensor, dprec.ZeroMat3())).To(dprectest.HaveMat3Elements(
				1.0, 0.0, 0.0,
				0.0, 2.0, 0.0,
				0.0, 0.0, 3.0,
			))
		})

		It("matches a single body when splitting its mass in two", func() {
			// Inertia is linear in mass, so two half-mass spheres that share
			// a center amount to a single full-mass sphere.
			half := physics.SolidSphereMomentOfInertia(1.5, 2.0)
			whole := physics.SolidSphereMomentOfInertia(3.0, 2.0)
			Expect(physics.MomentOfInertiaSum(half, half)).To(dprectest.HaveMat3Elements(
				whole.M11, whole.M12, whole.M13,
				whole.M21, whole.M22, whole.M23,
				whole.M31, whole.M32, whole.M33,
			))
		})
	})

	Describe("MomentOfInertiaMultiSum", func() {
		It("returns the tensor itself when there are no other tensors", func() {
			tensor := physics.DiagonalMomentOfInertia(1.0, 2.0, 3.0)
			Expect(physics.MomentOfInertiaMultiSum(tensor)).To(dprectest.HaveMat3Elements(
				1.0, 0.0, 0.0,
				0.0, 2.0, 0.0,
				0.0, 0.0, 3.0,
			))
		})

		It("adds all of the specified tensors", func() {
			first := physics.DiagonalMomentOfInertia(1.0, 2.0, 3.0)
			second := physics.SymmetricMomentOfInertia(10.0)
			third := physics.DiagonalMomentOfInertia(0.5, 0.5, 0.5)
			Expect(physics.MomentOfInertiaMultiSum(first, second, third)).To(dprectest.HaveMat3Elements(
				11.5, 0.0, 0.0,
				0.0, 12.5, 0.0,
				0.0, 0.0, 13.5,
			))
		})

		It("is equivalent to repeated use of MomentOfInertiaSum", func() {
			first := physics.DiagonalMomentOfInertia(1.0, 2.0, 3.0)
			second := physics.SymmetricMomentOfInertia(10.0)
			third := physics.DiagonalMomentOfInertia(0.5, 0.5, 0.5)
			expected := physics.MomentOfInertiaSum(physics.MomentOfInertiaSum(first, second), third)
			Expect(physics.MomentOfInertiaMultiSum(first, second, third)).To(dprectest.HaveMat3Elements(
				expected.M11, expected.M12, expected.M13,
				expected.M21, expected.M22, expected.M23,
				expected.M31, expected.M32, expected.M33,
			))
		})
	})

})
