package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

const transformEpsilon = 1e-9

func expectVec3(actual, expected dprec.Vec3) {
	GinkgoHelper()
	Expect(actual.X).To(BeNumerically("~", expected.X, transformEpsilon))
	Expect(actual.Y).To(BeNumerically("~", expected.Y, transformEpsilon))
	Expect(actual.Z).To(BeNumerically("~", expected.Z, transformEpsilon))
}

// expectRotation checks that two quaternions represent the same rotation,
// regardless of the hemisphere they are in.
func expectRotation(actual, expected dprec.Quat) {
	GinkgoHelper()
	Expect(actual.Norm()).To(BeNumerically("~", 1.0, transformEpsilon))
	Expect(dprec.Abs(dprec.QuatDot(actual, expected))).To(BeNumerically("~", 1.0, transformEpsilon))
}

func rotationAround(degrees float64, axis dprec.Vec3) dprec.Quat {
	return dprec.RotationQuat(dprec.Degrees(degrees), axis)
}

func yRotation(degrees float64) dprec.Quat {
	return rotationAround(degrees, dprec.BasisYVec3())
}

var _ = Describe("IdentityTransform", func() {
	It("has a zero translation, an identity rotation and a unit scale", func() {
		result := animation3.IdentityTransform()
		expectVec3(result.Translation, dprec.ZeroVec3())
		expectRotation(result.Rotation, dprec.IdentityQuat())
		expectVec3(result.Scale, dprec.NewVec3(1.0, 1.0, 1.0))
	})

	It("leaves points unchanged", func() {
		point := dprec.NewVec3(1.0, -2.0, 3.0)
		result := dprec.Mat4Vec3Transformation(animation3.IdentityTransform().Matrix(), point)
		expectVec3(result, point)
	})

	It("leaves the other transform unchanged when added on either side", func() {
		transform := animation3.Transform{
			Translation: dprec.NewVec3(1.0, 2.0, 3.0),
			Rotation:    rotationAround(60.0, dprec.UnitVec3(dprec.NewVec3(1.0, 1.0, 0.0))),
			Scale:       dprec.NewVec3(2.0, 3.0, 4.0),
		}
		for _, result := range []animation3.Transform{
			animation3.AddTransforms(transform, animation3.IdentityTransform()),
			animation3.AddTransforms(animation3.IdentityTransform(), transform),
		} {
			expectVec3(result.Translation, transform.Translation)
			expectRotation(result.Rotation, transform.Rotation)
			expectVec3(result.Scale, transform.Scale)
		}
	})
})

var _ = Describe("Transform", func() {
	Describe("Matrix", func() {
		It("scales, then rotates and finally translates a point", func() {
			transform := animation3.Transform{
				Translation: dprec.NewVec3(1.0, 2.0, 3.0),
				Rotation:    yRotation(90.0),
				Scale:       dprec.NewVec3(2.0, 3.0, 4.0),
			}
			// The X axis is scaled to (2, 0, 0), rotated onto (0, 0, -2) and
			// then translated to (1, 2, 1).
			point := dprec.Mat4Vec3Transformation(transform.Matrix(), dprec.BasisXVec3())
			expectVec3(point, dprec.NewVec3(1.0, 2.0, 1.0))
		})
	})
})

var _ = Describe("BlendTransforms", func() {
	var first, second animation3.Transform

	BeforeEach(func() {
		first = animation3.Transform{
			Translation: dprec.NewVec3(1.0, 2.0, 3.0),
			Rotation:    yRotation(0.0),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
		second = animation3.Transform{
			Translation: dprec.NewVec3(5.0, -2.0, 3.0),
			Rotation:    yRotation(90.0),
			Scale:       dprec.NewVec3(3.0, 2.0, 0.5),
		}
	})

	It("returns the first transform for a factor of 0.0", func() {
		result := animation3.BlendTransforms(first, second, 0.0)
		expectVec3(result.Translation, first.Translation)
		expectRotation(result.Rotation, first.Rotation)
		expectVec3(result.Scale, first.Scale)
	})

	It("returns the second transform for a factor of 1.0", func() {
		result := animation3.BlendTransforms(first, second, 1.0)
		expectVec3(result.Translation, second.Translation)
		expectRotation(result.Rotation, second.Rotation)
		expectVec3(result.Scale, second.Scale)
	})

	It("blends translation and scale linearly", func() {
		result := animation3.BlendTransforms(first, second, 0.25)
		expectVec3(result.Translation, dprec.NewVec3(2.0, 1.0, 3.0))
		expectVec3(result.Scale, dprec.NewVec3(1.5, 1.25, 0.875))
	})

	It("blends rotation with a constant angular velocity", func() {
		result := animation3.BlendTransforms(first, second, 0.25)
		expectRotation(result.Rotation, yRotation(22.5))

		result = animation3.BlendTransforms(first, second, 0.5)
		expectRotation(result.Rotation, yRotation(45.0))
	})

	It("extrapolates for a factor outside the range", func() {
		result := animation3.BlendTransforms(first, second, 2.0)
		expectVec3(result.Translation, dprec.NewVec3(9.0, -6.0, 3.0))
		expectRotation(result.Rotation, yRotation(180.0))
		expectVec3(result.Scale, dprec.NewVec3(5.0, 3.0, 0.0))
	})

	It("blends rotations along the shorter arc", func() {
		result := animation3.BlendTransforms(
			animation3.Transform{Rotation: yRotation(0.0)},
			animation3.Transform{Rotation: yRotation(270.0)},
			0.5,
		)
		expectRotation(result.Rotation, yRotation(-45.0))
	})

	It("blends rotations along the shorter arc when quaternions are in opposite hemispheres", func() {
		result := animation3.BlendTransforms(
			animation3.Transform{Rotation: yRotation(0.0)},
			animation3.Transform{Rotation: dprec.NegativeQuat(yRotation(90.0))},
			0.5,
		)
		expectRotation(result.Rotation, yRotation(45.0))
	})

	It("returns the same rotation when both rotations are equal", func() {
		rotation := rotationAround(60.0, dprec.UnitVec3(dprec.NewVec3(1.0, 1.0, 0.0)))
		result := animation3.BlendTransforms(
			animation3.Transform{Rotation: rotation},
			animation3.Transform{Rotation: rotation},
			0.3,
		)
		expectRotation(result.Rotation, rotation)
	})
})

var _ = Describe("AddTransforms", func() {
	var first, second animation3.Transform

	BeforeEach(func() {
		first = animation3.Transform{
			Translation: dprec.NewVec3(1.0, 2.0, 3.0),
			Rotation:    yRotation(30.0),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
		second = animation3.Transform{
			Translation: dprec.NewVec3(5.0, -2.0, 3.0),
			Rotation:    yRotation(60.0),
			Scale:       dprec.NewVec3(0.5, 1.0, 2.0),
		}
	})

	It("adds translation and multiplies scale per axis", func() {
		result := animation3.AddTransforms(first, second)
		expectVec3(result.Translation, dprec.NewVec3(6.0, 0.0, 6.0))
		expectVec3(result.Scale, dprec.NewVec3(0.5, 1.0, 2.0))
	})

	It("composes rotation about a shared axis", func() {
		result := animation3.AddTransforms(first, second)
		expectRotation(result.Rotation, yRotation(90.0))
	})

	It("applies the second rotation on top of the first", func() {
		// The first rotation maps the X axis onto -Z; the second then maps
		// -Z onto +Y. Applying the arguments in the opposite order would map
		// the X axis onto -Z instead, so this pins the composition order.
		result := animation3.AddTransforms(
			animation3.Transform{Rotation: yRotation(90.0)},
			animation3.Transform{Rotation: rotationAround(90.0, dprec.BasisXVec3())},
		)
		rotated := dprec.QuatVec3Rotation(result.Rotation, dprec.BasisXVec3())
		expectVec3(rotated, dprec.NewVec3(0.0, 1.0, 0.0))
	})

	It("leaves the other transform unchanged when adding the identity transform", func() {
		transform := animation3.Transform{
			Translation: dprec.NewVec3(1.0, 2.0, 3.0),
			Rotation:    yRotation(45.0),
			Scale:       dprec.NewVec3(2.0, 3.0, 4.0),
		}
		result := animation3.AddTransforms(transform, animation3.IdentityTransform())
		expectVec3(result.Translation, transform.Translation)
		expectRotation(result.Rotation, transform.Rotation)
		expectVec3(result.Scale, transform.Scale)
	})
})

var _ = Describe("DiffTransforms", func() {
	var first, second animation3.Transform

	BeforeEach(func() {
		first = animation3.Transform{
			Translation: dprec.NewVec3(6.0, 0.0, 6.0),
			Rotation:    yRotation(90.0),
			Scale:       dprec.NewVec3(0.5, 1.0, 2.0),
		}
		second = animation3.Transform{
			Translation: dprec.NewVec3(1.0, 2.0, 3.0),
			Rotation:    yRotation(30.0),
			Scale:       dprec.NewVec3(2.0, 4.0, 0.5),
		}
	})

	It("subtracts translation and divides scale per axis", func() {
		result := animation3.DiffTransforms(first, second)
		expectVec3(result.Translation, dprec.NewVec3(5.0, -2.0, 3.0))
		expectVec3(result.Scale, dprec.NewVec3(0.25, 0.25, 4.0))
	})

	It("subtracts rotation about a shared axis", func() {
		result := animation3.DiffTransforms(first, second)
		expectRotation(result.Rotation, yRotation(60.0))
	})

	It("is reversed by adding the result on top of the second transform", func() {
		// The rotations are about different axes, so this pins the
		// composition order, which a shared axis would not.
		pose := animation3.Transform{
			Translation: dprec.NewVec3(1.0, 2.0, 3.0),
			Rotation:    rotationAround(90.0, dprec.BasisXVec3()),
			Scale:       dprec.NewVec3(2.0, 3.0, 4.0),
		}
		reference := animation3.Transform{
			Translation: dprec.NewVec3(-1.0, 0.5, 2.0),
			Rotation:    yRotation(90.0),
			Scale:       dprec.NewVec3(0.5, 2.0, 1.0),
		}
		diff := animation3.DiffTransforms(pose, reference)
		result := animation3.AddTransforms(reference, diff)
		expectVec3(result.Translation, pose.Translation)
		expectRotation(result.Rotation, pose.Rotation)
		expectVec3(result.Scale, pose.Scale)
	})

	It("yields the identity transform when diffing a transform against itself", func() {
		transform := animation3.Transform{
			Translation: dprec.NewVec3(1.0, 2.0, 3.0),
			Rotation:    rotationAround(60.0, dprec.UnitVec3(dprec.NewVec3(1.0, 1.0, 0.0))),
			Scale:       dprec.NewVec3(2.0, 3.0, 4.0),
		}
		result := animation3.DiffTransforms(transform, transform)
		identity := animation3.IdentityTransform()
		expectVec3(result.Translation, identity.Translation)
		expectRotation(result.Rotation, identity.Rotation)
		expectVec3(result.Scale, identity.Scale)
	})
})
